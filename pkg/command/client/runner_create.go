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

type runnerCreateBind struct {
	Project string
	Slug    string
	Name    string
	Format  string
}

var (
	runnerCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a runner",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, runnerCreateAction)
		},
		Args: cobra.NoArgs,
	}

	runnerCreateArgs = runnerCreateBind{}
)

func init() {
	runnerCmd.AddCommand(runnerCreateCmd)

	runnerCreateCmd.Flags().StringVar(
		&runnerCreateArgs.Project,
		"project",
		"",
		"Project for runner",
	)

	runnerCreateCmd.Flags().StringVar(
		&runnerCreateArgs.Slug,
		"slug",
		"",
		"Slug for runner",
	)

	runnerCreateCmd.Flags().StringVar(
		&runnerCreateArgs.Name,
		"name",
		"",
		"Name for runner",
	)

	runnerCreateCmd.Flags().StringVar(
		&runnerCreateArgs.Format,
		"format",
		tmplRunnerShow,
		"Custom output format",
	)
}

func runnerCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if runnerCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	body := v1.CreateGlobalRunnerJSONRequestBody{}
	changed := false

	if val := runnerCreateArgs.Project; val != "" {
		body.ProjectID = v1.ToPtr(val)
		changed = true
	}

	if val := runnerCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := runnerCreateArgs.Name; val != "" {
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
		fmt.Sprintln(runnerCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateGlobalRunnerWithResponse(
		ccmd.Context(),
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
