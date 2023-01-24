package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ipcrm/sko-hol-ssrf/cli/cmd"
)

func main() {
	if err := cmd.Execute(cmd.NewReporterBackend()); err != nil {
		msg := color.HiRedString("x %s\n", err.Error())
		_, _ = fmt.Fprint(os.Stderr, msg)
		os.Exit(1)
	}
}
