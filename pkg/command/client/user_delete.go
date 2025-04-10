package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type userDeleteBind struct {
	UserID string
}

var (
	userDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	userDeleteArgs = userDeleteBind{}
)

func init() {
	userCmd.AddCommand(userDeleteCmd)

	userDeleteCmd.Flags().StringVar(
		&userDeleteArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)
}

func userDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userDeleteArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	resp, err := client.DeleteUserWithResponse(
		ccmd.Context(),
		userDeleteArgs.UserID,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Fprintln(os.Stderr, "Successfully deleted")
	case http.StatusForbidden:
		if resp.JSON403 != nil {
			return errors.New(v1.FromPtr(resp.JSON403.Message))
		}

		return errors.New(http.StatusText(http.StatusForbidden))
	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return errors.New(v1.FromPtr(resp.JSON400.Message))
		}

		return errors.New(http.StatusText(http.StatusBadRequest))
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
