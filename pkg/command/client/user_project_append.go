package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userProjectAppendBind struct {
	UserID    string
	ProjectID string
	Perm      string
}

var (
	userProjectAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append project to user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userProjectAppendAction)
		},
		Args: cobra.NoArgs,
	}

	userProjectAppendArgs = userProjectAppendBind{}
)

func init() {
	userProjectCmd.AddCommand(userProjectAppendCmd)

	userProjectAppendCmd.Flags().StringVar(
		&userProjectAppendArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userProjectAppendCmd.Flags().StringVar(
		&userProjectAppendArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	userProjectAppendCmd.Flags().StringVar(
		&userProjectAppendArgs.Perm,
		"perm",
		"",
		"Role for the project",
	)
}

func userProjectAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userProjectAppendArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if userProjectAppendArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	body := v1.AttachUserToProjectJSONRequestBody{
		Project: userProjectAppendArgs.ProjectID,
		Perm:    string(userProjectPerm(userProjectAppendArgs.Perm)),
	}

	resp, err := client.AttachUserToProjectWithResponse(
		ccmd.Context(),
		userProjectAppendArgs.UserID,
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
