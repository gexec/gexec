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

type projectTemplateCreateBind struct {
	ProjectID      string
	RepositoryID   string
	InventoryID    string
	EnvironmentID  string
	Slug           string
	Name           string
	Executor       string
	Description    string
	Playbook       string
	Arguments      string
	Limit          string
	Branch         string
	AllowmOverride bool
	Format         string
}

var (
	projectTemplateCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project template",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateCreateArgs = projectTemplateCreateBind{}
)

func init() {
	projectTemplateCmd.AddCommand(projectTemplateCreateCmd)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.RepositoryID,
		"repository-id",
		"",
		"Repository for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.InventoryID,
		"inventory-id",
		"",
		"Inventory for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.EnvironmentID,
		"environment-id",
		"",
		"Environment for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Slug,
		"slug",
		"",
		"Slug for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Name,
		"name",
		"",
		"Name for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Executor,
		"executor",
		"",
		"Executor for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Description,
		"description",
		"",
		"Description for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Playbook,
		"playbook",
		"",
		"Playbook for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Arguments,
		"arguments",
		"",
		"Arguments for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Limit,
		"limit",
		"",
		"Limit for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Branch,
		"branch",
		"",
		"Branch for project template",
	)

	projectTemplateCreateCmd.Flags().BoolVar(
		&projectTemplateCreateArgs.AllowmOverride,
		"allow-override",
		false,
		"Allow override for project template",
	)

	projectTemplateCreateCmd.Flags().StringVar(
		&projectTemplateCreateArgs.Format,
		"format",
		tmplProjectTemplateShow,
		"Custom output format",
	)
}

func projectTemplateCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	if projectTemplateCreateArgs.Executor == "" {
		return fmt.Errorf("you must provide an executor")
	}

	body := v1.CreateProjectTemplateJSONRequestBody{}
	changed := false

	if val := projectTemplateCreateArgs.RepositoryID; val != "" {
		body.RepositoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.InventoryID; val != "" {
		body.InventoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.EnvironmentID; val != "" {
		body.EnvironmentID = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Executor; val != "" {
		body.Executor = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Description; val != "" {
		body.Description = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Playbook; val != "" {
		body.Playbook = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Arguments; val != "" {
		body.Arguments = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Limit; val != "" {
		body.Limit = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.Branch; val != "" {
		body.Branch = v1.ToPtr(val)
		changed = true
	}

	if val := projectTemplateCreateArgs.AllowmOverride; val {
		body.AllowOverride = v1.ToPtr(val)
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
		fmt.Sprintln(projectTemplateCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectTemplateWithResponse(
		ccmd.Context(),
		projectTemplateCreateArgs.ProjectID,
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
