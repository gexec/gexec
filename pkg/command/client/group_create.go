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

type groupCreateBind struct {
	Slug   string
	Name   string
	Format string
}

var (
	groupCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupCreateAction)
		},
		Args: cobra.NoArgs,
	}

	groupCreateArgs = groupCreateBind{}
)

func init() {
	groupCmd.AddCommand(groupCreateCmd)

	groupCreateCmd.Flags().StringVar(
		&groupCreateArgs.Slug,
		"slug",
		"",
		"Slug for group",
	)

	groupCreateCmd.Flags().StringVar(
		&groupCreateArgs.Name,
		"name",
		"",
		"Name for group",
	)

	groupCreateCmd.Flags().StringVar(
		&groupCreateArgs.Format,
		"format",
		tmplGroupShow,
		"Custom output format",
	)
}

func groupCreateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUpdateArgs.Name == "" {
		return fmt.Errorf("you must provide a name")
	}

	body := v1.CreateGroupJSONRequestBody{}
	changed := false

	if val := groupCreateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := groupCreateArgs.Name; val != "" {
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
		fmt.Sprintln(groupCreateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.CreateGroupWithResponse(
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
