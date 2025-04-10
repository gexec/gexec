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

type eventListBind struct {
	Format string
}

// tmplEventList represents a row within event listing.
var tmplEventList = "{{ range . }}Created: \x1b[33m{{ .CreatedAt }} \x1b[0m" + `
{{ with .ProjectDisplay -}}
Project: {{ . }}
{{ end -}}
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
	eventListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all events",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, eventListAction)
		},
		Args: cobra.NoArgs,
	}

	eventListArgs = eventListBind{}
)

func init() {
	eventCmd.AddCommand(eventListCmd)

	eventListCmd.Flags().StringVar(
		&eventListArgs.Format,
		"format",
		tmplEventList,
		"Custom output format",
	)
}

func eventListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	resp, err := client.ListGlobalEventsWithResponse(
		ccmd.Context(),
		&v1.ListGlobalEventsParams{
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
		fmt.Sprintln(eventListArgs.Format),
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
