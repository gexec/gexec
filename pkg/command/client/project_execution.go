package command

import (
	"github.com/spf13/cobra"
)

var (
	projectExecutionCmd = &cobra.Command{
		Use:   "execution",
		Short: "Project execution commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectExecutionCmd)
}
