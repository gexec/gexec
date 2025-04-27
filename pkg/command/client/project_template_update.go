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

type projectTemplateUpdateBind struct {
	ProjectID        string
	TemplateID       string
	RepositoryID     string
	InventoryID      string
	EnvironmentID    string
	Slug             string
	Name             string
	Description      string
	Path             string
	Arguments        string
	Limit            string
	Branch           string
	AllowmOverride   bool
	NoAllowmOverride bool
	Format           string
}

var (
	projectTemplateUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project template",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateUpdateArgs = projectTemplateUpdateBind{}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateUpdateCmd)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.RepositoryID,
		"repository-id",
		"",
		"Repository for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.InventoryID,
		"inventory-id",
		"",
		"Inventory for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Name,
		"name",
		"",
		"Name for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Description,
		"description",
		"",
		"Description for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Path,
		"path",
		"",
		"Path for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Arguments,
		"arguments",
		"",
		"Arguments for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Limit,
		"limit",
		"",
		"Limit for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Branch,
		"branch",
		"",
		"Branch for project template",
	)

	projectTemplateUpdateCmd.Flags().BoolVar(
		&projectTemplateUpdateArgs.AllowmOverride,
		"allow-override",
		false,
		"Allow override for project template",
	)

	projectTemplateUpdateCmd.Flags().BoolVar(
		&projectTemplateUpdateArgs.NoAllowmOverride,
		"no-allow-override",
		false,
		"No allow override for project template",
	)

	projectTemplateUpdateCmd.Flags().StringVar(
		&projectTemplateUpdateArgs.Format,
		"format",
		tmplProjectTemplateShow,
		"Custom output format",
	)
}

func projectTemplateUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateUpdateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template ID or a slug")
	}

	body := v1.UpdateProjectTemplateJSONRequestBody{}
	changed := false

	if val := projectTemplateUpdateArgs.RepositoryID; val != "" {
		body.RepositoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.InventoryID; val != "" {
		body.InventoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.EnvironmentID; val != "" {
		body.EnvironmentID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Description; val != "" {
		body.Description = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Path; val != "" {
		body.Path = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Arguments; val != "" {
		body.Arguments = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Limit; val != "" {
		body.Limit = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.Branch; val != "" {
		body.Branch = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateUpdateArgs.AllowmOverride; val {
		body.AllowOverride = v1.ToPtr(true)
		changed = true
	}

	if val := projectTemplateUpdateArgs.NoAllowmOverride; val {
		body.AllowOverride = v1.ToPtr(false)
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
		fmt.Sprintln(projectTemplateUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectTemplateWithResponse(
		ccmd.Context(),
		projectTemplateUpdateArgs.ProjectID,
		projectTemplateUpdateArgs.TemplateID,
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
