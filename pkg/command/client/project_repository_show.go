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

// tmplProjectRepositoryShow represents a project repository within details view.
var tmplProjectRepositoryShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
{{ with .Credential -}}
Credential: {{ .Slug }}
{{ end -}}
URL: {{ .URL }}
Branch: {{ .Branch }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectRepositoryShowBind struct {
	ProjectID    string
	RepositoryID string
	Format       string
}

var (
	projectRepositoryShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project repository",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRepositoryShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectRepositoryShowArgs = projectRepositoryShowBind{}
)

func init() {
	projectRepositoryCmd.AddCommand(projectRepositoryShowCmd)

	projectRepositoryShowCmd.Flags().StringVar(
		&projectRepositoryShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRepositoryShowCmd.Flags().StringVar(
		&projectRepositoryShowArgs.RepositoryID,
		"repository-id",
		"",
		"Repository ID or slug",
	)

	projectRepositoryShowCmd.Flags().StringVar(
		&projectRepositoryShowArgs.Format,
		"format",
		tmplProjectRepositoryShow,
		"Custom output format",
	)
}

func projectRepositoryShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRepositoryShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRepositoryShowArgs.RepositoryID == "" {
		return fmt.Errorf("you must provide a repository ID or a slug")
	}

	resp, err := client.ShowProjectRepositoryWithResponse(
		ccmd.Context(),
		projectRepositoryShowArgs.ProjectID,
		projectRepositoryShowArgs.RepositoryID,
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
		fmt.Sprintln(projectRepositoryShowArgs.Format),
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
