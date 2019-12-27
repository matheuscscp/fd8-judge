package api

import (
	nethttp "net/http"
	"os"

	"github.com/matheuscscp/fd8-judge/api/controllers"
	"github.com/matheuscscp/fd8-judge/cmd/helpers"
	"github.com/matheuscscp/fd8-judge/pkg/grpc"
	"github.com/matheuscscp/fd8-judge/pkg/http"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	defineStartCommand()
}

// defineStartCommand defines the start command.
func defineStartCommand() {
	server := &http.Server{
		HandlerFactory: grpc.GetHandlerFactory(nil, &controllers.Controller{}),
		HealthHandler: nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			w.WriteHeader(nethttp.StatusOK)
		}),
		Logger: logrus.WithField("app", "api"),
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start fd8-judge API server.",
		Long:  "Start fd8-judge API server.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			server = http.NewServer(server, nil)
			go helpers.HandleInterrupt(func(signal os.Signal) bool {
				server.Logger.WithField("signal", signal).Info("Caught interruption signal.")
				server.GracefulShutdown()
				return true
			})
			cmd.SilenceUsage = true
			return server.Serve()
		},
	}
	helpers.BindServerFlags(startCmd, server, "", "")
	rootCmd.AddCommand(startCmd)
}
