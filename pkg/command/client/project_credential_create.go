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

type projectCredentialCreateBind struct {
	ProjectID string
	Slug      string
	Name      string
	Kind      string
	Override  bool
	Shell     struct {
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
	projectCredentialCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project credential",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectCredentialCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectCredentialCreateArgs = projectCredentialCreateBind{}
)

func init() {
	projectCredentialCmd.AddCommand(projectCredentialCreateCmd)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Slug,
		"slug",
		"",
		"Slug for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Name,
		"name",
		"",
		"Name for project credential",
	)

	projectCredentialCreateCmd.Flags().BoolVar(
		&projectCredentialCreateArgs.Override,
		"override",
		false,
		"Override for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Kind,
		"kind",
		"",
		"Kind for project credential, can be empty, shell or login",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Shell.Username,
		"shell-username",
		"",
		"Username for shell kind for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Shell.Password,
		"shell-password",
		"",
		"Password for shell kind for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Shell.PrivateKey,
		"shell-private-key",
		"",
		"Private key for shell kind for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Login.Username,
		"login-username",
		"",
		"Username for login kind for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Login.Password,
		"login-password",
		"",
		"Password for login kind for project credential",
	)

	projectCredentialCreateCmd.Flags().StringVar(
		&projectCredentialCreateArgs.Format,
		"format",
		tmplProjectCredentialShow,
		"Custom output format",
	)
}

func projectCredentialCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectCredentialCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectCredentialCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	if projectCredentialCreateArgs.Kind == "" {
		return fmt.Errorf("you must provide a kind")
	}

	body := v1.CreateProjectCredentialJSONRequestBody{}
	changed := false

	if val := projectCredentialCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Override; val {
		body.Override = v1.ToPtr(true)
		changed = true
	}

	if val := projectCredentialCreateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Shell.Username; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.Username = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Shell.Password; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.Password = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Shell.PrivateKey; val != "" {
		if body.Shell == nil {
			body.Shell = &v1.CredentialShell{}
		}

		body.Shell.PrivateKey = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Login.Username; val != "" {
		if body.Login == nil {
			body.Login = &v1.CredentialLogin{}
		}

		body.Login.Username = v1.ToPtr(val)
		changed = true
	}

	if val := projectCredentialCreateArgs.Login.Password; val != "" {
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
		fmt.Sprintln(projectCredentialCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectCredentialWithResponse(
		ccmd.Context(),
		projectCredentialCreateArgs.ProjectID,
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
