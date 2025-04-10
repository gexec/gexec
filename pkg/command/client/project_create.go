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

type projectCreateBind struct {
	Slug   string
	Name   string
	Demo   bool
	Format string
}

var (
	projectCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectCreateArgs = projectCreateBind{}
)

func init() {
	projectCmd.AddCommand(projectCreateCmd)

	projectCreateCmd.Flags().StringVar(
		&projectCreateArgs.Slug,
		"slug",
		"",
		"Slug for project",
	)

	projectCreateCmd.Flags().StringVar(
		&projectCreateArgs.Name,
		"name",
		"",
		"Name for project",
	)

	projectCreateCmd.Flags().BoolVar(
		&projectCreateArgs.Demo,
		"demo",
		false,
		"Create demo resources",
	)

	projectCreateCmd.Flags().StringVar(
		&projectCreateArgs.Format,
		"format",
		tmplProjectShow,
		"Custom output format",
	)
}

func projectCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	body := v1.CreateProjectJSONRequestBody{}
	changed := false

	if val := projectCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectCreateArgs.Demo; val {
		body.Demo = v1.ToPtr(val)
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
		fmt.Sprintln(projectCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectWithResponse(
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
