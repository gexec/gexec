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

type projectEnvironmentCreateBind struct {
	ProjectID string
	Slug      string
	Name      string
	Format    string
}

var (
	projectEnvironmentCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project environment",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentCreateArgs = projectEnvironmentCreateBind{}
)

func init() {
	projectEnvironmentCmd.AddCommand(projectEnvironmentCreateCmd)

	projectEnvironmentCreateCmd.Flags().StringVar(
		&projectEnvironmentCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentCreateCmd.Flags().StringVar(
		&projectEnvironmentCreateArgs.Slug,
		"slug",
		"",
		"Slug for project environment",
	)

	projectEnvironmentCreateCmd.Flags().StringVar(
		&projectEnvironmentCreateArgs.Name,
		"name",
		"",
		"Name for project environment",
	)

	projectEnvironmentCreateCmd.Flags().StringVar(
		&projectEnvironmentCreateArgs.Format,
		"format",
		tmplProjectEnvironmentShow,
		"Custom output format",
	)
}

func projectEnvironmentCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	body := v1.CreateProjectEnvironmentJSONRequestBody{}
	changed := false

	if val := projectEnvironmentCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentCreateArgs.Name; val != "" {
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
		fmt.Sprintln(projectEnvironmentCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectEnvironmentWithResponse(
		ccmd.Context(),
		projectEnvironmentCreateArgs.ProjectID,
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
