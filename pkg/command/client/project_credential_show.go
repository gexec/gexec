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

// tmplProjectCredentialShow represents a project credential within details view.
var tmplProjectCredentialShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
Override: {{ .Override }}
Kind: {{ .Kind }}
{{ with .Shell -}}
{{ if .Username -}}
Username: {{ .Username }}
{{ end -}}
{{ if .Password -}}
Password: {{ .Password }}
{{ end -}}
{{ if .PrivateKey -}}
PrivateKey: {{ .PrivateKey }}
{{ end -}}
{{ end -}}
{{ with .Login -}}
{{ if .Username -}}
Username: {{ .Username }}
{{ end -}}
{{ if .Password -}}
Password: {{ .Password }}
{{ end -}}
{{ end -}}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}`

type projectCredentialShowBind struct {
	ProjectID    string
	CredentialID string
	Format       string
}

var (
	projectCredentialShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project credential",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectCredentialShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectCredentialShowArgs = projectCredentialShowBind{}
)

func init() {
	projectCredentialCmd.AddCommand(projectCredentialShowCmd)

	projectCredentialShowCmd.Flags().StringVar(
		&projectCredentialShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectCredentialShowCmd.Flags().StringVar(
		&projectCredentialShowArgs.CredentialID,
		"credential-id",
		"",
		"Credential ID or slug",
	)

	projectCredentialShowCmd.Flags().StringVar(
		&projectCredentialShowArgs.Format,
		"format",
		tmplProjectCredentialShow,
		"Custom output format",
	)
}

func projectCredentialShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectCredentialShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectCredentialShowArgs.CredentialID == "" {
		return fmt.Errorf("you must provide a credential ID or a slug")
	}

	resp, err := client.ShowProjectCredentialWithResponse(
		ccmd.Context(),
		projectCredentialShowArgs.ProjectID,
		projectCredentialShowArgs.CredentialID,
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
		fmt.Sprintln(projectCredentialShowArgs.Format),
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
