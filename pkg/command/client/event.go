package command

import (
	"github.com/spf13/cobra"
)

var (
	eventCmd = &cobra.Command{
		Use:   "event",
		Short: "Event related sub-commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(eventCmd)
}
