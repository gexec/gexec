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

type projectScheduleListBind struct {
	ProjectID string
	Format    string
}

// tmplProjectScheduleList represents a row within project schedule listing.
var tmplProjectScheduleList = "{{ range . }}Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}

{{ end -}}
`

var (
	projectScheduleListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all schedules for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectScheduleListAction)
		},
		Args: cobra.NoArgs,
	}

	projectScheduleListArgs = projectScheduleListBind{}
)

func init() {
	projectScheduleCmd.AddCommand(projectScheduleListCmd)

	projectScheduleListCmd.Flags().StringVar(
		&projectScheduleListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectScheduleListCmd.Flags().StringVar(
		&projectScheduleListArgs.Format,
		"format",
		tmplProjectScheduleList,
		"Custom output format",
	)
}

func projectScheduleListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectScheduleListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectSchedulesWithResponse(
		ccmd.Context(),
		projectScheduleListArgs.ProjectID,
		&v1.ListProjectSchedulesParams{
			Limit:  v1.ToPtr(10000),
			Offset: v1.ToPtr(0),
		},
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
		fmt.Sprintln(projectScheduleListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Schedules

		if len(records) == 0 {
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
