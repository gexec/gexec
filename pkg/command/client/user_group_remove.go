package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userGroupRemoveBind struct {
	ID    string
	Group string
}

var (
	userGroupRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove group from user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userGroupRemoveAction)
		},
		Args: cobra.NoArgs,
	}

	userGroupRemoveArgs = userGroupRemoveBind{}
)

func init() {
	userGroupCmd.AddCommand(userGroupRemoveCmd)

	userGroupRemoveCmd.Flags().StringVarP(
		&userGroupRemoveArgs.ID,
		"id",
		"i",
		"",
		"User ID or slug",
	)

	userGroupRemoveCmd.Flags().StringVar(
		&userGroupRemoveArgs.Group,
		"group",
		"",
		"Group ID or slug",
	)
}

func userGroupRemoveAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userGroupRemoveArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	if userGroupRemoveArgs.Group == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	resp, err := client.DeleteUserFromGroupWithResponse(
		ccmd.Context(),
		userGroupRemoveArgs.ID,
		v1.DeleteUserFromGroupJSONRequestBody{
			Group: userGroupRemoveArgs.Group,
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
