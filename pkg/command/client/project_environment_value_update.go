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

type projectEnvironmentValueUpdateBind struct {
	ProjectID     string
	EnvironmentID string
	ValueID       string
	Kind          string
	Name          string
	Content       string
	Format        string
}

var (
	projectEnvironmentValueUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an environment value",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentValueUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentValueUpdateArgs = projectEnvironmentValueUpdateBind{}
)

func init() {
	projectEnvironmentValueCmd.AddCommand(projectEnvironmentValueUpdateCmd)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.ValueID,
		"value-id",
		"",
		"Value ID or slug",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.Kind,
		"kind",
		"",
		"Kind for environment value, can be var or env",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.Name,
		"name",
		"",
		"Name for environment value",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.Content,
		"content",
		"",
		"Content for environment value",
	)

	projectEnvironmentValueUpdateCmd.Flags().StringVar(
		&projectEnvironmentValueUpdateArgs.Format,
		"format",
		tmplProjectEnvironmentValueShow,
		"Custom output format",
	)
}

func projectEnvironmentValueUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentValueUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentValueUpdateArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide an environment ID or a slug")
	}

	if projectEnvironmentValueUpdateArgs.ValueID == "" {
		return fmt.Errorf("you must provide a value ID or a slug")
	}

	body := v1.UpdateProjectEnvironmentValueJSONRequestBody{}
	changed := false

	if val := projectEnvironmentValueUpdateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentValueUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentValueUpdateArgs.Content; val != "" {
		body.Content = v1.ToPtr(val)
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
		fmt.Sprintln(projectEnvironmentValueUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectEnvironmentValueWithResponse(
		ccmd.Context(),
		projectEnvironmentValueUpdateArgs.ProjectID,
		projectEnvironmentValueUpdateArgs.EnvironmentID,
		projectEnvironmentValueUpdateArgs.ValueID,
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
