package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectGroupRemoveBind struct {
	ProjectID string
	GroupID   string
}

var (
	projectGroupRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove group from project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectGroupRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	projectGroupRemoveArgs = projectGroupRemoveBind{}
)

func init() {
	projectGroupCmd.AddCommand(projectGroupRemoveCmd)

	projectGroupRemoveCmd.Flags().StringVar(
		&projectGroupRemoveArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectGroupRemoveCmd.Flags().StringVar(
		&projectGroupRemoveArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)
}

func projectGroupRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectGroupRemoveArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectGroupRemoveArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	resp, err := client.DeleteProjectFromGroupWithResponse(
		ccmd.Context(),
		projectGroupRemoveArgs.ProjectID,
		v1.DeleteProjectFromGroupJSONRequestBody{
			Group: projectGroupRemoveArgs.GroupID,
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
