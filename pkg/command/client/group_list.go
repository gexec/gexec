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

// tmplGroupList represents a row within user listing.
var tmplGroupList = "{{ range . }}Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}

{{ end -}}`

type groupListBind struct {
	Format string
}

var (
	groupListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all groups",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupListAction)
		},
		Args: cobra.NoArgs,
	}

	groupListArgs = groupListBind{}
)

func init() {
	groupCmd.AddCommand(groupListCmd)

	groupListCmd.Flags().StringVar(
		&groupListArgs.Format,
		"format",
		tmplGroupList,
		"Custom output format",
	)
}

func groupListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	resp, err := client.ListGroupsWithResponse(
		ccmd.Context(),
		&v1.ListGroupsParams{
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
		fmt.Sprintln(groupListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Groups

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
