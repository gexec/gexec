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

type projectTemplateVaultUpdateBind struct {
	ProjectID  string
	TemplateID string
	VaultID    string

	// TODO: add attributes

	Format string
}

var (
	projectTemplateVaultUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an template vault",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateVaultUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateVaultUpdateArgs = projectTemplateVaultUpdateBind{}
)

func init() {
	projectTemplateVaultCmd.AddCommand(projectTemplateVaultUpdateCmd)

	projectTemplateVaultUpdateCmd.Flags().StringVar(
		&projectTemplateVaultUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateVaultUpdateCmd.Flags().StringVar(
		&projectTemplateVaultUpdateArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateVaultUpdateCmd.Flags().StringVar(
		&projectTemplateVaultUpdateArgs.VaultID,
		"vault-id",
		"",
		"Vault ID or slug",
	)

	// TODO: add attributes kind/name/content

	projectTemplateVaultUpdateCmd.Flags().StringVar(
		&projectTemplateVaultUpdateArgs.Format,
		"format",
		tmplProjectTemplateVaultShow,
		"Custom output format",
	)
}

func projectTemplateVaultUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateVaultUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateVaultUpdateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide an template ID or a slug")
	}

	if projectTemplateVaultUpdateArgs.VaultID == "" {
		return fmt.Errorf("you must provide a vault ID or a slug")
	}

	body := v1.UpdateProjectTemplateVaultJSONRequestBody{}
	changed := false

	// TODO: add attributes

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
		fmt.Sprintln(projectTemplateVaultUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectTemplateVaultWithResponse(
		ccmd.Context(),
		projectTemplateVaultUpdateArgs.ProjectID,
		projectTemplateVaultUpdateArgs.TemplateID,
		projectTemplateVaultUpdateArgs.VaultID,
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
