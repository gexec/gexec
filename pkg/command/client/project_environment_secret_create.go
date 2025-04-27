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

type projectEnvironmentSecretCreateBind struct {
	ProjectID     string
	EnvironmentID string
	Kind          string
	Name          string
	Content       string
	Format        string
}

var (
	projectEnvironmentSecretCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an environment secret",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentSecretCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentSecretCreateArgs = projectEnvironmentSecretCreateBind{}
)

func init() {
	projectEnvironmentSecretCmd.AddCommand(projectEnvironmentSecretCreateCmd)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.Kind,
		"kind",
		"",
		"Kind for environment secret, can be var or env",
	)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.Name,
		"name",
		"",
		"Name for environment secret",
	)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.Content,
		"content",
		"",
		"Content for environment secret",
	)

	projectEnvironmentSecretCreateCmd.Flags().StringVar(
		&projectEnvironmentSecretCreateArgs.Format,
		"format",
		tmplProjectEnvironmentSecretShow,
		"Custom output format",
	)
}

func projectEnvironmentSecretCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentSecretCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentSecretCreateArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide a environment ID or a slug")
	}

	body := v1.CreateProjectEnvironmentSecretJSONRequestBody{}
	changed := false

	if val := projectEnvironmentSecretCreateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentSecretCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentSecretCreateArgs.Content; val != "" {
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
		fmt.Sprintln(projectEnvironmentSecretCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectEnvironmentSecretWithResponse(
		ccmd.Context(),
		projectEnvironmentSecretCreateArgs.ProjectID,
		projectEnvironmentSecretCreateArgs.EnvironmentID,
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
