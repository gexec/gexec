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

// tmplProjectShow represents a user within details view.
var tmplProjectShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}`

type projectShowBind struct {
	ProjectID string
	Format    string
}

var (
	projectShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectShowArgs = projectShowBind{}
)

func init() {
	projectCmd.AddCommand(projectShowCmd)

	projectShowCmd.Flags().StringVar(
		&projectShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectShowCmd.Flags().StringVar(
		&projectShowArgs.Format,
		"format",
		tmplProjectShow,
		"Custom output format",
	)
}

func projectShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ShowProjectWithResponse(
		ccmd.Context(),
		projectShowArgs.ProjectID,
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
		fmt.Sprintln(projectShowArgs.Format),
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
