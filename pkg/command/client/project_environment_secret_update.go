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

type projectEnvironmentSecretUpdateBind struct {
	ProjectID     string
	EnvironmentID string
	SecretID      string
	Kind          string
	Name          string
	Content       string
	Format        string
}

var (
	projectEnvironmentSecretUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an environment secret",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentSecretUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentSecretUpdateArgs = projectEnvironmentSecretUpdateBind{}
)

func init() {
	projectEnvironmentSecretCmd.AddCommand(projectEnvironmentSecretUpdateCmd)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.SecretID,
		"secret-id",
		"",
		"Secret ID or slug",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.Kind,
		"kind",
		"",
		"Kind for environment secret, can be var or env",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.Name,
		"name",
		"",
		"Name for environment secret",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.Content,
		"content",
		"",
		"Content for environment secret",
	)

	projectEnvironmentSecretUpdateCmd.Flags().StringVar(
		&projectEnvironmentSecretUpdateArgs.Format,
		"format",
		tmplProjectEnvironmentSecretShow,
		"Custom output format",
	)
}

func projectEnvironmentSecretUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentSecretUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentSecretUpdateArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide an environment ID or a slug")
	}

	if projectEnvironmentSecretUpdateArgs.SecretID == "" {
		return fmt.Errorf("you must provide a secret ID or a slug")
	}

	body := v1.UpdateProjectEnvironmentSecretJSONRequestBody{}
	changed := false

	if val := projectEnvironmentSecretUpdateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentSecretUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentSecretUpdateArgs.Content; val != "" {
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
		fmt.Sprintln(projectEnvironmentSecretUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectEnvironmentSecretWithResponse(
		ccmd.Context(),
		projectEnvironmentSecretUpdateArgs.ProjectID,
		projectEnvironmentSecretUpdateArgs.EnvironmentID,
		projectEnvironmentSecretUpdateArgs.SecretID,
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
