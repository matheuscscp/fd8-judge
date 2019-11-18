package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func init() {
	defineMonsterCommand()
}

// monsterFlags holds the flags for the monster command.
type monsterFlags struct {
	loopForever bool
}

// defineMonsterCommand defines the monster command.
func defineMonsterCommand() {
	flags := &monsterFlags{}
	cmd := &cobra.Command{
		Use:   "monster",
		Short: "Program used to test the cage.",
		Long:  "Program with options to set various bad behaviors to test the cage.",
		Run: func(_ *cobra.Command, _ []string) {
			runMonster(flags)
		},
	}
	bindMonsterFlags(cmd, flags)
	rootCmd.AddCommand(cmd)
}

// runMonster actually runs the monster command.
func runMonster(flags *monsterFlags) {
	sigxcpuChannel := make(chan os.Signal, 1)
	signal.Notify(sigxcpuChannel, syscall.SIGXCPU)
	go func() {
		<-sigxcpuChannel
		os.Exit(int(syscall.SIGXCPU))
	}()

	if flags.loopForever {
		for {
			flags.loopForever = !flags.loopForever
		}
	}

	fmt.Println("hello, world!")
}

// bindMonsterFlags binds monster command flags.
func bindMonsterFlags(cmd *cobra.Command, flags *monsterFlags) {
	cmd.Flags().BoolVar(
		&flags.loopForever, "loop-forever", false,
		"Set to true if the program should loop forever.",
	)
}
