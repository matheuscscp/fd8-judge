package main

import (
	"github.com/matheuscscp/fd8-judge/application"
)

func main() {
	application.SetupConfigs()

	var app application.App

	switch application.GetAppType() {
	case application.ApiServerAppType:
		app = application.NewApiServerApp()
	case application.CliAppType:
		app = application.NewCliApp()
	default:
		app = application.NewCliApp()
	}

	app.Run()
	app.Shutdown()
}
