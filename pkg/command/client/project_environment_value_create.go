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

type projectEnvironmentValueCreateBind struct {
	ProjectID     string
	EnvironmentID string
	Kind          string
	Name          string
	Content       string
	Format        string
}

var (
	projectEnvironmentValueCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an environment value",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectEnvironmentValueCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectEnvironmentValueCreateArgs = projectEnvironmentValueCreateBind{}
)

func init() {
	projectEnvironmentValueCmd.AddCommand(projectEnvironmentValueCreateCmd)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment ID or slug",
	)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.Kind,
		"kind",
		"",
		"Kind for environment value, can be var or env",
	)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.Name,
		"name",
		"",
		"Name for environment value",
	)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.Content,
		"content",
		"",
		"Content for environment value",
	)

	projectEnvironmentValueCreateCmd.Flags().StringVar(
		&projectEnvironmentValueCreateArgs.Format,
		"format",
		tmplProjectEnvironmentValueShow,
		"Custom output format",
	)
}

func projectEnvironmentValueCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectEnvironmentValueCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectEnvironmentValueCreateArgs.EnvironmentID == "" {
		return fmt.Errorf("you must provide a environment ID or a slug")
	}

	body := v1.CreateProjectEnvironmentValueJSONRequestBody{}
	changed := false

	if val := projectEnvironmentValueCreateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentValueCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectEnvironmentValueCreateArgs.Content; val != "" {
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
		fmt.Sprintln(projectEnvironmentValueCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectEnvironmentValueWithResponse(
		ccmd.Context(),
		projectEnvironmentValueCreateArgs.ProjectID,
		projectEnvironmentValueCreateArgs.EnvironmentID,
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
