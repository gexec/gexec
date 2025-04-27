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

type projectEnvironmentSecretListBind struct {
	ProjectID     string
	EnvironmentID string
	Format        string
}

// tmplProjectEnvironmentSecretList represents the list of project environment secrets.
var tmplProjectEnvironmentSecretList = "{{ range . }}ID: \x1b[33m{{ .ID }} \x1b[0m" + `
Kind: {{ .Kind }}
Name: {{ .Name }}
Content: {{ .Content }}

{{ end -}}
`

var (
	projectEnvironmentSecretListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all environment secrets",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentSecretListAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentSecretListArgs = projectEnvironmentSecretListBind{}
)

func init() {
	projectEnvironmentSecretCmd.AddCommand(projectEnvironmentSecretListCmd)

	projectEnvironmentSecretListCmd.Flags().StringVar(
		&projectEnvironmentSecretListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentSecretListCmd.Flags().StringVar(
		&projectEnvironmentSecretListArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentSecretListCmd.Flags().StringVar(
		&projectEnvironmentSecretListArgs.Format,
		"format",
		tmplProjectEnvironmentSecretList,
		"Custom output format",
	)
}

func projectEnvironmentSecretListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentSecretListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentSecretListArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide an environment ID or a slug")
	}

	resp, err := client.ShowProjectEnvironmentWithResponse(
		ccmd.Context(),
		projectEnvironmentSecretListArgs.ProjectID,
		projectEnvironmentSecretListArgs.EnvironmentID,
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
		fmt.Sprintln(projectEnvironmentSecretListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Secrets

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
