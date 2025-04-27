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

// tmplProjectExecutionShow represents a project execution within details view.
var tmplProjectExecutionShow = "Name: \x1b[33m{{ .Template.Slug }}{{ .Name }} \x1b[0m" + `
ID: {{ .ID }}
{{ with .Template -}}
Template: {{ .Slug }}
{{ end -}}
Status: {{ .Status }}
Debug: {{ .Debug }}
{{ with .Playbook -}}
Playbook: {{ . }}
{{ end -}}
{{ with .Environment -}}
Environment: {{ . }}
{{ end -}}
{{ with .Secret -}}
Secret: {{ . }}
{{ end -}}
{{ with .Limit -}}
Limit: {{ . }}
{{ end -}}
{{ with .Branch -}}
Branch: {{ . }}
{{ end -}}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectExecutionShowBind struct {
	ProjectID   string
	ExecutionID string
	Format      string
}

var (
	projectExecutionShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project execution",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectExecutionShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectExecutionShowArgs = projectExecutionShowBind{}
)

func init() {
	projectExecutionCmd.AddCommand(projectExecutionShowCmd)

	projectExecutionShowCmd.Flags().StringVar(
		&projectExecutionShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectExecutionShowCmd.Flags().StringVar(
		&projectExecutionShowArgs.ExecutionID,
		"execution-id",
		"",
		"Execution ID or slug",
	)

	projectExecutionShowCmd.Flags().StringVar(
		&projectExecutionShowArgs.Format,
		"format",
		tmplProjectExecutionShow,
		"Custom output format",
	)
}

func projectExecutionShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectExecutionShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectExecutionShowArgs.ExecutionID == "" {
		return fmt.Errorf("you must provide a execution ID or a slug")
	}

	resp, err := client.ShowProjectExecutionWithResponse(
		ccmd.Context(),
		projectExecutionShowArgs.ProjectID,
		projectExecutionShowArgs.ExecutionID,
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
		fmt.Sprintln(projectExecutionShowArgs.Format),
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
