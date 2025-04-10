package command

import (
	"github.com/spf13/cobra"
)

var (
	runnerCmd = &cobra.Command{
		Use:   "runner",
		Short: "Runner related sub-commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(runnerCmd)
}
