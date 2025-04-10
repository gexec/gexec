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

// tmplProjectRunnerShow represents a project runner within details view.
var tmplProjectRunnerShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
Token: {{ .Token }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}`

type projectRunnerShowBind struct {
	ProjectID string
	RunnerID  string
	Format    string
}

var (
	projectRunnerShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRunnerShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectRunnerShowArgs = projectRunnerShowBind{}
)

func init() {
	projectRunnerCmd.AddCommand(projectRunnerShowCmd)

	projectRunnerShowCmd.Flags().StringVar(
		&projectRunnerShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRunnerShowCmd.Flags().StringVar(
		&projectRunnerShowArgs.RunnerID,
		"runner-id",
		"",
		"Runner ID or slug",
	)

	projectRunnerShowCmd.Flags().StringVar(
		&projectRunnerShowArgs.Format,
		"format",
		tmplProjectRunnerShow,
		"Custom output format",
	)
}

func projectRunnerShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRunnerShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRunnerShowArgs.RunnerID == "" {
		return fmt.Errorf("you must provide a runner ID or a slug")
	}

	resp, err := client.ShowProjectRunnerWithResponse(
		ccmd.Context(),
		projectRunnerShowArgs.ProjectID,
		projectRunnerShowArgs.RunnerID,
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
		fmt.Sprintln(projectRunnerShowArgs.Format),
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
