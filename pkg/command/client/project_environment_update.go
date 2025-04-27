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

type projectEnvironmentUpdateBind struct {
	ProjectID     string
	EnvironmentID string
	Slug          string
	Name          string
	Format        string
}

var (
	projectEnvironmentUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project environment",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentUpdateArgs = projectEnvironmentUpdateBind{}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentUpdateCmd)

	projectEnvironmentUpdateCmd.Flags().StringVar(
		&projectEnvironmentUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentUpdateCmd.Flags().StringVar(
		&projectEnvironmentUpdateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentUpdateCmd.Flags().StringVar(
		&projectEnvironmentUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project tunner",
	)

	projectEnvironmentUpdateCmd.Flags().StringVar(
		&projectEnvironmentUpdateArgs.Name,
		"name",
		"",
		"Name for project environment",
	)

	projectEnvironmentUpdateCmd.Flags().StringVar(
		&projectEnvironmentUpdateArgs.Format,
		"format",
		tmplProjectEnvironmentShow,
		"Custom output format",
	)
}

func projectEnvironmentUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentUpdateArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide a environment ID or a slug")
	}

	body := v1.UpdateProjectEnvironmentJSONRequestBody{}
	changed := false

	if val := projectEnvironmentUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentUpdateArgs.Name; val != "" {
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
		fmt.Sprintln(projectEnvironmentUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectEnvironmentWithResponse(
		ccmd.Context(),
		projectEnvironmentUpdateArgs.ProjectID,
		projectEnvironmentUpdateArgs.EnvironmentID,
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
