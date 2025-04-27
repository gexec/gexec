package command

import (
	"github.com/spf13/cobra"
)

var (
	projectCredentialCmd = &cobra.Command{
		Use:   "credential",
		Short: "Project credential commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectCredentialCmd)
}
