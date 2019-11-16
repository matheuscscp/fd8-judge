package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/matheuscscp/fd8-judge/judge"
	"github.com/matheuscscp/fd8-judge/pkg/services"
	"github.com/spf13/cobra"
)

type (
	executeFlags struct {
		bundleRequestURL          string
		bundleRequestHeaders      string
		solutionRequestURL        string
		solutionRequestHeaders    string
		interactor                string
		uploadAuthorizedServerURL string
		interactorProgramService  string
		solutionProgramService    string
	}
)

func init() {
	judgeCmd := &cobra.Command{
		Use:   "judge",
		Short: "Automatic judge to execute and check problem solutions.",
		Long:  "Execute or check a problem solution.",
	}
	rootCmd.AddCommand(judgeCmd)

	executeFlags := &executeFlags{}
	executeCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a problem solution.",
		Long:  "Execute a problem solution and store the outputs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			executor, err := parseExecuteFlags(executeFlags)
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

// parseExecuteFlags parses the execute command flags.
func parseExecuteFlags(flags *executeFlags) (*judge.Executor, error) {
	bundleHeaders := make(http.Header)
	if err := json.Unmarshal([]byte(flags.bundleRequestHeaders), &bundleHeaders); err != nil {
		return nil, fmt.Errorf("error unmarshaling problem bundle request headers: %w", err)
	}
	solutionHeaders := make(http.Header)
	if err := json.Unmarshal([]byte(flags.solutionRequestHeaders), &solutionHeaders); err != nil {
		return nil, fmt.Errorf("error unmarshaling problem solution request headers: %w", err)
	}
	interactorProgramService, err := services.NewProgramService(flags.interactorProgramService, nil)
	if flags.interactorProgramService != "" && err != nil {
		return nil, fmt.Errorf("error creating program service for interactor: %w", err)
	}
	solutionProgramService, err := services.NewProgramService(flags.solutionProgramService, nil)
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
	return &judge.Executor{
		BundleRequestURL:          flags.bundleRequestURL,
		BundleRequestHeaders:      bundleHeaders,
		SolutionRequestURL:        flags.solutionRequestURL,
		SolutionRequestHeaders:    solutionHeaders,
		Interactor:                flags.interactor,
		UploadAuthorizedServerURL: flags.uploadAuthorizedServerURL,
		Runtime:                   &judge.ExecutorDefaultRuntime{},
		FileService:               services.NewFileService(nil),
		InteractorProgramService:  interactorProgramService,
		SolutionProgramService:    solutionProgramService,
	}, nil
}

// bindExecuteFlags binds execute command flags.
func bindExecuteFlags(cmd *cobra.Command, flags *executeFlags) {
	availableProgramServices := strings.Join(services.GetProgramServices(), ", ")
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
		"HTTP GET endpoint expecting an X-Content-Length header parameter to return an authorized upload request.",
	)
	cmd.Flags().StringVar(
		&flags.interactorProgramService, "interactor-program-service", "",
		fmt.Sprintf("Program service for interactor. Only works if --interactor=custom-interactor. (one in {%s})", availableProgramServices),
	)
	cmd.Flags().StringVar(
		&flags.solutionProgramService, "solution-program-service", "",
		fmt.Sprintf("Program service for solution. (one in {%s})", availableProgramServices),
	)
}
