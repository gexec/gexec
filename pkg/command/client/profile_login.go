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

// tmplProfileLogin represents a expiring login token.
var tmplProfileLogin = "Token: \x1b[33m{{ .Token }} \x1b[0m" + `
Expires: {{ .ExpiresAt }}`

type profileLoginBind struct {
	Username string
	Password string
	Format   string
}

var (
	profileLoginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login by credentials",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, profileLoginAction)
		},
		Args: cobra.NoArgs,
	}

	profileLoginArgs = profileLoginBind{}
)

func init() {
	profileCmd.AddCommand(profileLoginCmd)

	profileLoginCmd.Flags().StringVar(
		&profileLoginArgs.Username,
		"username",
		"",
		"Username for authentication",
	)

	profileLoginCmd.Flags().StringVar(
		&profileLoginArgs.Password,
		"password",
		"",
		"Password for authentication",
	)

	profileLoginCmd.Flags().StringVar(
		&profileLoginArgs.Format,
		"format",
		tmplProfileLogin,
		"Custom output format",
	)
}

func profileLoginAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if profileLoginArgs.Username == "" {
		return fmt.Errorf("please provide a username")
	}

	if profileLoginArgs.Password == "" {
		return fmt.Errorf("please provide a password")
	}

	resp, err := client.LoginAuthWithResponse(
		ccmd.Context(),
		v1.LoginAuthJSONRequestBody{
			Username: profileLoginArgs.Username,
			Password: profileLoginArgs.Password,
		},
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
		fmt.Sprintln(profileLoginArgs.Format),
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
	case http.StatusUnauthorized:
		return errors.New(v1.FromPtr(resp.JSON401.Message))
	case http.StatusInternalServerError:
		if resp.JSON500 != nil {
			return errors.New(v1.FromPtr(resp.JSON500.Message))
		}

		return errors.New(http.StatusText(http.StatusInternalServerError))
	default:
		return ErrUnknownServerResponse
	}

	return nil
}
