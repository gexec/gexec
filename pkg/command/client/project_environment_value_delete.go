package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectEnvironmentValueDeleteBind struct {
	ProjectID     string
	EnvironmentID string
	ValueID       string
}

var (
	projectEnvironmentValueDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an environment value",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentValueDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentValueDeleteArgs = projectEnvironmentValueDeleteBind{}
)

func init() {
	projectEnvironmentValueCmd.AddCommand(projectEnvironmentValueDeleteCmd)

	projectEnvironmentValueDeleteCmd.Flags().StringVar(
		&projectEnvironmentValueDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentValueDeleteCmd.Flags().StringVar(
		&projectEnvironmentValueDeleteArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentValueDeleteCmd.Flags().StringVar(
		&projectEnvironmentValueDeleteArgs.ValueID,
		"value-id",
		"",
		"Value ID or slug",
	)
}

func projectEnvironmentValueDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentValueDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentValueDeleteArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide an environment ID or a slug")
	}

	if projectEnvironmentValueDeleteArgs.ValueID == "" {
		return fmt.Errorf("you must provide a value ID or a slug")
	}

	resp, err := client.DeleteProjectEnvironmentValueWithResponse(
		ccmd.Context(),
		projectEnvironmentValueDeleteArgs.ProjectID,
		projectEnvironmentValueDeleteArgs.EnvironmentID,
		projectEnvironmentValueDeleteArgs.ValueID,
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
