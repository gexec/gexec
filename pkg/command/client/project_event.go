package command

import (
	"github.com/spf13/cobra"
)

var (
	projectEventCmd = &cobra.Command{
		Use:   "event",
		Short: "Project event commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectEventCmd)
}
