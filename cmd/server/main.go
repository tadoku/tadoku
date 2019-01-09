package main

import (
	"github.com/tadoku/api/app"
)

func main() {
	s := app.NewServer()
	app.RunServer(s)
}
