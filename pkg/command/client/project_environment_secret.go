package command

import (
	"github.com/spf13/cobra"
)

var (
	projectEnvironmentSecretCmd = &cobra.Command{
		Use:   "secret",
		Short: "Secret management for project environment",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentSecretCmd)
}
