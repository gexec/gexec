package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userGroupPermitBind struct {
	UserID  string
	GroupID string
	Perm    string
}

var (
	userGroupPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit group for user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userGroupPermitAction)
		},
		Args: cobra.NoArgs,
	}

	userGroupPermitArgs = userGroupPermitBind{}
)

func init() {
	userGroupCmd.AddCommand(userGroupPermitCmd)

	userGroupPermitCmd.Flags().StringVar(
		&userGroupPermitArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userGroupPermitCmd.Flags().StringVar(
		&userGroupPermitArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	userGroupPermitCmd.Flags().StringVar(
		&userGroupPermitArgs.Perm,
		"perm",
		"",
		"Role for the group",
	)
}

func userGroupPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userGroupPermitArgs.UserID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	if userGroupPermitArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if userGroupPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like user, admin or owner")
	}

	body := v1.PermitUserGroupJSONRequestBody{
		Group: userGroupPermitArgs.GroupID,
		Perm:  string(userGroupPerm(userGroupPermitArgs.Perm)),
	}

	resp, err := client.PermitUserGroupWithResponse(
		ccmd.Context(),
		userGroupPermitArgs.UserID,
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
