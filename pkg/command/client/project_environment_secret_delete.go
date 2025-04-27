package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectEnvironmentSecretDeleteBind struct {
	ProjectID     string
	EnvironmentID string
	SecretID      string
}

var (
	projectEnvironmentSecretDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an environment secret",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentSecretDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentSecretDeleteArgs = projectEnvironmentSecretDeleteBind{}
)

func init() {
	projectEnvironmentSecretCmd.AddCommand(projectEnvironmentSecretDeleteCmd)

	projectEnvironmentSecretDeleteCmd.Flags().StringVar(
		&projectEnvironmentSecretDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentSecretDeleteCmd.Flags().StringVar(
		&projectEnvironmentSecretDeleteArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentSecretDeleteCmd.Flags().StringVar(
		&projectEnvironmentSecretDeleteArgs.SecretID,
		"secret-id",
		"",
		"Secret ID or slug",
	)
}

func projectEnvironmentSecretDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentSecretDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentSecretDeleteArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide an environment ID or a slug")
	}

	if projectEnvironmentSecretDeleteArgs.SecretID == "" {
		return fmt.Errorf("you must provide a secret ID or a slug")
	}

	resp, err := client.DeleteProjectEnvironmentSecretWithResponse(
		ccmd.Context(),
		projectEnvironmentSecretDeleteArgs.ProjectID,
		projectEnvironmentSecretDeleteArgs.EnvironmentID,
		projectEnvironmentSecretDeleteArgs.SecretID,
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
