package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectCredentialDeleteBind struct {
	ProjectID    string
	CredentialID string
}

var (
	projectCredentialDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project credential",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectCredentialDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectCredentialDeleteArgs = projectCredentialDeleteBind{}
)

func init() {
	projectCredentialCmd.AddCommand(projectCredentialDeleteCmd)

	projectCredentialDeleteCmd.Flags().StringVar(
		&projectCredentialDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectCredentialDeleteCmd.Flags().StringVar(
		&projectCredentialDeleteArgs.CredentialID,
		"credential-id",
		"",
		"Credential ID or slug",
	)
}

func projectCredentialDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectCredentialDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectCredentialDeleteArgs.CredentialID == "" {
		return fmt.Errorf("you must provide a credential ID or a slug")
	}

	resp, err := client.DeleteProjectCredentialWithResponse(
		ccmd.Context(),
		projectCredentialDeleteArgs.ProjectID,
		projectCredentialDeleteArgs.CredentialID,
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
