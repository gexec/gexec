package command

import (
	"github.com/spf13/cobra"
)

var (
	projectInventoryCmd = &cobra.Command{
		Use:   "inventory",
		Short: "Project inventory commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectInventoryCmd)
}
