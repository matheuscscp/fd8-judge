package helpers

import (
	"github.com/matheuscscp/fd8-judge/pkg/http"

	"github.com/spf13/cobra"
)

// BindServerFlags binds the server inputs for a command.
func BindServerFlags(cmd *cobra.Command, server *http.Server, flagPrefix, descriptionPrefix string) {
	cmd.Flags().StringVar(
		&server.HTTPEndpoint,
		flagPrefix+"http-endpoint",
		":8080",
		descriptionPrefix+"TCP endpoint to listen for the main HTTP server.",
	)
	cmd.Flags().StringVar(
		&server.InternalEndpoint,
		flagPrefix+"internal-endpoint",
		":8081",
		descriptionPrefix+"TCP endpoint to listen for the internal HTTP endpoints.",
	)
}
