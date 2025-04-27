package command

import (
	"github.com/spf13/cobra"
)

var (
	projectTemplateSurveyCmd = &cobra.Command{
		Use:   "survey",
		Short: "Survey management for project template",
		Args:  cobra.NoArgs,
	}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateSurveyCmd)
}
