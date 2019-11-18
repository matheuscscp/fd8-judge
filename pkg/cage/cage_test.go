// +build integration

package cage_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/matheuscscp/fd8-judge/pkg/cage"
	"github.com/stretchr/testify/assert"
)

func TestEncage(t *testing.T) {
	second := time.Second
	halfSecond := 500 * time.Millisecond

	var tests = map[string]struct {
		cagedCommandFlags []string
		cage              *cage.DefaultCage
		cageFlags         []string
		output            string
		returnsError      bool
		exitCodes         []int
	}{
		"hello-world": {
			cage:      &cage.DefaultCage{},
			output:    "hello, world!\n",
			exitCodes: []int{0},
		},
		"time-limit-exceeded": {
			cagedCommandFlags: []string{
				"--loop-forever",
			},
			cage: &cage.DefaultCage{
				TimeLimit: &second,
			},
			cageFlags: []string{
				"--time-limit",
				"1s",
			},
			returnsError: true,
			exitCodes:    []int{int(syscall.SIGXCPU), int(syscall.SIGKILL)},
		},
		"time-limit-defaults-to-one-second": {
			cagedCommandFlags: []string{
				"--loop-forever",
			},
			cage: &cage.DefaultCage{
				TimeLimit: &halfSecond,
			},
			cageFlags: []string{
				"--time-limit",
				"500ms",
			},
			returnsError: true,
			exitCodes:    []int{int(syscall.SIGXCPU), int(syscall.SIGKILL)},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("../../bin/fd8-judge", append([]string{"monster"}, test.cagedCommandFlags...)...)

			cmd = cage.EnsureRuntime(test.cage, nil).Encage(cmd)
			assert.Equal(t, os.Args[0], cmd.Path)
			expectedArgs := []string{
				filepath.Base(os.Args[0]),
				cage.CommandLineCommand,
			}
			expectedArgs = append(expectedArgs, test.cageFlags...)
			expectedArgs = append(expectedArgs,
				"--exec-path",
				"../../bin/fd8-judge",
				"--exec-args",
				"../../bin/fd8-judge",
				"--exec-args",
				"monster",
			)
			for _, arg := range test.cagedCommandFlags {
				expectedArgs = append(expectedArgs,
					"--exec-args",
					arg,
				)
			}
			assert.Equal(t, expectedArgs, cmd.Args)

			// replacing results asserted above because they won't work in a test environment
			cmd.Path = "../../bin/fd8-judge"
			cmd.Args = append([]string{cmd.Path}, cmd.Args[1:]...)

			outputBytes, err := cmd.Output()
			assert.Equal(t, test.returnsError, err != nil)
			assert.Equal(t, test.output, string(outputBytes))

			codes := map[int]bool{
				cmd.ProcessState.ExitCode():                      true,
				int(cmd.ProcessState.Sys().(syscall.WaitStatus)): true,
			}
			matches := 0
			for _, exitCode := range test.exitCodes {
				_, match := codes[exitCode]
				if match {
					matches++
				}
			}
			assert.Equal(t, true, matches > 0)
		})
	}
}
