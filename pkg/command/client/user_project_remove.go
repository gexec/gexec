package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userProjectRemoveBind struct {
	UserID    string
	ProjectID string
}

var (
	userProjectRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove project from user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userProjectRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	userProjectRemoveArgs = userProjectRemoveBind{}
)

func init() {
	userProjectCmd.AddCommand(userProjectRemoveCmd)

	userProjectRemoveCmd.Flags().StringVar(
		&userProjectRemoveArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userProjectRemoveCmd.Flags().StringVar(
		&userProjectRemoveArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)
}

func userProjectRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userProjectRemoveArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if userProjectRemoveArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.DeleteUserFromProjectWithResponse(
		ccmd.Context(),
		userProjectRemoveArgs.UserID,
		v1.DeleteUserFromProjectJSONRequestBody{
			Project: userProjectRemoveArgs.ProjectID,
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
