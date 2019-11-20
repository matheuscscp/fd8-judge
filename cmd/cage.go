package cmd

import (
	"fmt"
	"time"

	"github.com/matheuscscp/fd8-judge/pkg/cage"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	defineCageCommand()
}

// cageFlags holds the flags for the cage command.
type cageFlags struct {
	timeLimit time.Duration
	execPath  string
	execArgs  []string
}

// defineCageCommand defines the cage command.
func defineCageCommand() {
	cageFlags := &cageFlags{}
	cageCmd := &cobra.Command{
		Use:   cage.CommandLineCommand,
		Short: "Execute a process safely.",
		Long:  "Execute a process in a minimally safe environment.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cage, err := parseCageFlags(cmd, cageFlags, "" /* flagPrefix */)
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return cage.Execute()
		},
	}
	bindCageFlags(
		cageCmd,
		cageFlags,
		"",   // flagPrefix
		true, // bindExecFlags
	)
	rootCmd.AddCommand(cageCmd)
}

// parseCageFlags parses the cage command flags.
func parseCageFlags(cmd *cobra.Command, flags *cageFlags, flagPrefix string) (cage.Cage, error) {
	if flags.execPath == "" {
		return nil, fmt.Errorf("exec path not specified")
	}
	defaultCage := &cage.DefaultCage{
		ExecPath: flags.execPath,
		ExecArgs: flags.execArgs,
	}
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case flagPrefix + cage.TimeLimitFlag:
			defaultCage.TimeLimit = &flags.timeLimit
		}
	})
	return cage.New(defaultCage, nil), nil
}

// bindCageFlags binds cage command flags.
func bindCageFlags(cmd *cobra.Command, flags *cageFlags, flagPrefix string, bindExecFlags bool) {
	cmd.Flags().DurationVar(
		&flags.timeLimit, flagPrefix+cage.TimeLimitFlag, time.Duration(0),
		"Maximum time duration for which the program can stay running. Follows go's time package syntax.",
	)
	if bindExecFlags {
		cmd.Flags().StringVar(
			&flags.execPath, cage.ExecPathFlag, "",
			"Path to the file to be safely executed.",
		)
		cmd.Flags().StringArrayVar(
			&flags.execArgs, cage.ExecArgsFlag, []string{},
			"Arguments to be passed to the process.",
		)
	}
}
