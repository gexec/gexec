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

type userProjectListBind struct {
	UserID string
	Format string
}

// tmplUserProjectList represents a row within user project listing.
var tmplUserProjectList = "{{ range . }}Slug: \x1b[33m{{ .Project.Slug }} \x1b[0m" + `
ID: {{ .Project.ID }}
Name: {{ .Project.Name }}
Perm: {{ .Perm }}

{{ end -}}`

var (
	userProjectListCmd = &cobra.Command{
		Use:   "list",
		Short: "List assigned projects for a user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userProjectListAction)
		},
		Args: cobra.NoArgs,
	}

	userProjectListArgs = userProjectListBind{}
)

func init() {
	userProjectCmd.AddCommand(userProjectListCmd)

	userProjectListCmd.Flags().StringVar(
		&userProjectListArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userProjectListCmd.Flags().StringVar(
		&userProjectListArgs.Format,
		"format",
		tmplUserProjectList,
		"Custom output format",
	)
}

func userProjectListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userProjectListArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	resp, err := client.ListUserProjectsWithResponse(
		ccmd.Context(),
		userProjectListArgs.UserID,
		&v1.ListUserProjectsParams{
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
		fmt.Sprintln(userProjectListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Projects

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
