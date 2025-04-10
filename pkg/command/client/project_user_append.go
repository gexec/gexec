package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectUserAppendBind struct {
	ProjectID string
	UserID    string
	Perm      string
}

var (
	projectUserAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append user to project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectUserAppendAction)
		},
		Args: cobra.NoArgs,
	}

	projectUserAppendArgs = projectUserAppendBind{}
)

func init() {
	projectUserCmd.AddCommand(projectUserAppendCmd)

	projectUserAppendCmd.Flags().StringVar(
		&projectUserAppendArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectUserAppendCmd.Flags().StringVar(
		&projectUserAppendArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	projectUserAppendCmd.Flags().StringVar(
		&projectUserAppendArgs.Perm,
		"perm",
		"",
		"Role for the user",
	)
}

func projectUserAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectUserAppendArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectUserAppendArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	body := v1.AttachProjectToUserJSONRequestBody{
		User: projectUserAppendArgs.UserID,
		Perm: string(projectUserPerm(projectUserAppendArgs.Perm)),
	}

	resp, err := client.AttachProjectToUserWithResponse(
		ccmd.Context(),
		projectUserAppendArgs.ProjectID,
		body,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Fprintln(os.Stderr, v1.FromPtr(resp.JSON200.Message))
	case http.StatusUnprocessableEntity:
		return validationError(resp.JSON422)
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
