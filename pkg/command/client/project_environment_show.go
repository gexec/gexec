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

// tmplProjectEnvironmentShow represents a project environment within details view.
var tmplProjectEnvironmentShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
{{ with .Secrets -}}
Secrets: {{ len . }}
{{ else -}}
Secrets: 0
{{ end -}}
{{ with .Values -}}
Values: {{ len . }}
{{ else -}}
Values: 0
{{ end -}}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectEnvironmentShowBind struct {
	ProjectID     string
	EnvironmentID string
	Format        string
}

var (
	projectEnvironmentShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project environment",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentShowArgs = projectEnvironmentShowBind{}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentShowCmd)

	projectEnvironmentShowCmd.Flags().StringVar(
		&projectEnvironmentShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentShowCmd.Flags().StringVar(
		&projectEnvironmentShowArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentShowCmd.Flags().StringVar(
		&projectEnvironmentShowArgs.Format,
		"format",
		tmplProjectEnvironmentShow,
		"Custom output format",
	)
}

func projectEnvironmentShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentShowArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide a environment ID or a slug")
	}

	resp, err := client.ShowProjectEnvironmentWithResponse(
		ccmd.Context(),
		projectEnvironmentShowArgs.ProjectID,
		projectEnvironmentShowArgs.EnvironmentID,
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
		fmt.Sprintln(projectEnvironmentShowArgs.Format),
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
