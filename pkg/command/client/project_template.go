package command

import (
	"github.com/spf13/cobra"
)

var (
	projectTemplateCmd = &cobra.Command{
		Use:   "template",
		Short: "Project template commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectCmd.AddCommand(projectTemplateCmd)
}
