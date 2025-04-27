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

type projectExecutionCreateBind struct {
	ProjectID  string
	TemplateID string
	Debug      bool
	Format     string
}

var (
	projectExecutionCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project execution",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectExecutionCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectExecutionCreateArgs = projectExecutionCreateBind{}
)

func init() {
	projectExecutionCmd.AddCommand(projectExecutionCreateCmd)

	projectExecutionCreateCmd.Flags().StringVar(
		&projectExecutionCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectExecutionCreateCmd.Flags().StringVar(
		&projectExecutionCreateArgs.TemplateID,
		"template-id",
		"",
		"Template for project execution",
	)

	projectExecutionCreateCmd.Flags().BoolVar(
		&projectExecutionCreateArgs.Debug,
		"debug",
		false,
		"Debug for project execution",
	)

	projectExecutionCreateCmd.Flags().StringVar(
		&projectExecutionCreateArgs.Format,
		"format",
		tmplProjectExecutionShow,
		"Custom output format",
	)
}

func projectExecutionCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectExecutionCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectExecutionCreateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template")
	}

	body := v1.CreateProjectExecutionJSONRequestBody{}
	changed := false

	if val := projectExecutionCreateArgs.TemplateID; val != "" {
		body.TemplateID = v1.ToPtr(val)
		changed = true
	}

	if val := projectExecutionCreateArgs.Debug; val {
		body.Debug = v1.ToPtr(val)
		changed = true
	}

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
		fmt.Sprintln(projectExecutionCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectExecutionWithResponse(
		ccmd.Context(),
		projectExecutionCreateArgs.ProjectID,
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
