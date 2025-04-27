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

type projectRepositoryCreateBind struct {
	ProjectID    string
	CredentialID string
	Slug         string
	Name         string
	URL          string
	Branch       string
	Format       string
}

var (
	projectRepositoryCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project repository",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRepositoryCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectRepositoryCreateArgs = projectRepositoryCreateBind{}
)

func init() {
	projectRepositoryCmd.AddCommand(projectRepositoryCreateCmd)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.CredentialID,
		"credential-id",
		"",
		"Credential for project repository",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.Slug,
		"slug",
		"",
		"Slug for project repository",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.Name,
		"name",
		"",
		"Name for project repository",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.URL,
		"url",
		"",
		"URL for project repository",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.Branch,
		"branch",
		"",
		"Branch for project repository",
	)

	projectRepositoryCreateCmd.Flags().StringVar(
		&projectRepositoryCreateArgs.Format,
		"format",
		tmplProjectRepositoryShow,
		"Custom output format",
	)
}

func projectRepositoryCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRepositoryCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRepositoryCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	if projectRepositoryCreateArgs.URL == "" {
		return fmt.Errorf("you must provide an url")
	}

	body := v1.CreateProjectRepositoryJSONRequestBody{}
	changed := false

	if val := projectRepositoryCreateArgs.CredentialID; val != "" {
		body.CredentialID = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryCreateArgs.URL; val != "" {
		body.URL = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryCreateArgs.Branch; val != "" {
		body.Branch = v1.ToPtr(val)
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
		fmt.Sprintln(projectRepositoryCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectRepositoryWithResponse(
		ccmd.Context(),
		projectRepositoryCreateArgs.ProjectID,
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
