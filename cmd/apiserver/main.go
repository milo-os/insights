package main

import (
	"os"

	"k8s.io/component-base/cli"

	"github.com/datum-cloud/insights/cmd/apiserver/app"
)

func main() {
	command := app.NewServerCommand()
	code := cli.Run(command)
	os.Exit(code)
}
