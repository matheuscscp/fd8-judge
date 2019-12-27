package api

import (
	root "github.com/matheuscscp/fd8-judge/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Start or call the fd8-judge API.",
	Long:  "Start a server serving fd8-judge API or call the API endpoints with CRUD operations.",
}

func init() {
	root.AddCommand(rootCmd)
}
