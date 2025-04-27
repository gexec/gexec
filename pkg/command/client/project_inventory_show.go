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

// tmplProjectInventoryShow represents a project inventory within details view.
var tmplProjectInventoryShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}
{{ with .Repository -}}
Repository: {{ .Slug }}
{{ end -}}
{{ with .Credential -}}
Credential: {{ .Slug }}
{{ end -}}
{{ with .Become -}}
Become: {{ .Slug }}
{{ end -}}
Kind: {{ .Kind }}
Content: {{ .Content }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type projectInventoryShowBind struct {
	ProjectID   string
	InventoryID string
	Format      string
}

var (
	projectInventoryShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a project inventory",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectInventoryShowAction)
		},
		Args: cobra.NoArgs,
	}

	projectInventoryShowArgs = projectInventoryShowBind{}
)

func init() {
	projectInventoryCmd.AddCommand(projectInventoryShowCmd)

	projectInventoryShowCmd.Flags().StringVar(
		&projectInventoryShowArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectInventoryShowCmd.Flags().StringVar(
		&projectInventoryShowArgs.InventoryID,
		"inventory-id",
		"",
		"Inventory ID or slug",
	)

	projectInventoryShowCmd.Flags().StringVar(
		&projectInventoryShowArgs.Format,
		"format",
		tmplProjectInventoryShow,
		"Custom output format",
	)
}

func projectInventoryShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectInventoryShowArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectInventoryShowArgs.InventoryID == "" {
		return fmt.Errorf("you must provide a inventory ID or a slug")
	}

	resp, err := client.ShowProjectInventoryWithResponse(
		ccmd.Context(),
		projectInventoryShowArgs.ProjectID,
		projectInventoryShowArgs.InventoryID,
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
		fmt.Sprintln(projectInventoryShowArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if err := tmpl.Execute(
			os.Stdout,
			resp.JSON200,
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
