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
	"github.com/matheuscscp/fd8-judge/test/helpers"
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
		"time-limit-lower-than-one-second-does-not-break": {
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

			osArgs0 := os.Args[0]
			os.Args[0] = "go"

			encaged, err := cage.New(test.cage, nil).Encage(cmd)
			assert.Equal(t, nil, err)
			assert.Equal(t, os.Args[0], filepath.Base(encaged.Path))
			expectedArgs := []string{
				os.Args[0],
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
			assert.Equal(t, expectedArgs, encaged.Args)

			os.Args[0] = osArgs0

			helpers.ReplaceCageCommandPathAndArgs("../.." /* path to root */, encaged)

			outputBytes, err := encaged.Output()
			assert.Equal(t, test.returnsError, err != nil)
			assert.Equal(t, test.output, string(outputBytes))

			codes := map[int]bool{
				encaged.ProcessState.ExitCode():                      true,
				int(encaged.ProcessState.Sys().(syscall.WaitStatus)): true,
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
