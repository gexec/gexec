package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupProjectRemoveBind struct {
	GroupID   string
	ProjectID string
}

var (
	groupProjectRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove project from group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupProjectRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	groupProjectRemoveArgs = groupProjectRemoveBind{}
)

func init() {
	groupProjectCmd.AddCommand(groupProjectRemoveCmd)

	groupProjectRemoveCmd.Flags().StringVar(
		&groupProjectRemoveArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupProjectRemoveCmd.Flags().StringVar(
		&groupProjectRemoveArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)
}

func groupProjectRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupProjectRemoveArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if groupProjectRemoveArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.DeleteGroupFromProjectWithResponse(
		ccmd.Context(),
		groupProjectRemoveArgs.GroupID,
		v1.DeleteGroupFromProjectJSONRequestBody{
			Project: groupProjectRemoveArgs.ProjectID,
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
