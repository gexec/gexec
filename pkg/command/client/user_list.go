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

// tmplUserList represents a row within user listing.
var tmplUserList = "{{ range . }}Username: \x1b[33m{{ .Username }} \x1b[0m" + `
ID: {{ .ID }}
Email: {{ .Email }}

{{ end -}}`

type userListBind struct {
	Format string
}

var (
	userListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userListAction)
		},
		Args: cobra.NoArgs,
	}

	userListArgs = userListBind{}
)

func init() {
	userCmd.AddCommand(userListCmd)

	userListCmd.Flags().StringVar(
		&userListArgs.Format,
		"format",
		tmplUserList,
		"Custom output format",
	)
}

func userListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	resp, err := client.ListUsersWithResponse(
		ccmd.Context(),
		&v1.ListUsersParams{
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
		fmt.Sprintln(userListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Users

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
