// +build unit

package services_test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/matheuscscp/fd8-judge/pkg/services"
	mockServices "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestNewProgramServiceError(t *testing.T) {
	svc, err := services.NewProgramService("inv", nil)
	assert.Equal(t, nil, svc)
	assert.Equal(t, fmt.Errorf("invalid program service, want one in {'c++11'}, got 'inv'"), err)
}

func TestCompileError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockServices.MockProgramServiceRuntime

	type (
		testInput struct {
			programService     string
			ctx                context.Context
			sourceRelativePath string
			binaryRelativePath string
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func()
	}{
		"compile-c++11+error": {
			input: testInput{
				programService: "c++11",
				ctx:            context.TODO(),
			},
			output: testOutput{
				err: fmt.Errorf("error compiling for c++11: %w", fmt.Errorf("error")),
			},
			mocks: func() {
				mockRuntime.EXPECT().RunCommand(exec.CommandContext(context.TODO(), "g++", "-std=c++11", "", "-o", "")).Return(fmt.Errorf("error"))
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockServices.NewMockProgramServiceRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			programSvc, err := services.NewProgramService(test.input.programService, mockRuntime)
			assert.Equal(t, nil, err)
			err = programSvc.Compile(test.input.ctx, test.input.sourceRelativePath, test.input.binaryRelativePath)
			assert.Equal(t, test.output, testOutput{
				err: err,
			})
		})
	}
}
