package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectEnvironmentDeleteBind struct {
	ProjectID     string
	EnvironmentID string
}

var (
	projectEnvironmentDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project environment",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentDeleteArgs = projectEnvironmentDeleteBind{}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentDeleteCmd)

	projectEnvironmentDeleteCmd.Flags().StringVar(
		&projectEnvironmentDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentDeleteCmd.Flags().StringVar(
		&projectEnvironmentDeleteArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)
}

func projectEnvironmentDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentDeleteArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide a environment ID or a slug")
	}

	resp, err := client.DeleteProjectEnvironmentWithResponse(
		ccmd.Context(),
		projectEnvironmentDeleteArgs.ProjectID,
		projectEnvironmentDeleteArgs.EnvironmentID,
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
