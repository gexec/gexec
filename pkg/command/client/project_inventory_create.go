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

type projectInventoryCreateBind struct {
	ProjectID    string
	RepositoryID string
	CredentialID string
	BecomeID     string
	Slug         string
	Name         string
	Kind         string
	Content      string
	Format       string
}

var (
	projectInventoryCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a project inventory",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectInventoryCreateAction)
		},
		Args: cobra.NoArgs,
	}

	projectInventoryCreateArgs = projectInventoryCreateBind{}
)

func init() {
	projectInventoryCmd.AddCommand(projectInventoryCreateCmd)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.RepositoryID,
		"repository-id",
		"",
		"Repository for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.CredentialID,
		"credential-id",
		"",
		"Credential for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.BecomeID,
		"become-id",
		"",
		"Become for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.Slug,
		"slug",
		"",
		"Slug for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.Name,
		"name",
		"",
		"Name for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.Kind,
		"kind",
		"",
		"Kind for project inventory, can be static or file",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.Content,
		"content",
		"",
		"Content for project inventory",
	)

	projectInventoryCreateCmd.Flags().StringVar(
		&projectInventoryCreateArgs.Format,
		"format",
		tmplProjectInventoryShow,
		"Custom output format",
	)
}

func projectInventoryCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectInventoryCreateArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectInventoryCreateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	if projectInventoryCreateArgs.Kind == "" {
		return fmt.Errorf("you must provide a kind")
	}

	if projectInventoryCreateArgs.Content == "" {
		return fmt.Errorf("you must provide a content")
	}

	body := v1.CreateProjectInventoryJSONRequestBody{}
	changed := false

	if val := projectInventoryCreateArgs.RepositoryID; val != "" {
		body.RepositoryID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.CredentialID; val != "" {
		body.CredentialID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.BecomeID; val != "" {
		body.BecomeID = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.Name; val != "" {
		body.Name = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.Kind; val != "" {
		body.Kind = v1.ToPtr(val)
		changed = true
	}

	if val := projectInventoryCreateArgs.Content; val != "" {
		body.Content = v1.ToPtr(val)
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
		fmt.Sprintln(projectInventoryCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateProjectInventoryWithResponse(
		ccmd.Context(),
		projectInventoryCreateArgs.ProjectID,
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
