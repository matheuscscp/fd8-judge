// +build unit

package grpc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/matheuscscp/fd8-judge/pkg/grpc"
	mocks "github.com/matheuscscp/fd8-judge/test/mocks/gen/pkg/grpc"

	"github.com/golang/mock/gomock"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
)

func TestGetHandlerFactoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRegisterable := mocks.NewMockRegisterable(ctrl)
	mockRegisterable.EXPECT().Register(gomock.Any())
	mockRegisterable.EXPECT().GetGatewayRegisterFunc().Return(
		grpc.GatewayRegisterFunc(func(
			ctx context.Context,
			mux *runtime.ServeMux,
			endpoint string,
			opts []googlegrpc.DialOption,
		) error {
			return errors.New("error")
		}),
	)

	_, err := grpc.GetHandlerFactory(nil, mockRegisterable)(context.Background(), "")
	assert.Equal(t, fmt.Errorf("error registering to gRPC gateway: %w", errors.New("error")), err)
}
