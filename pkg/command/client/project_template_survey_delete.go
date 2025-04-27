package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectTemplateSurveyDeleteBind struct {
	ProjectID  string
	TemplateID string
	SurveyID   string
}

var (
	projectTemplateSurveyDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an template survey",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateSurveyDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateSurveyDeleteArgs = projectTemplateSurveyDeleteBind{}
)

func init() {
	projectTemplateSurveyCmd.AddCommand(projectTemplateSurveyDeleteCmd)

	projectTemplateSurveyDeleteCmd.Flags().StringVar(
		&projectTemplateSurveyDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateSurveyDeleteCmd.Flags().StringVar(
		&projectTemplateSurveyDeleteArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateSurveyDeleteCmd.Flags().StringVar(
		&projectTemplateSurveyDeleteArgs.SurveyID,
		"survey-id",
		"",
		"Survey ID or slug",
	)
}

func projectTemplateSurveyDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateSurveyDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateSurveyDeleteArgs.TemplateID == "" {
		return fmt.Errorf("you must provide an template ID or a slug")
	}

	if projectTemplateSurveyDeleteArgs.SurveyID == "" {
		return fmt.Errorf("you must provide a survey ID or a slug")
	}

	resp, err := client.DeleteProjectTemplateSurveyWithResponse(
		ccmd.Context(),
		projectTemplateSurveyDeleteArgs.ProjectID,
		projectTemplateSurveyDeleteArgs.TemplateID,
		projectTemplateSurveyDeleteArgs.SurveyID,
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
