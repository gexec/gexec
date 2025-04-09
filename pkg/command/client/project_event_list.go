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

type projectEventListBind struct {
	Project string
	Format  string
}

// tmplProjectEventList represents a row within project event listing.
var tmplProjectEventList = "{{ range . }}Created: \x1b[33m{{ .CreatedAt }} \x1b[0m" + `
{{ with .UserDisplay -}}
User: {{ . }}
{{ end -}}
{{ with .ObjectDisplay -}}
Object: {{ . }}
{{ end -}}
Action: {{ .Action }}
{{ with .Attrs -}}
{{ range $key, $val := . -}}
{{ $key | camelize }}: {{ $val }}
{{ end -}}
{{ end }}
{{ end -}}`

var (
	projectEventListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all events for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEventListAction)
		},
		Args: cobra.NoArgs,
	}

	projectEventListArgs = projectEventListBind{}
)

func init() {
	projectEventCmd.AddCommand(projectEventListCmd)

	projectEventListCmd.Flags().StringVarP(
		&projectEventListArgs.Project,
		"project",
		"p",
		"",
		"Project ID or slug",
	)

	projectEventListCmd.Flags().StringVar(
		&projectEventListArgs.Format,
		"format",
		tmplProjectEventList,
		"Custom output format",
	)
}

func projectEventListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEventListArgs.Project == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectEventsWithResponse(
		ccmd.Context(),
		projectEventListArgs.Project,
		&v1.ListProjectEventsParams{
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
		fmt.Sprintln(projectEventListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Events

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
