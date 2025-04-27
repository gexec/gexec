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

type projectScheduleCreateBind struct {
	ProjectID  string
	TemplateID string
	Slug       string
	Name       string
	Cron       string
	Active     bool
	Format     string
}

var (
	projectScheduleCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project schedule",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectScheduleCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectScheduleCreateArgs = projectScheduleCreateBind{}
)

func init() {
	projectScheduleCmd.AddCommand(projectScheduleCreateCmd)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.TemplateID,
		"template-id",
		"",
		"Template for project schedule",
	)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.Slug,
		"slug",
		"",
		"Slug for project schedule",
	)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.Name,
		"name",
		"",
		"Name for project schedule",
	)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.Cron,
		"cron",
		"",
		"Cron for project schedule",
	)

	projectScheduleCreateCmd.Flags().BoolVar(
		&projectScheduleCreateArgs.Active,
		"active",
		false,
		"Active for project schedule",
	)

	projectScheduleCreateCmd.Flags().StringVar(
		&projectScheduleCreateArgs.Format,
		"format",
		tmplProjectScheduleShow,
		"Custom output format",
	)
}

func projectScheduleCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectScheduleCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectScheduleCreateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template")
	}

	if projectScheduleCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	if projectScheduleCreateArgs.Cron == "" {
		return fmt.Errorf("you must provide a cron")
	}

	body := v1.CreateProjectScheduleJSONRequestBody{}
	changed := false

	if val := projectScheduleCreateArgs.TemplateID; val != "" {
		body.TemplateID = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleCreateArgs.Cron; val != "" {
		body.Cron = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleCreateArgs.Active; val {
		body.Active = v1.ToPtr(val)
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
		fmt.Sprintln(projectScheduleCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectScheduleWithResponse(
		ccmd.Context(),
		projectScheduleCreateArgs.ProjectID,
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
