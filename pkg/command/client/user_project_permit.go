package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userProjectPermitBind struct {
	UserID    string
	ProjectID string
	Perm      string
}

var (
	userProjectPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit project for user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userProjectPermitAction)
		},
		Args: cobra.NoArgs,
	}

	userProjectPermitArgs = userProjectPermitBind{}
)

func init() {
	userProjectCmd.AddCommand(userProjectPermitCmd)

	userProjectPermitCmd.Flags().StringVar(
		&userProjectPermitArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userProjectPermitCmd.Flags().StringVar(
		&userProjectPermitArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	userProjectPermitCmd.Flags().StringVar(
		&userProjectPermitArgs.Perm,
		"perm",
		"",
		"Role for the project",
	)
}

func userProjectPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userProjectPermitArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if userProjectPermitArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if userProjectPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like user, admin or owner")
	}

	body := v1.PermitUserProjectJSONRequestBody{
		Project: userProjectPermitArgs.ProjectID,
		Perm:    string(userProjectPerm(userProjectPermitArgs.Perm)),
	}

	resp, err := client.PermitUserProjectWithResponse(
		ccmd.Context(),
		userProjectPermitArgs.UserID,
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
