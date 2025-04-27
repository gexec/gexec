package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectTemplateDeleteBind struct {
	ProjectID  string
	TemplateID string
}

var (
	projectTemplateDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project template",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateDeleteArgs = projectTemplateDeleteBind{}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateDeleteCmd)

	projectTemplateDeleteCmd.Flags().StringVar(
		&projectTemplateDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateDeleteCmd.Flags().StringVar(
		&projectTemplateDeleteArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)
}

func projectTemplateDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateDeleteArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template ID or a slug")
	}

	resp, err := client.DeleteProjectTemplateWithResponse(
		ccmd.Context(),
		projectTemplateDeleteArgs.ProjectID,
		projectTemplateDeleteArgs.TemplateID,
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
