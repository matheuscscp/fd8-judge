// +build unit

package cage_test

import (
	"os"
	"testing"
	"time"

	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/matheuscscp/fd8-judge/pkg/cage"
	mockCage "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/cage"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	var mockRuntime *mockCage.MockDefaultCageRuntime

	second := time.Second

	var tests = map[string]struct {
		cage  *cage.DefaultCage
		err   error
		mocks func()
	}{
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
		"exec-error": {
			cage: &cage.DefaultCage{},
			err:  fmt.Errorf("error exec()ing command: %w", fmt.Errorf("error")),
			mocks: func() {
				mockRuntime.EXPECT().Exec("", nil, os.Environ()).Return(fmt.Errorf("error"))
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mockCage.NewMockDefaultCageRuntime(ctrl)
			if test.mocks != nil {
				test.mocks()
			}

			cage := cage.New(test.cage, mockRuntime)
			err := cage.Execute()
			assert.Equal(t, test.err, err)
		})
	}
}
