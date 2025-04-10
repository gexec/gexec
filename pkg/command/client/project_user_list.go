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

type projectUserListBind struct {
	ProjectID string
	Format    string
}

// tmplProjectUserList represents a row within project user listing.
var tmplProjectUserList = "{{ range . }}Slug: \x1b[33m{{ .User.Username }} \x1b[0m" + `
ID: {{ .User.ID }}
Email: {{ .User.Email }}
Perm: {{ .Perm }}

{{ end -}}`

var (
	projectUserListCmd = &cobra.Command{
		Use:   "list",
		Short: "List assigned users for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectUserListAction)
		},
		Args: cobra.NoArgs,
	}

	projectUserListArgs = projectUserListBind{}
)

func init() {
	projectUserCmd.AddCommand(projectUserListCmd)

	projectUserListCmd.Flags().StringVar(
		&projectUserListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectUserListCmd.Flags().StringVar(
		&projectUserListArgs.Format,
		"format",
		tmplProjectUserList,
		"Custom output format",
	)
}

func projectUserListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectUserListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectUsersWithResponse(
		ccmd.Context(),
		projectUserListArgs.ProjectID,
		&v1.ListProjectUsersParams{
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
		fmt.Sprintln(projectUserListArgs.Format),
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
