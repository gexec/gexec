package command

import (
	"github.com/spf13/cobra"
)

var (
	projectEnvironmentValueCmd = &cobra.Command{
		Use:   "value",
		Short: "Value management for project environment",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentValueCmd)
}
