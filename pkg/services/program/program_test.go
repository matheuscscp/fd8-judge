// +build integration

package program_test

import (
	"context"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/services/program"
	"github.com/matheuscscp/fd8-judge/test/factories"
	"github.com/matheuscscp/fd8-judge/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestCompileAndExecute(t *testing.T) {
	programFactory := &factories.ProgramFactory{}

	var tests = map[string]struct {
		programService string
		program        string
		sourcePath     string
		binaryPath     string
		output         string
	}{
		"c++11": {
			programService: "c++11",
			program:        fixtures.ProgramCpp11HelloWorld,
			sourcePath:     "./c++11HelloWorld.cpp",
			binaryPath:     "./c++11HelloWorld",
			output:         "hello, world!\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc, err := program.NewService(test.programService, nil)
			assert.Equal(t, nil, err)

			err = programFactory.Create(test.program, test.sourcePath)
			assert.Equal(t, nil, err)

			err = svc.Compile(context.TODO(), test.sourcePath, test.binaryPath)
			assert.Equal(t, nil, err)

			cmd := svc.GetExecutionCommand(context.TODO(), test.sourcePath, test.binaryPath)

			output, err := cmd.Output()
			assert.Equal(t, nil, err)
			assert.Equal(t, []byte(test.output), output)

			forestToRemove := []factories.FileTreeNode{
				&factories.File{Name: test.binaryPath},
				&factories.File{Name: test.sourcePath},
			}
			for _, tree := range forestToRemove {
				err = tree.Remove(".")
				assert.Equal(t, nil, err)
			}
		})
	}
}
