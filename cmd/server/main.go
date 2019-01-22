package main

import (
	"fmt"

	"github.com/tadoku/api/app"
)

func main() {
	deps := app.NewServerDependencies()
	err := deps.AutoConfigure()
	if err != nil {
		panic(fmt.Sprintf("Server cannot be started: %v\n", err))
	}

	app.RunServer(deps)
}
