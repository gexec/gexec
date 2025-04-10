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

type projectRunnerUpdateBind struct {
	ProjectID string
	RunnerID  string
	Slug      string
	Name      string
	Format    string
}

var (
	projectRunnerUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRunnerUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectRunnerUpdateArgs = projectRunnerUpdateBind{}
)

func init() {
	projectRunnerCmd.AddCommand(projectRunnerUpdateCmd)

	projectRunnerUpdateCmd.Flags().StringVar(
		&projectRunnerUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRunnerUpdateCmd.Flags().StringVar(
		&projectRunnerUpdateArgs.RunnerID,
		"runner-id",
		"",
		"Runner ID or slug",
	)

	projectRunnerUpdateCmd.Flags().StringVar(
		&projectRunnerUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project tunner",
	)

	projectRunnerUpdateCmd.Flags().StringVar(
		&projectRunnerUpdateArgs.Name,
		"name",
		"",
		"Name for project runner",
	)

	projectRunnerUpdateCmd.Flags().StringVar(
		&projectRunnerUpdateArgs.Format,
		"format",
		tmplProjectRunnerShow,
		"Custom output format",
	)
}

func projectRunnerUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRunnerUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRunnerUpdateArgs.RunnerID == "" {
		return fmt.Errorf("you must provide a runner ID or a slug")
	}

	body := v1.UpdateProjectRunnerJSONRequestBody{}
	changed := false

	if val := projectRunnerUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectRunnerUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
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
		fmt.Sprintln(projectRunnerUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectRunnerWithResponse(
		ccmd.Context(),
		projectRunnerUpdateArgs.ProjectID,
		projectRunnerUpdateArgs.RunnerID,
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
