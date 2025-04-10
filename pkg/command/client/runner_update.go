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

type runnerUpdateBind struct {
	RunnerID string
	Project  string
	Slug     string
	Name     string
	Format   string
}

var (
	runnerUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, runnerUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	runnerUpdateArgs = runnerUpdateBind{}
)

func init() {
	runnerCmd.AddCommand(runnerUpdateCmd)

	runnerUpdateCmd.Flags().StringVar(
		&runnerUpdateArgs.RunnerID,
		"id",
		"",
		"Runner ID or slug",
	)

	runnerUpdateCmd.Flags().StringVar(
		&runnerUpdateArgs.Project,
		"project",
		"",
		"Project for runner",
	)

	runnerUpdateCmd.Flags().StringVar(
		&runnerUpdateArgs.Slug,
		"slug",
		"",
		"Slug for tunner",
	)

	runnerUpdateCmd.Flags().StringVar(
		&runnerUpdateArgs.Name,
		"name",
		"",
		"Name for runner",
	)

	runnerUpdateCmd.Flags().StringVar(
		&runnerUpdateArgs.Format,
		"format",
		tmplRunnerShow,
		"Custom output format",
	)
}

func runnerUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if runnerUpdateArgs.RunnerID == "" {
		return fmt.Errorf("you must provide a runner ID or a slug")
	}

	body := v1.UpdateGlobalRunnerJSONRequestBody{}
	changed := false

	if val := runnerUpdateArgs.Project; val != "" {
		body.ProjectID = v1.ToPtr(val)
		changed = true
	}

	if val := runnerUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := runnerUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
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
		fmt.Sprintln(runnerUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateGlobalRunnerWithResponse(
		ccmd.Context(),
		runnerUpdateArgs.RunnerID,
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
