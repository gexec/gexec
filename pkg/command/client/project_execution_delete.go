package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectExecutionDeleteBind struct {
	ProjectID   string
	ExecutionID string
}

var (
	projectExecutionDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project execution",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectExecutionDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectExecutionDeleteArgs = projectExecutionDeleteBind{}
)

func init() {
	projectExecutionCmd.AddCommand(projectExecutionDeleteCmd)

	projectExecutionDeleteCmd.Flags().StringVar(
		&projectExecutionDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectExecutionDeleteCmd.Flags().StringVar(
		&projectExecutionDeleteArgs.ExecutionID,
		"execution-id",
		"",
		"Execution ID or slug",
	)
}

func projectExecutionDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectExecutionDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectExecutionDeleteArgs.ExecutionID == "" {
		return fmt.Errorf("you must provide a execution ID or a slug")
	}

	resp, err := client.DeleteProjectExecutionWithResponse(
		ccmd.Context(),
		projectExecutionDeleteArgs.ProjectID,
		projectExecutionDeleteArgs.ExecutionID,
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
