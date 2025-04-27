package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"text/template"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

// tmplProjectTemplateShow represents a project template within details view.
var tmplProjectTemplateShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
Executor: {{ .Executor }}
{{ with .Repository -}}
Repository: {{ .ID }} {{ .Slug }}
{{ end -}}
{{ with .Branch -}}
Branch: {{ . }}
{{ end -}}
{{ with .Inventory -}}
Inventory: {{ .ID }} {{ .Slug }}
{{ end -}}
{{ with .Environment -}}
Environment: {{ .ID }} {{ .Slug }}
{{ end -}}
{{ with .Description -}}
Description: {{ . }}
{{ end -}}
{{ with .Playbook -}}
Playbook: {{ . }}
{{ end -}}
{{ with .Arguments -}}
Arguments: {{ . }}
{{ end -}}
{{ with .Limit -}}
Limit: {{ . }}
{{ end -}}
AllowOverride: {{ .AllowOverride }}
{{ with .Surveys -}}
Surveys: {{ len . }}
{{ else -}}
Surveys: 0
{{ end -}}
{{ with .Vaults -}}
Vaults: {{ len . }}
{{ else -}}
Vaults: 0
{{ end -}}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectTemplateShowBind struct {
	ProjectID  string
	TemplateID string
	Format     string
}

var (
	projectTemplateShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project template",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateShowArgs = projectTemplateShowBind{}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateShowCmd)

	projectTemplateShowCmd.Flags().StringVar(
		&projectTemplateShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateShowCmd.Flags().StringVar(
		&projectTemplateShowArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateShowCmd.Flags().StringVar(
		&projectTemplateShowArgs.Format,
		"format",
		tmplProjectTemplateShow,
		"Custom output format",
	)
}

func projectTemplateShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateShowArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template ID or a slug")
	}

	resp, err := client.ShowProjectTemplateWithResponse(
		ccmd.Context(),
		projectTemplateShowArgs.ProjectID,
		projectTemplateShowArgs.TemplateID,
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
		fmt.Sprintln(projectTemplateShowArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if err := tmpl.Execute(
			os.Stdout,
			resp.JSON200,
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
