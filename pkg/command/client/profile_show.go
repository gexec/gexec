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

// tmplProfileShow represents a profile within details view.
var tmplProfileShow = "Username: \x1b[33m{{ .Username }} \x1b[0m" + `
ID: {{ .ID }}
Email: {{ .Email }}
Fullname: {{ .Fullname }}
Active: {{ .Active }}
Admin: {{ .Admin }}
Created: {{ .CreatedAt }}
Updated: {{ .UpdatedAt }}
`

type profileShowBind struct {
	Format string
}

var (
	profileShowCmd = &cobra.Command{
		Use:   "show",
		Short: "Show profile details",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, profileShowAction)
		},
		Args: cobra.NoArgs,
	}

	profileShowArgs = profileShowBind{}
)

func init() {
	profileCmd.AddCommand(profileShowCmd)

	profileShowCmd.Flags().StringVar(
		&profileShowArgs.Format,
		"format",
		tmplProfileShow,
		"Custom output format",
	)
}

func profileShowAction(ccmd *cobra.Command, _ []string, client *Client) error {
	resp, err := client.ShowProfileWithResponse(
		ccmd.Context(),
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
		fmt.Sprintln(profileShowArgs.Format),
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
