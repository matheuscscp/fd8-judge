package application

import "fmt"

type ApiServerApp struct {
	db    int
	cache int
}

const ApiServerAppType = "ApiServer"

func NewApiServerApp() *ApiServerApp {
	return &ApiServerApp{2, 5}
}

func (app *ApiServerApp) Run() {
	fmt.Printf("db=%d cache=%d\n", app.db, app.cache)
}

func (app *ApiServerApp) Shutdown() {
	fmt.Printf("shutting down api server...\n")
}
