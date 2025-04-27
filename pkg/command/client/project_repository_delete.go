package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectRepositoryDeleteBind struct {
	ProjectID    string
	RepositoryID string
}

var (
	projectRepositoryDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project repository",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRepositoryDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectRepositoryDeleteArgs = projectRepositoryDeleteBind{}
)

func init() {
	projectRepositoryCmd.AddCommand(projectRepositoryDeleteCmd)

	projectRepositoryDeleteCmd.Flags().StringVar(
		&projectRepositoryDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRepositoryDeleteCmd.Flags().StringVar(
		&projectRepositoryDeleteArgs.RepositoryID,
		"repository-id",
		"",
		"Repository ID or slug",
	)
}

func projectRepositoryDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRepositoryDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRepositoryDeleteArgs.RepositoryID == "" {
		return fmt.Errorf("you must provide a repository ID or a slug")
	}

	resp, err := client.DeleteProjectRepositoryWithResponse(
		ccmd.Context(),
		projectRepositoryDeleteArgs.ProjectID,
		projectRepositoryDeleteArgs.RepositoryID,
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
