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

type projectTemplateVaultCreateBind struct {
	ProjectID  string
	TemplateID string

	// TODO: add attributes

	Format string
}

var (
	projectTemplateVaultCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an template vault",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectTemplateVaultCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectTemplateVaultCreateArgs = projectTemplateVaultCreateBind{}
)

func init() {
	projectTemplateVaultCmd.AddCommand(projectTemplateVaultCreateCmd)

	projectTemplateVaultCreateCmd.Flags().StringVar(
		&projectTemplateVaultCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectTemplateVaultCreateCmd.Flags().StringVar(
		&projectTemplateVaultCreateArgs.TemplateID,
		"template-id",
		"",
		"Template ID or slug",
	)

	// TODO: add attributes

	projectTemplateVaultCreateCmd.Flags().StringVar(
		&projectTemplateVaultCreateArgs.Format,
		"format",
		tmplProjectTemplateVaultShow,
		"Custom output format",
	)
}

func projectTemplateVaultCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectTemplateVaultCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectTemplateVaultCreateArgs.TemplateID == "" {
		return fmt.Errorf("you must provide a template ID or a slug")
	}

	body := v1.CreateProjectTemplateVaultJSONRequestBody{}
	changed := false

	// TODO: add attributes

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
		fmt.Sprintln(projectTemplateVaultCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectTemplateVaultWithResponse(
		ccmd.Context(),
		projectTemplateVaultCreateArgs.ProjectID,
		projectTemplateVaultCreateArgs.TemplateID,
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
