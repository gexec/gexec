package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupUserRemoveBind struct {
	GroupID string
	UserID  string
}

var (
	groupUserRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove user from group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupUserRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	groupUserRemoveArgs = groupUserRemoveBind{}
)

func init() {
	groupUserCmd.AddCommand(groupUserRemoveCmd)

	groupUserRemoveCmd.Flags().StringVar(
		&groupUserRemoveArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupUserRemoveCmd.Flags().StringVar(
		&groupUserRemoveArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)
}

func groupUserRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUserRemoveArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if groupUserRemoveArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	resp, err := client.DeleteGroupFromUserWithResponse(
		ccmd.Context(),
		groupUserRemoveArgs.GroupID,
		v1.DeleteGroupFromUserJSONRequestBody{
			User: groupUserRemoveArgs.UserID,
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
