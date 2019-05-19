package application

import "fmt"

type CliApp struct {
	db int
}

const CliAppType = "Cli"

func NewCliApp() *CliApp {
	return &CliApp{7}
}

func (app *CliApp) Run() {
	fmt.Printf("db=%d\n", app.db)
}

func (app *CliApp) Shutdown() {
	fmt.Printf("shutting down cli...\n")
}
