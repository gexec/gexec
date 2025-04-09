package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupUserAppendBind struct {
	ID   string
	User string
	Perm string
}

var (
	groupUserAppendCmd = &cobra.Command{
		Use:   "append",
		Short: "Append user to group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupUserAppendAction)
		},
		Args: cobra.NoArgs,
	}

	groupUserAppendArgs = groupUserAppendBind{}
)

func init() {
	groupUserCmd.AddCommand(groupUserAppendCmd)

	groupUserAppendCmd.Flags().StringVarP(
		&groupUserAppendArgs.ID,
		"id",
		"i",
		"",
		"Group ID or slug",
	)

	groupUserAppendCmd.Flags().StringVar(
		&groupUserAppendArgs.User,
		"user",
		"",
		"User ID or slug",
	)

	groupUserAppendCmd.Flags().StringVar(
		&groupUserAppendArgs.Perm,
		"perm",
		"",
		"Role for the user",
	)
}

func groupUserAppendAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUserAppendArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	if groupUserAppendArgs.User == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	body := v1.AttachGroupToUserJSONRequestBody{
		User: groupUserAppendArgs.User,
		Perm: string(groupUserPerm(groupUserAppendArgs.Perm)),
	}

	resp, err := client.AttachGroupToUserWithResponse(
		ccmd.Context(),
		groupUserAppendArgs.ID,
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
