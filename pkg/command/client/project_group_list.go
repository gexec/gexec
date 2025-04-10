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

type projectGroupListBind struct {
	ProjectID string
	Format    string
}

// tmplProjectGroupList represents a row within project group listing.
var tmplProjectGroupList = "{{ range . }}Slug: \x1b[33m{{ .Group.Slug }} \x1b[0m" + `
ID: {{ .Group.ID }}
Name: {{ .Group.Name }}
Perm: {{ .Perm }}

{{ end -}}`

var (
	projectGroupListCmd = &cobra.Command{
		Use:   "list",
		Short: "List assigned groups for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectGroupListAction)
		},
		Args: cobra.NoArgs,
	}

	projectGroupListArgs = projectGroupListBind{}
)

func init() {
	projectGroupCmd.AddCommand(projectGroupListCmd)

	projectGroupListCmd.Flags().StringVar(
		&projectGroupListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectGroupListCmd.Flags().StringVar(
		&projectGroupListArgs.Format,
		"format",
		tmplProjectGroupList,
		"Custom output format",
	)
}

func projectGroupListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectGroupListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectGroupsWithResponse(
		ccmd.Context(),
		projectGroupListArgs.ProjectID,
		&v1.ListProjectGroupsParams{
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
		fmt.Sprintln(projectGroupListArgs.Format),
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
