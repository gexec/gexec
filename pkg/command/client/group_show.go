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

// tmplGroupShow represents a user within details view.
var tmplGroupShow = "Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .GroupID }}
Name: {{ .Name }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}`

type groupShowBind struct {
	GroupID string
	Format  string
}

var (
	groupShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show a group",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, groupShowAction)
		},
		Args: cobra.NoArgs,
	}

	groupShowArgs = groupShowBind{}
)

func init() {
	groupCmd.AddCommand(groupShowCmd)

	groupShowCmd.Flags().StringVar(
		&groupShowArgs.GroupID,
		"group-id",
		"",
		"Group ID or slug",
	)

	groupShowCmd.Flags().StringVar(
		&groupShowArgs.Format,
		"format",
		tmplGroupShow,
		"Custom output format",
	)
}

func groupShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if groupShowArgs.GroupID == "" {
		return fmt.Errorf("you must provide a group ID or a slug")
	}

	resp, err := client.ShowGroupWithResponse(
		ccmd.Context(),
		groupShowArgs.GroupID,
	)

	if err != nil {
		return err
	}

	tmpl, err := template.New(
		"_",
	).Funcs(
		globalFuncMap,
	).Funcs(
		basicFuncMap,
	).Parse(
		fmt.Sprintln(groupShowArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		if err := tmpl.Execute(
			os.Stdout,
			resp.JSON200,
		); err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
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
