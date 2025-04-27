package command

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectTemplateSurveyListBind struct {
	ProjectID  string
	TemplateID string
	Format     string
}

// tmplProjectTemplateSurveyList represents the list of project template surveys.
var tmplProjectTemplateSurveyList = "{{ range . }}ID: \x1b[33m{{ .ID }} \x1b[0m" + `
Name: {{ .Name }}

{{ end -}}
`

var (
	projectTemplateSurveyListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all template surveys",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateSurveyListAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateSurveyListArgs = projectTemplateSurveyListBind{}
)

func init() {
	projectTemplateSurveyCmd.AddCommand(projectTemplateSurveyListCmd)

	projectTemplateSurveyListCmd.Flags().StringVar(
		&projectTemplateSurveyListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateSurveyListCmd.Flags().StringVar(
		&projectTemplateSurveyListArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateSurveyListCmd.Flags().StringVar(
		&projectTemplateSurveyListArgs.Format,
		"format",
		tmplProjectTemplateSurveyList,
		"Custom output format",
	)
}

func projectTemplateSurveyListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateSurveyListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateSurveyListArgs.TemplateID == "" {
		return fmt.Errorf("you must provide an template ID or a slug")
	}

	resp, err := client.ShowProjectTemplateWithResponse(
		ccmd.Context(),
		projectTemplateSurveyListArgs.ProjectID,
		projectTemplateSurveyListArgs.TemplateID,
	)

	if err != nil {
		return err
	}

	tmpl, err := template.New(
		"_",
	).Funcs(
		globalFuncMap,
	).Funcs(
		basicFuncMap,
	).Parse(
		fmt.Sprintln(projectTemplateSurveyListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Surveys

		if records == nil || len(v1.FromPtr(records)) == 0 {
			fmt.Fprintln(os.Stderr, "Empty result")
			return nil
		}

		if err := tmpl.Execute(
			os.Stdout,
			records,
		); err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	case http.StatusForbidden:
		if resp.JSON403 != nil {
			return errors.New(v1.FromPtr(resp.JSON403.Message))
		}

		return errors.New(http.StatusText(http.StatusForbidden))
	case http.StatusNotFound:
		if resp.JSON404 != nil {
			return errors.New(v1.FromPtr(resp.JSON404.Message))
		}

		return errors.New(http.StatusText(http.StatusNotFound))
	case http.StatusInternalServerError:
		if resp.JSON500 != nil {
			return errors.New(v1.FromPtr(resp.JSON500.Message))
		}

		return errors.New(http.StatusText(http.StatusInternalServerError))
	case http.StatusUnauthorized:
		return ErrMissingRequiredCredentials
	default:
		return ErrUnknownServerResponse
	}

	return nil
}
