package command

import (
	"github.com/spf13/cobra"
)

var (
	projectEnvironmentCmd = &cobra.Command{
		Use:   "environment",
		Short: "Project environment commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectEnvironmentCmd)
}
