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

type projectInventoryListBind struct {
	ProjectID string
	Format    string
}

// tmplProjectInventoryList represents a row within project inventory listing.
var tmplProjectInventoryList = "{{ range . }}Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}

{{ end -}}`

var (
	projectInventoryListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all inventories for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectInventoryListAction)
		},
		Args: cobra.NoArgs,
	}

	projectInventoryListArgs = projectInventoryListBind{}
)

func init() {
	projectInventoryCmd.AddCommand(projectInventoryListCmd)

	projectInventoryListCmd.Flags().StringVar(
		&projectInventoryListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectInventoryListCmd.Flags().StringVar(
		&projectInventoryListArgs.Format,
		"format",
		tmplProjectInventoryList,
		"Custom output format",
	)
}

func projectInventoryListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectInventoryListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectInventoriesWithResponse(
		ccmd.Context(),
		projectInventoryListArgs.ProjectID,
		&v1.ListProjectInventoriesParams{
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
		fmt.Sprintln(projectInventoryListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Inventories

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
