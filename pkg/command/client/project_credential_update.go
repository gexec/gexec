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

type projectCredentialUpdateBind struct {
	ProjectID    string
	CredentialID string
	Slug         string
	Name         string
	Kind         string
	Override     bool
	NoOverride   bool
	Shell        struct {
		Username   string
		Password   string
		PrivateKey string
	}
	Login struct {
		Username string
		Password string
	}
	Format string
}

var (
	projectCredentialUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a project credential",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectCredentialUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	projectCredentialUpdateArgs = projectCredentialUpdateBind{}
)

func init() {
	projectCredentialCmd.AddCommand(projectCredentialUpdateCmd)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.CredentialID,
		"credential-id",
		"",
		"Credential ID or slug",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Slug,
		"slug",
		"",
		"Slug for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Name,
		"name",
		"",
		"Name for project credential",
	)

	projectCredentialUpdateCmd.Flags().BoolVar(
		&projectCredentialUpdateArgs.Override,
		"override",
		false,
		"Override for project credential",
	)

	projectCredentialUpdateCmd.Flags().BoolVar(
		&projectCredentialUpdateArgs.NoOverride,
		"no-override",
		false,
		"Disable override for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Kind,
		"kind",
		"",
		"Kind for project credential, can be empty, shell or login",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Shell.Username,
		"shell-username",
		"",
		"Username for shell kind for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Shell.Password,
		"shell-password",
		"",
		"Password for shell kind for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Shell.PrivateKey,
		"shell-private-key",
		"",
		"Private key for shell kind for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Login.Username,
		"login-username",
		"",
		"Username for login kind for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Login.Password,
		"login-password",
		"",
		"Password for login kind for project credential",
	)

	projectCredentialUpdateCmd.Flags().StringVar(
		&projectCredentialUpdateArgs.Format,
		"format",
		tmplProjectCredentialShow,
		"Custom output format",
	)
}

func projectCredentialUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectCredentialUpdateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectCredentialUpdateArgs.CredentialID == "" {
		return fmt.Errorf("you must provide a credential ID or a slug")
	}

	body := v1.UpdateProjectCredentialJSONRequestBody{}
	changed := false

	if val := projectCredentialUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Override; val {
		body.Override = v1.ToPtr(true)
		changed = true
	}

	if val := projectCredentialUpdateArgs.NoOverride; val {
		body.Override = v1.ToPtr(false)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Shell.Username; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.Username = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Shell.Password; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.Password = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Shell.PrivateKey; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.PrivateKey = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Login.Username; val != "" {
		if body.Login == nil {
			body.Login = &v1.CredentialLogin{}
		}

		body.Login.Username = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialUpdateArgs.Login.Password; val != "" {
		if body.Login == nil {
			body.Login = &v1.CredentialLogin{}
		}

		body.Login.Password = v1.ToPtr(val)
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
		fmt.Sprintln(projectCredentialUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateProjectCredentialWithResponse(
		ccmd.Context(),
		projectCredentialUpdateArgs.ProjectID,
		projectCredentialUpdateArgs.CredentialID,
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
