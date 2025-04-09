package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectGroupPermitBind struct {
	ID    string
	Group string
	Perm  string
}

var (
	projectGroupPermitCmd = &cobra.Command{
		Use:   "permit",
		Short: "Permit group for project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectGroupPermitAction)
		},
		Args: cobra.NoArgs,
	}

	projectGroupPermitArgs = projectGroupPermitBind{}
)

func init() {
	projectGroupCmd.AddCommand(projectGroupPermitCmd)

	projectGroupPermitCmd.Flags().StringVarP(
		&projectGroupPermitArgs.ID,
		"id",
		"i",
		"",
		"Project ID or slug",
	)

	projectGroupPermitCmd.Flags().StringVar(
		&projectGroupPermitArgs.Group,
		"group",
		"",
		"Group ID or slug",
	)

	projectGroupPermitCmd.Flags().StringVar(
		&projectGroupPermitArgs.Perm,
		"perm",
		"",
		"Role for the group",
	)
}

func projectGroupPermitAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectGroupPermitArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	if projectGroupPermitArgs.Group == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	if projectGroupPermitArgs.Perm == "" {
		return fmt.Errorf("you must provide a a permission level like project, admin or owner")
	}

	body := v1.PermitProjectGroupJSONRequestBody{
		Group: projectGroupPermitArgs.Group,
		Perm:  string(projectGroupPerm(projectGroupPermitArgs.Perm)),
	}

	resp, err := client.PermitProjectGroupWithResponse(
		ccmd.Context(),
		projectGroupPermitArgs.ID,
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
