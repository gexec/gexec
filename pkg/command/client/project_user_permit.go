package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectUserPermitBind struct {
	ProjectID string
	UserID    string
	Perm      string
}

var (
	projectUserPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit user for project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectUserPermitAction)
		},
		Args: cobra.NoArgs,
	}

	projectUserPermitArgs = projectUserPermitBind{}
)

func init() {
	projectUserCmd.AddCommand(projectUserPermitCmd)

	projectUserPermitCmd.Flags().StringVar(
		&projectUserPermitArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectUserPermitCmd.Flags().StringVar(
		&projectUserPermitArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	projectUserPermitCmd.Flags().StringVar(
		&projectUserPermitArgs.Perm,
		"perm",
		"",
		"Role for the user",
	)
}

func projectUserPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectUserPermitArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectUserPermitArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if projectUserPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like user, admin or owner")
	}

	body := v1.PermitProjectUserJSONRequestBody{
		User: projectUserPermitArgs.UserID,
		Perm: string(projectUserPerm(projectUserPermitArgs.Perm)),
	}

	resp, err := client.PermitProjectUserWithResponse(
		ccmd.Context(),
		projectUserPermitArgs.ProjectID,
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
