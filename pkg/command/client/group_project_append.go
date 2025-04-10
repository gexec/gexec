package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupProjectAppendBind struct {
	GroupID   string
	ProjectID string
	Perm      string
}

var (
	groupProjectAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append project to group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupProjectAppendAction)
		},
		Args: cobra.NoArgs,
	}

	groupProjectAppendArgs = groupProjectAppendBind{}
)

func init() {
	groupProjectCmd.AddCommand(groupProjectAppendCmd)

	groupProjectAppendCmd.Flags().StringVar(
		&groupProjectAppendArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupProjectAppendCmd.Flags().StringVar(
		&groupProjectAppendArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	groupProjectAppendCmd.Flags().StringVar(
		&groupProjectAppendArgs.Perm,
		"perm",
		"",
		"Role for the project",
	)
}

func groupProjectAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupProjectAppendArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if groupProjectAppendArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	body := v1.AttachGroupToProjectJSONRequestBody{
		Project: groupProjectAppendArgs.ProjectID,
		Perm:    string(groupProjectPerm(groupProjectAppendArgs.Perm)),
	}

	resp, err := client.AttachGroupToProjectWithResponse(
		ccmd.Context(),
		groupProjectAppendArgs.GroupID,
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
