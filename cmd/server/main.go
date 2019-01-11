package main

import (
	"github.com/tadoku/api/app"
)

func main() {
	deps := app.NewServerDependencies()
	app.RunServer(deps)
}
