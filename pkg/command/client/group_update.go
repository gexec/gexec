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

type groupUpdateBind struct {
	ID     string
	Slug   string
	Name   string
	Format string
}

var (
	groupUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	groupUpdateArgs = groupUpdateBind{}
)

func init() {
	groupCmd.AddCommand(groupUpdateCmd)

	groupUpdateCmd.Flags().StringVarP(
		&groupUpdateArgs.ID,
		"id",
		"i",
		"",
		"Group ID or slug",
	)

	groupUpdateCmd.Flags().StringVar(
		&groupUpdateArgs.Slug,
		"slug",
		"",
		"Slug for group",
	)

	groupUpdateCmd.Flags().StringVar(
		&groupUpdateArgs.Name,
		"name",
		"",
		"Name for group",
	)

	groupUpdateCmd.Flags().StringVar(
		&groupUpdateArgs.Format,
		"format",
		tmplGroupShow,
		"Custom output format",
	)
}

func groupUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupUpdateArgs.ID == "" {
		return fmt.Errorf("you must provide an ID or a slug")
	}

	body := v1.UpdateGroupJSONRequestBody{}
	changed := false

	if val := groupUpdateArgs.Slug; val != "" {
		body.Slug = v1.ToPtr(val)
		changed = true
	}

	if val := groupUpdateArgs.Name; val != "" {
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
		fmt.Sprintln(groupUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateGroupWithResponse(
		ccmd.Context(),
		groupUpdateArgs.ID,
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
