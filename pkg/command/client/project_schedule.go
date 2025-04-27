package command

import (
	"github.com/spf13/cobra"
)

var (
	projectScheduleCmd = &cobra.Command{
		Use:   "schedule",
		Short: "Project schedule commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectScheduleCmd)
}
