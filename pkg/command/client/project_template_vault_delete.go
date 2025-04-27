package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectTemplateVaultDeleteBind struct {
	ProjectID  string
	TemplateID string
	VaultID    string
}

var (
	projectTemplateVaultDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an template vault",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateVaultDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateVaultDeleteArgs = projectTemplateVaultDeleteBind{}
)

func init() {
	projectTemplateVaultCmd.AddCommand(projectTemplateVaultDeleteCmd)

	projectTemplateVaultDeleteCmd.Flags().StringVar(
		&projectTemplateVaultDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateVaultDeleteCmd.Flags().StringVar(
		&projectTemplateVaultDeleteArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateVaultDeleteCmd.Flags().StringVar(
		&projectTemplateVaultDeleteArgs.VaultID,
		"vault-id",
		"",
		"Vault ID or slug",
	)
}

func projectTemplateVaultDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateVaultDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateVaultDeleteArgs.TemplateID == "" {
		return fmt.Errorf("you must provide an template ID or a slug")
	}

	if projectTemplateVaultDeleteArgs.VaultID == "" {
		return fmt.Errorf("you must provide a vault ID or a slug")
	}

	resp, err := client.DeleteProjectTemplateVaultWithResponse(
		ccmd.Context(),
		projectTemplateVaultDeleteArgs.ProjectID,
		projectTemplateVaultDeleteArgs.TemplateID,
		projectTemplateVaultDeleteArgs.VaultID,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Fprintln(os.Stderr, "Successfully deleted")
	case http.StatusForbidden:
		if resp.JSON403 != nil {
			return errors.New(v1.FromPtr(resp.JSON403.Message))
		}

		return errors.New(http.StatusText(http.StatusForbidden))
	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return errors.New(v1.FromPtr(resp.JSON400.Message))
		}

		return errors.New(http.StatusText(http.StatusBadRequest))
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
