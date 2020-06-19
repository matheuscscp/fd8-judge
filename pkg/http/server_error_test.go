// +build unit

package http

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	mocks "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/http"
	mockInterfaces "github.com/matheuscscp/fd8-judge/test/mocks/gen/test/mocks"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestServeError(t *testing.T) {
	var mockRuntime *mocks.MockserverRuntime

	type (
		testInput struct {
			server *Server
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func(*gomock.Controller)
	}{
		"state-starting-error": {
			input: testInput{
				server: &Server{
					settleChannel: make(chan struct{}, 1),
					state:         starting,
				},
			},
			output: testOutput{
				err: errors.New("cannot start the server more than once"),
			},
			mocks: func(*gomock.Controller) {
				mockRuntime.EXPECT().Intn(16384).Return(0)
			},
		},
		"state-serving-error": {
			input: testInput{
				server: &Server{
					settleChannel: make(chan struct{}, 1),
					state:         serving,
				},
			},
			output: testOutput{
				err: errors.New("cannot start the server more than once"),
			},
			mocks: func(*gomock.Controller) {
				mockRuntime.EXPECT().Intn(16384).Return(0)
			},
		},
		"state-stopped-error": {
			input: testInput{
				server: &Server{
					settleChannel: make(chan struct{}, 1),
					state:         stopped,
				},
			},
			output: testOutput{
				err: errors.New("cannot start the server more than once"),
			},
			mocks: func(*gomock.Controller) {
				mockRuntime.EXPECT().Intn(16384).Return(0)
			},
		},
		"open-listeners-error": {
			input: testInput{
				server: &Server{
					HTTPEndpoint:  ":12345",
					HTTPSEndpoint: ":12346",
					settleChannel: make(chan struct{}, 1),
					state:         idle,
				},
			},
			output: testOutput{
				err: fmt.Errorf("error listening http at ':12346': %w", fmt.Errorf("error")),
			},
			mocks: func(ctrl *gomock.Controller) {
				mockListener := mockInterfaces.NewMockListener(ctrl)
				mockRuntime.EXPECT().Intn(16384).Return(0)
				mockRuntime.EXPECT().Listen("tcp", ":12345").Return(mockListener, nil)
				mockRuntime.EXPECT().Listen("tcp", ":12346").Return(nil, fmt.Errorf("error"))
				mockListener.EXPECT().Close().Return(nil)
			},
		},
		"create-handler-error": {
			input: testInput{
				server: &Server{
					HandlerFactory: func(context.Context, string) (nethttp.Handler, error) {
						return nil, fmt.Errorf("error")
					},
					settleChannel: make(chan struct{}, 1),
					state:         idle,
				},
			},
			output: testOutput{
				err: fmt.Errorf("error creating http server handler: %w", fmt.Errorf("error")),
			},
			mocks: func(*gomock.Controller) {
				mockRuntime.EXPECT().Intn(16384).Return(0)
				mockRuntime.EXPECT().Listen("tcp", gomock.Any()).Return(nil, nil)
			},
		},
		"stop-error": {
			input: testInput{
				server: &Server{
					HTTPRandomPort:  true,
					HTTPSRandomPort: true,
					HandlerFactory: func(context.Context, string) (nethttp.Handler, error) {
						return nil, nil
					},
					Logger:        logrus.WithField("app", "test"),
					stopChannel:   make(chan struct{}, 3),
					settleChannel: make(chan struct{}, 1),
					state:         idle,
				},
			},
			output: testOutput{
				err: &multierror.Error{Errors: []error{
					fmt.Errorf("error"),
					fmt.Errorf("error"),
					fmt.Errorf("error"),
				}},
			},
			mocks: func(ctrl *gomock.Controller) {
				httpListener := mockInterfaces.NewMockListener(ctrl)
				httpsListener := mockInterfaces.NewMockListener(ctrl)
				internalListener := mockInterfaces.NewMockListener(ctrl)
				mockRuntime.EXPECT().Intn(16382).Return(0)
				mockRuntime.EXPECT().Listen("tcp", "localhost:49152").Return(httpListener, nil)
				mockRuntime.EXPECT().Listen("tcp", "localhost:49153").Return(httpsListener, nil)
				mockRuntime.EXPECT().Listen("tcp", "localhost:49154").Return(internalListener, nil)
				mockRuntime.EXPECT().Serve(gomock.Any(), httpListener).Return(fmt.Errorf("error"))
				mockRuntime.EXPECT().Serve(gomock.Any(), httpsListener).Return(fmt.Errorf("error"))
				mockRuntime.EXPECT().Serve(gomock.Any(), internalListener).Return(fmt.Errorf("error"))
				httpListener.EXPECT().Close().Return(nil)
				httpsListener.EXPECT().Close().Return(nil)
				internalListener.EXPECT().Close().Return(nil)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mocks.NewMockserverRuntime(ctrl)
			if test.mocks != nil {
				test.mocks(ctrl)
			}

			test.input.server.runtime = mockRuntime
			assert.Equal(t, test.output.err, test.input.server.Serve())
		})
	}
}

func TestWaitForReadyError(t *testing.T) {
	var mockRuntime *mocks.MockserverRuntime

	type (
		testInput struct {
			server  *Server
			timeout time.Duration
		}
		testOutput struct {
			err error
		}
	)
	var tests = map[string]struct {
		input  testInput
		output testOutput
		mocks  func(*gomock.Controller, *testInput)
	}{
		"state-serving-nil-error": {
			input: testInput{
				server: &Server{
					state: serving,
				},
			},
		},
		"timeout-settle-endpoints-error": {
			input: testInput{
				server: &Server{
					state: starting,
				},
			},
			output: testOutput{
				err: errors.New("timed out waiting for server to settle endpoints"),
			},
		},
		"timeout-healthy-error": {
			input: testInput{
				server: &Server{
					settleChannel: make(chan struct{}, 1),
					state:         starting,
				},
			},
			output: testOutput{
				err: errors.New("timed out waiting for server to be healthy"),
			},
			mocks: func(_ *gomock.Controller, input *testInput) {
				input.server.settleChannel <- struct{}{}
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRuntime = mocks.NewMockserverRuntime(ctrl)
			if test.mocks != nil {
				test.mocks(ctrl, &test.input)
			}

			test.input.server.runtime = mockRuntime
			err := test.input.server.WaitForReady(test.input.timeout)
			assert.Equal(t, test.output.err, err)
		})
	}
}
