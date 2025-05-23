package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userGroupAppendBind struct {
	UserID  string
	GroupID string
	Perm    string
}

var (
	userGroupAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append group to user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userGroupAppendAction)
		},
		Args: cobra.NoArgs,
	}

	userGroupAppendArgs = userGroupAppendBind{}
)

func init() {
	userGroupCmd.AddCommand(userGroupAppendCmd)

	userGroupAppendCmd.Flags().StringVar(
		&userGroupAppendArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userGroupAppendCmd.Flags().StringVar(
		&userGroupAppendArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	userGroupAppendCmd.Flags().StringVar(
		&userGroupAppendArgs.Perm,
		"perm",
		"",
		"Role for the group",
	)
}

func userGroupAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userGroupAppendArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if userGroupAppendArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	body := v1.AttachUserToGroupJSONRequestBody{
		Group: userGroupAppendArgs.GroupID,
		Perm:  string(userGroupPerm(userGroupAppendArgs.Perm)),
	}

	resp, err := client.AttachUserToGroupWithResponse(
		ccmd.Context(),
		userGroupAppendArgs.UserID,
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
