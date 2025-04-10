package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectUserRemoveBind struct {
	ProjectID string
	UserID    string
}

var (
	projectUserRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove user from project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectUserRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	projectUserRemoveArgs = projectUserRemoveBind{}
)

func init() {
	projectUserCmd.AddCommand(projectUserRemoveCmd)

	projectUserRemoveCmd.Flags().StringVar(
		&projectUserRemoveArgs.ProjectID,
		"id",
		"",
		"Project ID or slug",
	)

	projectUserRemoveCmd.Flags().StringVar(
		&projectUserRemoveArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)
}

func projectUserRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectUserRemoveArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectUserRemoveArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	resp, err := client.DeleteProjectFromUserWithResponse(
		ccmd.Context(),
		projectUserRemoveArgs.ProjectID,
		v1.DeleteProjectFromUserJSONRequestBody{
			User: projectUserRemoveArgs.UserID,
		},
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Fprintln(os.Stderr, v1.FromPtr(resp.JSON200.Message))
	case http.StatusPreconditionFailed:
		return errors.New(v1.FromPtr(resp.JSON412.Message))
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
