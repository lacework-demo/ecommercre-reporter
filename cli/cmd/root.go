package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var (
	// All the following "unknown" variables are being injected at
	// build time via the cross-platform directive inside the Makefile
	//
	// Version is the semver coming from the VERSION file
	Version = "unknown"

	// GitSHA is the git ref that the cli was built from
	GitSHA = "unknown"

	// BuildTime is a human-readable time when the cli was built at
	BuildTime = "unknown"

	Logger zerolog.Logger

	Debug bool
	Trace bool

	ConfigurationPath     string
	AuthTokenPath         string
	ReporterEndpoint      string
	SpaBuildRoot          string
	DBName                string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPass                string
	DBType                string
	Endpoint              string
	Homedir               string
	ObjectStorageEndpoint string
	BucketName            string
	AccessKey             string
	SecretAccessKey       string
	StaticRegion          string
)

const (
	Name = "ecomm-reporter"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// determine homedir
	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "failed to locate homedir")
	}
	Homedir = home

	// init commands
	rootCmd := newRootCommand()
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newReporterFrontend())
	rootCmd.AddCommand(newReporterBackend())

	// first, verify if the user provided a command to execute,
	// if no command was provided, only print out the usage message
	if noCommandProvided() {
		errcheckWARN(rootCmd.Help())
		os.Exit(127)
	}

	// Run command
	return rootCmd.Execute()
}

// noCommandProvided checks if a command or argument was provided
func noCommandProvided() bool {
	t := len(os.Args) <= 1
	return t
}

func errcheckWARN(err error) {
	if err != nil {
		fmt.Printf("%s", err)
	}
}
