package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"text/template"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectTemplateSurveyCreateBind struct {
	ProjectID  string
	TemplateID string

	// TODO: add attributes

	Format string
}

var (
	projectTemplateSurveyCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an template survey",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateSurveyCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateSurveyCreateArgs = projectTemplateSurveyCreateBind{}
)

func init() {
	projectTemplateSurveyCmd.AddCommand(projectTemplateSurveyCreateCmd)

	projectTemplateSurveyCreateCmd.Flags().StringVar(
		&projectTemplateSurveyCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateSurveyCreateCmd.Flags().StringVar(
		&projectTemplateSurveyCreateArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	// TODO: add attributes

	projectTemplateSurveyCreateCmd.Flags().StringVar(
		&projectTemplateSurveyCreateArgs.Format,
		"format",
		tmplProjectTemplateSurveyShow,
		"Custom output format",
	)
}

func projectTemplateSurveyCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateSurveyCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateSurveyCreateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template ID or a slug")
	}

	body := v1.CreateProjectTemplateSurveyJSONRequestBody{}
	changed := false

	// TODO: add attributes

	if !changed {
		fmt.Fprintln(os.Stderr, "Nothing to create...")
		return nil
	}

	tmpl, err := template.New(
		"_",
	).Funcs(
		globalFuncMap,
	).Funcs(
		basicFuncMap,
	).Parse(
		fmt.Sprintln(projectTemplateSurveyCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectTemplateSurveyWithResponse(
		ccmd.Context(),
		projectTemplateSurveyCreateArgs.ProjectID,
		projectTemplateSurveyCreateArgs.TemplateID,
		body,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if err := tmpl.Execute(
			os.Stdout,
			resp.JSON200,
		); err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	case http.StatusUnprocessableEntity:
		return validationError(resp.JSON422)
	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return errors.New(v1.FromPtr(resp.JSON400.Message))
		}

		return errors.New(http.StatusText(http.StatusBadRequest))
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
