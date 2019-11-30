// +build unit

package cage_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/matheuscscp/fd8-judge/pkg/cage"
	mockCage "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/cage"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestEncageError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockCage.MockdefaultCageRuntime

	var tests = map[string]struct {
		cage    *cage.DefaultCage
		monster *exec.Cmd
		encaged *exec.Cmd
		err     error
		mocks   func()
	}{
		"look-path-error": {
			cage: &cage.DefaultCage{},
			err:  fmt.Errorf("error looking path for fd8-judge: %w", fmt.Errorf("error")),
			mocks: func() {
				mockRuntime.EXPECT().LookPath("fd8-judge").Return("", fmt.Errorf("error"))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockCage.NewMockdefaultCageRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			osArgs0 := os.Args[0]
			os.Args[0] = "fd8-judge"

			cage := cage.New(test.cage, mockRuntime)
			encaged, err := cage.Encage(test.monster)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.encaged, encaged)
			assert.Equal(t, true, err != nil || encaged == test.monster)

			os.Args[0] = osArgs0
		})
	}
}

func TestExecuteError(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockCage.MockdefaultCageRuntime

	second := time.Second

	var tests = map[string]struct {
		cage  *cage.DefaultCage
		err   error
		mocks func()
	}{
		"exec-error": {
			cage: &cage.DefaultCage{},
			err:  fmt.Errorf("error in exec syscall: %w", fmt.Errorf("error")),
			mocks: func() {
				mockRuntime.EXPECT().Exec("", nil, nil).Return(fmt.Errorf("error"))
			},
		},
		"restrict-time-limit-error": {
			cage: &cage.DefaultCage{
				TimeLimit: &second,
			},
			err: fmt.Errorf("error restricting time limit: %w", fmt.Errorf("error")),
			mocks: func() {
				mockRuntime.EXPECT().Setrlimit(unix.RLIMIT_CPU, &unix.Rlimit{
					Cur: 1,
					Max: 1,
				}).Return(fmt.Errorf("error"))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockCage.NewMockdefaultCageRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			cage := cage.New(test.cage, mockRuntime)
			err := cage.Execute()
			assert.Equal(t, test.err, err)
		})
	}
}
