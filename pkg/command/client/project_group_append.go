package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectGroupAppendBind struct {
	ProjectID string
	GroupID   string
	Perm      string
}

var (
	projectGroupAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append group to project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectGroupAppendAction)
		},
		Args: cobra.NoArgs,
	}

	projectGroupAppendArgs = projectGroupAppendBind{}
)

func init() {
	projectGroupCmd.AddCommand(projectGroupAppendCmd)

	projectGroupAppendCmd.Flags().StringVar(
		&projectGroupAppendArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectGroupAppendCmd.Flags().StringVar(
		&projectGroupAppendArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	projectGroupAppendCmd.Flags().StringVar(
		&projectGroupAppendArgs.Perm,
		"perm",
		"",
		"Role for the group",
	)
}

func projectGroupAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectGroupAppendArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectGroupAppendArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	body := v1.AttachProjectToGroupJSONRequestBody{
		Group: projectGroupAppendArgs.GroupID,
		Perm:  string(projectGroupPerm(projectGroupPermitArgs.Perm)),
	}

	resp, err := client.AttachProjectToGroupWithResponse(
		ccmd.Context(),
		projectGroupAppendArgs.ProjectID,
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
