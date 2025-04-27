package command

import (
	"github.com/spf13/cobra"
)

var (
	projectRepositoryCmd = &cobra.Command{
		Use:   "repository",
		Short: "Project repository commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectRepositoryCmd)
}
