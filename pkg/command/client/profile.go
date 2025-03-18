package command

import (
	"github.com/spf13/cobra"
)

var (
	profileCmd = &cobra.Command{
		Use:   "profile",
		Short: "Profile related sub-commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(profileCmd)
}
