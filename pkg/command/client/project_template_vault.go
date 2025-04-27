package command

import (
	"github.com/spf13/cobra"
)

var (
	projectTemplateVaultCmd = &cobra.Command{
		Use:   "vault",
		Short: "Vault management for project template",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateVaultCmd)
}
