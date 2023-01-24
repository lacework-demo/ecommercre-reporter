package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	ObjectStorageEndpoint string
	BucketName            string
	AccessKey             string
	SecretAccessKey       string
	StaticRegion          string
)

const (
	Name         = "ecomm-rpt"
	envVarPrefix = "ECOMM"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(cmd *cobra.Command) error {
	// Run command
	return cmd.Execute()
}

// initializeCmd is called once per command to parse flags and corresponding env vars
func initializeCmd(cmd *cobra.Command) error {
	// require all vars used for flag values be prefixed <envVarPrefix>_<thing>
	v := viper.New()
	v.SetEnvPrefix(envVarPrefix)
	v.AutomaticEnv()

	// bind env vars to each flag so it can be set in the ENV or via commands
	if err := bindFlags(cmd, v); err != nil {
		return errors.Wrap(err, "failed to bind flags")
	}

	return nil
}

// bindFlags binds cobra flag to its associated viper configuration
func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	var retError error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if retError != nil {
			return
		}

		// environment variables can't have dashes in them, so remove dashes
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", ""))
			envVarName := fmt.Sprintf("%s_%s", envVarPrefix, envVarSuffix)
			retError = v.BindEnv(f.Name, envVarName)
		}

		// if the flag wasn't set apply the viper configuration value if it has a value
		if retError == nil && !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			retError = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return retError
}
