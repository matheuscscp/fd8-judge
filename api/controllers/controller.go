package controllers

import (
	protos "github.com/matheuscscp/fd8-judge/api/protogen"
	"github.com/matheuscscp/fd8-judge/pkg/grpc"

	googlegrpc "google.golang.org/grpc"
)

type (
	// Controller implements the gRPC entry-points of the fd8-judge API.
	Controller struct {
	}
)

// New creates a Controller returning it as a pkg/grpc.Registerable to ensure Controller implements
// this interface.
func New() grpc.Registerable {
	return &Controller{}
}

// Register implements pkg/grpc.Registerable (and ensures Controller implements the gRPC service).
func (c *Controller) Register(server *googlegrpc.Server) {
	protos.RegisterServiceServer(server, c)
}

// GetGatewayRegisterFunc implements pkg/grpc.Registerable.
func (c *Controller) GetGatewayRegisterFunc() grpc.GatewayRegisterFunc {
	return protos.RegisterServiceHandlerFromEndpoint
}
