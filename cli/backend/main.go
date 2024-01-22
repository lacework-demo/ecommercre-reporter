package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/lacework-demo/ecommercre-reporter/cli/cmd"
)

func main() {
	if err := cmd.Execute(cmd.NewReporterBackend()); err != nil {
		msg := color.HiRedString("x %s\n", err.Error())
		_, _ = fmt.Fprint(os.Stderr, msg)
		os.Exit(1)
	}
}
