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

type projectRepositoryUpdateBind struct {
	ProjectID      string
	RepositoryID   string
	CredentialID   string
	NoCredentialID bool
	Slug           string
	Name           string
	URL            string
	Branch         string
	Format         string
}

var (
	projectRepositoryUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project repository",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRepositoryUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectRepositoryUpdateArgs = projectRepositoryUpdateBind{}
)

func init() {
	projectRepositoryCmd.AddCommand(projectRepositoryUpdateCmd)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.RepositoryID,
		"repository-id",
		"",
		"Repository ID or slug",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.CredentialID,
		"credential-id",
		"",
		"Credential for project repository",
	)

	projectRepositoryUpdateCmd.Flags().BoolVar(
		&projectRepositoryUpdateArgs.NoCredentialID,
		"no-credential-id",
		false,
		"Remove credential for project repository",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project repository",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.Name,
		"name",
		"",
		"Name for project repository",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.URL,
		"url",
		"",
		"URL for project repository",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.Branch,
		"branch",
		"",
		"Branch for project repository",
	)

	projectRepositoryUpdateCmd.Flags().StringVar(
		&projectRepositoryUpdateArgs.Format,
		"format",
		tmplProjectRepositoryShow,
		"Custom output format",
	)
}

func projectRepositoryUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRepositoryUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectRepositoryUpdateArgs.RepositoryID == "" {
		return fmt.Errorf("you must provide a repository ID or a slug")
	}

	body := v1.UpdateProjectRepositoryJSONRequestBody{}
	changed := false

	if val := projectRepositoryUpdateArgs.CredentialID; val != "" {
		body.CredentialID = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryUpdateArgs.NoCredentialID; val {
		body.CredentialID = v1.ToPtr("")
		changed = true
	}

	if val := projectRepositoryUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryUpdateArgs.URL; val != "" {
		body.URL = v1.ToPtr(val)
		changed = true
	}

	if val := projectRepositoryUpdateArgs.Branch; val != "" {
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
		fmt.Sprintln(projectRepositoryUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectRepositoryWithResponse(
		ccmd.Context(),
		projectRepositoryUpdateArgs.ProjectID,
		projectRepositoryUpdateArgs.RepositoryID,
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
