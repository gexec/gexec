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

type projectScheduleUpdateBind struct {
	ProjectID  string
	ScheduleID string
	TemplateID string
	Slug       string
	Name       string
	Cron       string
	Active     bool
	NoActive   bool
	Format     string
}

var (
	projectScheduleUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project schedule",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectScheduleUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectScheduleUpdateArgs = projectScheduleUpdateBind{}
)

func init() {
	projectScheduleCmd.AddCommand(projectScheduleUpdateCmd)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.ScheduleID,
		"schedule-id",
		"",
		"Schedule ID or slug",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.TemplateID,
		"template-id",
		"",
		"Template for project schedule",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project schedule",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.Name,
		"name",
		"",
		"Name for project schedule",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.Cron,
		"cron",
		"",
		"Cron for project schedule",
	)

	projectScheduleUpdateCmd.Flags().BoolVar(
		&projectScheduleUpdateArgs.Active,
		"active",
		false,
		"Active for project schedule",
	)

	projectScheduleUpdateCmd.Flags().BoolVar(
		&projectScheduleUpdateArgs.NoActive,
		"no-active",
		false,
		"Disable for project credential",
	)

	projectScheduleUpdateCmd.Flags().StringVar(
		&projectScheduleUpdateArgs.Format,
		"format",
		tmplProjectScheduleShow,
		"Custom output format",
	)
}

func projectScheduleUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectScheduleUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectScheduleUpdateArgs.ScheduleID == "" {
		return fmt.Errorf("you must provide a schedule ID or a slug")
	}

	body := v1.UpdateProjectScheduleJSONRequestBody{}
	changed := false

	if val := projectScheduleUpdateArgs.TemplateID; val != "" {
		body.TemplateID = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleUpdateArgs.Cron; val != "" {
		body.Cron = v1.ToPtr(val)
		changed = true
	}

	if val := projectScheduleUpdateArgs.Active; val {
		body.Active = v1.ToPtr(true)
		changed = true
	}

	if val := projectScheduleUpdateArgs.NoActive; val {
		body.Active = v1.ToPtr(false)
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
		fmt.Sprintln(projectScheduleUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectScheduleWithResponse(
		ccmd.Context(),
		projectScheduleUpdateArgs.ProjectID,
		projectScheduleUpdateArgs.ScheduleID,
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
