package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupProjectPermitBind struct {
	GroupID   string
	ProjectID string
	Perm      string
}

var (
	groupProjectPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit project for group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupProjectPermitAction)
		},
		Args: cobra.NoArgs,
	}

	groupProjectPermitArgs = groupProjectPermitBind{}
)

func init() {
	groupProjectCmd.AddCommand(groupProjectPermitCmd)

	groupProjectPermitCmd.Flags().StringVar(
		&groupProjectPermitArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupProjectPermitCmd.Flags().StringVar(
		&groupProjectPermitArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	groupProjectPermitCmd.Flags().StringVar(
		&groupProjectPermitArgs.Perm,
		"perm",
		"",
		"Role for the project",
	)
}

func groupProjectPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupProjectPermitArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if groupProjectPermitArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if groupProjectPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like user, admin or owner")
	}

	body := v1.PermitGroupProjectJSONRequestBody{
		Project: groupProjectPermitArgs.ProjectID,
		Perm:    string(groupProjectPerm(groupProjectPermitArgs.Perm)),
	}

	resp, err := client.PermitGroupProjectWithResponse(
		ccmd.Context(),
		groupProjectPermitArgs.GroupID,
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
