package cmd

import (
	"errors"
	"fmt"

	"github.com/ipcrm/sko-hol-ssrf/backend"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: Name,
	}
	return rootCmd
}

func newVersionCommand() *cobra.Command {
	// versionCmd represents the version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Long:  `Prints out the installed version of the CLI`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s cli v%s (sha:%s) (time:%s)", Name, Version, GitSHA, BuildTime)
		},
	}

	return versionCmd
}

func newReporterFrontend() *cobra.Command {
	// feCmd represents the version command
	feCmd := &cobra.Command{
		Use:   "frontend",
		Short: "Starts the fronted",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if ReporterEndpoint == "" {
				return errors.New("must pass --reporter-endpoint flag (-r)")
			}

			if DBHost == "" || DBUser == "" || DBPass == "" || DBName == "" {
				return errors.New("must set at minimum database name, host, user, and password")
			}

			connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s", DBUser, DBPass, fmt.Sprintf("%s:%s", DBHost, DBPort), DBName)
			backend.StartFrontend(ReporterEndpoint, DBType, connStr, SpaBuildRoot)
			return nil
		},
	}
	feCmd.PersistentFlags().StringVarP(&SpaBuildRoot, "app-build-path", "b", "./frontend/build", "SPA build output path")
	feCmd.PersistentFlags().StringVarP(&DBName, "database-name", "n", "", "database name")
	feCmd.PersistentFlags().StringVarP(&DBHost, "database-host", "H", "", "database host")
	feCmd.PersistentFlags().StringVarP(&DBPort, "database-port", "P", "3306", "database port")
	feCmd.PersistentFlags().StringVarP(&DBUser, "database-user", "u", "", "database user")
	feCmd.PersistentFlags().StringVarP(&DBPass, "database-pass", "p", "", "database password")
	feCmd.PersistentFlags().StringVarP(&DBType, "database-type", "t", "mysql", "database type")
	feCmd.PersistentFlags().StringVarP(&ReporterEndpoint, "reporter-endpoint", "r", "", "url for the reporter service")

	return feCmd
}

func newReporterBackend() *cobra.Command {
	// rbCmd represents the version command
	rbCmd := &cobra.Command{
		Use:   "backend",
		Short: "Starts the backend",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if BucketName == "" {
				return errors.New("must pass --bucket flag (-b)")
			}

			if AccessKey != "" || SecretAccessKey != "" {
				if AccessKey == "" || SecretAccessKey == "" {
					return errors.New("must set both access key and secret access key together")
				}
			}

			backend.StartReporter(ObjectStorageEndpoint, BucketName, AccessKey, SecretAccessKey, StaticRegion)
			return nil
		},
	}
	rbCmd.PersistentFlags().StringVarP(&ObjectStorageEndpoint, "object-storage-endpoint", "o", "", "url for the object storage api if not AWS S3")
	rbCmd.PersistentFlags().StringVarP(&BucketName, "bucket", "b", "", "bucket name for the object storage; required")
	rbCmd.PersistentFlags().StringVarP(&AccessKey, "accesskey", "a", "", "access key for object storage if not set in environment")
	rbCmd.PersistentFlags().StringVarP(&SecretAccessKey, "secretaccesskey", "s", "", "secret access key for object storage if not set in environment")
	rbCmd.PersistentFlags().StringVarP(&StaticRegion, "static-region", "r", "", "region to use for object storage if required")

	return rbCmd
}
