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

// tmplProjectScheduleShow represents a project schedule within details view.
var tmplProjectScheduleShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
{{ with .Template -}}
Template: {{ .Slug }}
{{ end -}}
Cron: {{ .Cron }}
Active: {{ .Active }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectScheduleShowBind struct {
	ProjectID  string
	ScheduleID string
	Format     string
}

var (
	projectScheduleShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project schedule",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectScheduleShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectScheduleShowArgs = projectScheduleShowBind{}
)

func init() {
	projectScheduleCmd.AddCommand(projectScheduleShowCmd)

	projectScheduleShowCmd.Flags().StringVar(
		&projectScheduleShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectScheduleShowCmd.Flags().StringVar(
		&projectScheduleShowArgs.ScheduleID,
		"schedule-id",
		"",
		"Schedule ID or slug",
	)

	projectScheduleShowCmd.Flags().StringVar(
		&projectScheduleShowArgs.Format,
		"format",
		tmplProjectScheduleShow,
		"Custom output format",
	)
}

func projectScheduleShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectScheduleShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectScheduleShowArgs.ScheduleID == "" {
		return fmt.Errorf("you must provide a schedule ID or a slug")
	}

	resp, err := client.ShowProjectScheduleWithResponse(
		ccmd.Context(),
		projectScheduleShowArgs.ProjectID,
		projectScheduleShowArgs.ScheduleID,
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
		fmt.Sprintln(projectScheduleShowArgs.Format),
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
