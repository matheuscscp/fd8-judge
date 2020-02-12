package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/matheuscscp/fd8-judge/judge"
	"github.com/matheuscscp/fd8-judge/pkg/services/file"
	"github.com/matheuscscp/fd8-judge/pkg/services/program"
	"github.com/spf13/cobra"
)

func init() {
	defineJudgeCommand()
}

type (
	// executeFlags holds the flags for the judge execute command.
	executeFlags struct {
		bundleRequestURL          string
		bundleRequestHeaders      string
		solutionRequestURL        string
		solutionRequestHeaders    string
		interactor                string
		uploadAuthorizedServerURL string
		interactorProgramService  string
		solutionProgramService    string
		interactorCage            cageFlags
		solutionCage              cageFlags
	}
)

const (
	executeCmdInteractorCageFlagPrefix = "interactor-cage-"
	executeCmdSolutionCageFlagPrefix   = "solution-cage-"
)

// defineJudgeCommand defines the judge command and its subcommands.
func defineJudgeCommand() {
	judgeCmd := &cobra.Command{
		Use:   "judge",
		Short: "Execute or check a problem solution.",
		Long:  "Automatic judge to execute and check problem solutions.",
	}
	rootCmd.AddCommand(judgeCmd)

	executeFlags := &executeFlags{}
	executeCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a problem solution.",
		Long:  "Execute a problem solution and store the outputs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			executor, err := parseExecuteFlags(cmd, executeFlags)
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return executor.Execute()
		},
	}
	bindExecuteFlags(executeCmd, executeFlags)
	judgeCmd.AddCommand(executeCmd)
}

// parseExecuteFlags parses the judge execute command flags.
func parseExecuteFlags(cmd *cobra.Command, flags *executeFlags) (*judge.Executor, error) {
	bundleHeaders := make(http.Header)
	if err := json.Unmarshal([]byte(flags.bundleRequestHeaders), &bundleHeaders); err != nil {
		return nil, fmt.Errorf("error unmarshaling problem bundle request headers: %w", err)
	}
	solutionHeaders := make(http.Header)
	if err := json.Unmarshal([]byte(flags.solutionRequestHeaders), &solutionHeaders); err != nil {
		return nil, fmt.Errorf("error unmarshaling problem solution request headers: %w", err)
	}
	interactorProgramService, err := program.NewService(flags.interactorProgramService, nil)
	if flags.interactorProgramService != "" && err != nil {
		return nil, fmt.Errorf("error creating program service for interactor: %w", err)
	}
	solutionProgramService, err := program.NewService(flags.solutionProgramService, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating program service for solution: %w", err)
	}
	switch flags.interactor {
	case judge.NoInteractor:
		if interactorProgramService != nil {
			return nil, fmt.Errorf("interactor program service created for NoInteractor request")
		}
	case judge.DefaultInteractor:
		if interactorProgramService != nil {
			return nil, fmt.Errorf("interactor program service created for DefaultInteractor request")
		}
	case judge.CustomInteractor:
		if interactorProgramService == nil {
			return nil, fmt.Errorf("interactor program service missing for CustomInteractor request")
		}
	default:
		return nil, fmt.Errorf("invalid interactor")
	}
	interactorCage, err := parseCageFlags(cmd, &flags.interactorCage, executeCmdInteractorCageFlagPrefix)
	if err != nil {
		return nil, fmt.Errorf("error parsing interactor cage flags: %w", err)
	}
	solutionCage, err := parseCageFlags(cmd, &flags.solutionCage, executeCmdSolutionCageFlagPrefix)
	if err != nil {
		return nil, fmt.Errorf("error parsing solution cage flags: %w", err)
	}
	return &judge.Executor{
		BundleRequestURL:          flags.bundleRequestURL,
		BundleRequestHeaders:      bundleHeaders,
		SolutionRequestURL:        flags.solutionRequestURL,
		SolutionRequestHeaders:    solutionHeaders,
		Interactor:                flags.interactor,
		UploadAuthorizedServerURL: flags.uploadAuthorizedServerURL,
		FileService:               file.NewService(nil),
		InteractorProgramService:  interactorProgramService,
		SolutionProgramService:    solutionProgramService,
		InteractorCage:            interactorCage,
		SolutionCage:              solutionCage,
	}, nil
}

// bindExecuteFlags binds judge execute command flags.
func bindExecuteFlags(cmd *cobra.Command, flags *executeFlags) {
	availableProgramServices := strings.Join(program.GetServices(), ", ")
	cmd.Flags().StringVar(
		&flags.bundleRequestURL, "bundle-request-url", "",
		"HTTP GET endpoint to download the problem bundle.",
	)
	cmd.Flags().StringVar(
		&flags.bundleRequestHeaders, "bundle-request-headers", "{}",
		"HTTP headers to send in the download request for the problem bundle (JSON text to map[string][]string).",
	)
	cmd.Flags().StringVar(
		&flags.solutionRequestURL, "solution-request-url", "",
		"HTTP GET endpoint to download the problem solution.",
	)
	cmd.Flags().StringVar(
		&flags.solutionRequestHeaders, "solution-request-headers", "{}",
		"HTTP headers to send in the download request for the problem solution (JSON text to map[string][]string).",
	)
	cmd.Flags().StringVar(
		&flags.interactor, "interactor", "",
		"Empty means no interactor, 'default-interactor' tells the judge to use the default interactor and "+
			"'custom-interactor' tells the judge to use a custom interactor supposed to be inside the problem bundle.",
	)
	cmd.Flags().StringVar(
		&flags.uploadAuthorizedServerURL, "upload-authorized-server-url", "",
		"HTTP endpoint to GET an one-time authorized upload request.",
	)
	cmd.Flags().StringVar(
		&flags.interactorProgramService, "interactor-program-service", "",
		fmt.Sprintf("Program service for interactor. Only works if --interactor=custom-interactor. (one in {%s})", availableProgramServices),
	)
	cmd.Flags().StringVar(
		&flags.solutionProgramService, "solution-program-service", "",
		fmt.Sprintf("Program service for solution. (one in {%s})", availableProgramServices),
	)
	bindCageFlags(cmd, &flags.interactorCage, executeCmdInteractorCageFlagPrefix, false /* bindExecFlags */)
	bindCageFlags(cmd, &flags.solutionCage, executeCmdSolutionCageFlagPrefix, false /* bindExecFlags */)
}
