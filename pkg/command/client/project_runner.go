package command

import (
	"github.com/spf13/cobra"
)

var (
	projectRunnerCmd = &cobra.Command{
		Use:   "runner",
		Short: "Project runner commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectRunnerCmd)
}
