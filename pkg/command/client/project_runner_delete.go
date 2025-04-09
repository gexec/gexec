package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectRunnerDeleteBind struct {
	Project string
	ID      string
}

var (
	projectRunnerDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRunnerDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectRunnerDeleteArgs = projectRunnerDeleteBind{}
)

func init() {
	projectRunnerCmd.AddCommand(projectRunnerDeleteCmd)

	projectRunnerDeleteCmd.Flags().StringVarP(
		&projectRunnerDeleteArgs.Project,
		"project",
		"p",
		"",
		"Project ID or slug",
	)

	projectRunnerDeleteCmd.Flags().StringVarP(
		&projectRunnerDeleteArgs.ID,
		"id",
		"i",
		"",
		"Runner ID or slug",
	)
}

func projectRunnerDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRunnerDeleteArgs.Project == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRunnerDeleteArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	resp, err := client.DeleteProjectRunnerWithResponse(
		ccmd.Context(),
		projectRunnerDeleteArgs.Project,
		projectRunnerDeleteArgs.ID,
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
