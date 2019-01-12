package main

import (
	"github.com/tadoku/api/app"
)

func main() {
	deps := app.NewServerDependencies()
	deps.AutoConfigure()
	app.RunServer(deps)
}
