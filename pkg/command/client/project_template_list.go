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

type projectTemplateListBind struct {
	ProjectID string
	Format    string
}

// tmplProjectTemplateList represents a row within project template listing.
var tmplProjectTemplateList = "{{ range . }}Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}

{{ end -}}`

var (
	projectTemplateListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all templates for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateListAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateListArgs = projectTemplateListBind{}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateListCmd)

	projectTemplateListCmd.Flags().StringVar(
		&projectTemplateListArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateListCmd.Flags().StringVar(
		&projectTemplateListArgs.Format,
		"format",
		tmplProjectTemplateList,
		"Custom output format",
	)
}

func projectTemplateListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateListArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectTemplatesWithResponse(
		ccmd.Context(),
		projectTemplateListArgs.ProjectID,
		&v1.ListProjectTemplatesParams{
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
		fmt.Sprintln(projectTemplateListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Templates

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
