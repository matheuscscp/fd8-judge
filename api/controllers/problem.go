package controllers

import (
	"context"

	protos "github.com/matheuscscp/fd8-judge/api/protogen"
)

// CreateProblem creates a Problem.
func (c *Controller) CreateProblem(
	ctx context.Context,
	in *protos.CreateProblemRequest,
) (*protos.Problem, error) {
	return &protos.Problem{Key: in.GetProblem().GetKey()}, nil
}
