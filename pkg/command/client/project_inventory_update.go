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

type projectInventoryUpdateBind struct {
	ProjectID      string
	InventoryID    string
	RepositoryID   string
	NoRepositoryID bool
	CredentialID   string
	NoCredentialID bool
	BecomeID       string
	NoBecomeID     bool
	Slug           string
	Name           string
	Kind           string
	Content        string
	Format         string
}

var (
	projectInventoryUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project inventory",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectInventoryUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectInventoryUpdateArgs = projectInventoryUpdateBind{}
)

func init() {
	projectInventoryCmd.AddCommand(projectInventoryUpdateCmd)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.InventoryID,
		"inventory-id",
		"",
		"Inventory ID or slug",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.RepositoryID,
		"repository-id",
		"",
		"Repository for project inventory",
	)

	projectInventoryUpdateCmd.Flags().BoolVar(
		&projectInventoryUpdateArgs.NoRepositoryID,
		"no-repository-id",
		false,
		"Remove repository for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.CredentialID,
		"credential-id",
		"",
		"Credential for project inventory",
	)

	projectInventoryUpdateCmd.Flags().BoolVar(
		&projectInventoryUpdateArgs.NoCredentialID,
		"no-credential-id",
		false,
		"Remove credential for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.BecomeID,
		"become-id",
		"",
		"Become for project inventory",
	)

	projectInventoryUpdateCmd.Flags().BoolVar(
		&projectInventoryUpdateArgs.NoBecomeID,
		"no-become-id",
		false,
		"Remove become for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.Name,
		"name",
		"",
		"Name for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.Kind,
		"kind",
		"",
		"Kind for project inventory, can be static or file",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.Content,
		"content",
		"",
		"Content for project inventory",
	)

	projectInventoryUpdateCmd.Flags().StringVar(
		&projectInventoryUpdateArgs.Format,
		"format",
		tmplProjectInventoryShow,
		"Custom output format",
	)
}

func projectInventoryUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectInventoryUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectInventoryUpdateArgs.InventoryID == "" {
		return fmt.Errorf("you must provide a inventory ID or a slug")
	}

	body := v1.UpdateProjectInventoryJSONRequestBody{}
	changed := false

	if val := projectInventoryUpdateArgs.RepositoryID; val != "" {
		body.RepositoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.NoRepositoryID; val {
		body.RepositoryID = v1.ToPtr("")
		changed = true
	}

	if val := projectInventoryUpdateArgs.CredentialID; val != "" {
		body.CredentialID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.NoCredentialID; val {
		body.CredentialID = v1.ToPtr("")
		changed = true
	}

	if val := projectInventoryUpdateArgs.BecomeID; val != "" {
		body.BecomeID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.NoBecomeID; val {
		body.BecomeID = v1.ToPtr("")
		changed = true
	}

	if val := projectInventoryUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryUpdateArgs.Content; val != "" {
		body.Content = v1.ToPtr(val)
		changed = true
	}

	if !changed {
		fmt.Fprintln(os.Stderr, "Nothing to create...")
		return nil
	}

	tmpl, err := template.New(
		"_",
	).Funcs(
		globalFuncMap,
	).Funcs(
		basicFuncMap,
	).Parse(
		fmt.Sprintln(projectInventoryUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectInventoryWithResponse(
		ccmd.Context(),
		projectInventoryUpdateArgs.ProjectID,
		projectInventoryUpdateArgs.InventoryID,
		body,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if err := tmpl.Execute(
			os.Stdout,
			resp.JSON200,
		); err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	case http.StatusUnprocessableEntity:
		return validationError(resp.JSON422)
	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return errors.New(v1.FromPtr(resp.JSON400.Message))
		}

		return errors.New(http.StatusText(http.StatusBadRequest))
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
