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

// tmplRunnerShow represents a project runner within details view.
var tmplRunnerShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
{{ with .Project- }}
Project: {{ .Name }}
{{ end -}}
Token: {{ .Token }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}`

type runnerShowBind struct {
	RunnerID string
	Format   string
}

var (
	runnerShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, runnerShowAction)
		},
		Args: cobra.NoArgs,
	}

	runnerShowArgs = runnerShowBind{}
)

func init() {
	runnerCmd.AddCommand(runnerShowCmd)

	runnerShowCmd.Flags().StringVar(
		&runnerShowArgs.RunnerID,
		"runner-id",
		"",
		"Runner ID or slug",
	)

	runnerShowCmd.Flags().StringVar(
		&runnerShowArgs.Format,
		"format",
		tmplRunnerShow,
		"Custom output format",
	)
}

func runnerShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if runnerShowArgs.RunnerID == "" {
		return fmt.Errorf("you must provide a runner ID or a slug")
	}

	resp, err := client.ShowGlobalRunnerWithResponse(
		ccmd.Context(),
		runnerShowArgs.RunnerID,
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
		fmt.Sprintln(runnerShowArgs.Format),
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
