package command

import (
	"github.com/spf13/cobra"
)

var (
	projectCmd = &cobra.Command{
		Use:   "project",
		Short: "Project related sub-commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(projectCmd)
}
