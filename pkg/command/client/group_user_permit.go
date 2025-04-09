package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type groupUserPermitBind struct {
	ID   string
	User string
	Perm string
}

var (
	groupUserPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit user for group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupUserPermitAction)
		},
		Args: cobra.NoArgs,
	}

	groupUserPermitArgs = groupUserPermitBind{}
)

func init() {
	groupUserCmd.AddCommand(groupUserPermitCmd)

	groupUserPermitCmd.Flags().StringVarP(
		&groupUserPermitArgs.ID,
		"id",
		"i",
		"",
		"Group ID or slug",
	)

	groupUserPermitCmd.Flags().StringVar(
		&groupUserPermitArgs.User,
		"user",
		"",
		"User ID or slug",
	)

	groupUserPermitCmd.Flags().StringVar(
		&groupUserPermitArgs.Perm,
		"perm",
		"",
		"Role for the user",
	)
}

func groupUserPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUserPermitArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	if groupUserPermitArgs.User == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	if groupUserPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like user, admin or owner")
	}

	body := v1.PermitGroupUserJSONRequestBody{
		User: groupUserPermitArgs.User,
		Perm: string(groupUserPerm(groupUserPermitArgs.Perm)),
	}

	resp, err := client.PermitGroupUserWithResponse(
		ccmd.Context(),
		groupUserPermitArgs.ID,
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
