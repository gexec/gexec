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

type groupUserListBind struct {
	GroupID string
	Format  string
}

// tmplGroupUserList represents a row within group user listing.
var tmplGroupUserList = "{{ range . }}Slug: \x1b[33m{{ .User.Username }} \x1b[0m" + `
ID: {{ .User.ID }}
Email: {{ .User.Email }}
Perm: {{ .Perm }}

{{ end -}}`

var (
	groupUserListCmd = &cobra.Command{
		Use:   "list",
		Short: "List assigned users for a group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupUserListAction)
		},
		Args: cobra.NoArgs,
	}

	groupUserListArgs = groupUserListBind{}
)

func init() {
	groupUserCmd.AddCommand(groupUserListCmd)

	groupUserListCmd.Flags().StringVar(
		&groupUserListArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupUserListCmd.Flags().StringVar(
		&groupUserListArgs.Format,
		"format",
		tmplGroupUserList,
		"Custom output format",
	)
}

func groupUserListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUserListArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	resp, err := client.ListGroupUsersWithResponse(
		ccmd.Context(),
		groupUserListArgs.GroupID,
		&v1.ListGroupUsersParams{
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
		fmt.Sprintln(groupUserListArgs.Format),
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
