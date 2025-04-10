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

type userUpdateBind struct {
	UserID   string
	Username string
	Password string
	Email    string
	Fullname string
	Active   bool
	Inactive bool
	Admin    bool
	Regular  bool
	Format   string
}

var (
	userUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update an user",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, userUpdateAction)
		},
		Args: cobra.NoArgs,
	}

	userUpdateArgs = userUpdateBind{}
)

func init() {
	userCmd.AddCommand(userUpdateCmd)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.UserID,
		"user-id",
		"",
		"User ID or slug",
	)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.Username,
		"username",
		"",
		"Username for user",
	)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.Password,
		"password",
		"",
		"Password for user",
	)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.Email,
		"email",
		"",
		"Email for user",
	)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.Fullname,
		"fullname",
		"",
		"Fullname for user",
	)

	userUpdateCmd.Flags().BoolVar(
		&userUpdateArgs.Active,
		"active",
		false,
		"Mark user as active",
	)

	userUpdateCmd.Flags().BoolVar(
		&userUpdateArgs.Inactive,
		"inactive",
		false,
		"Mark user as inactive",
	)

	userUpdateCmd.Flags().BoolVar(
		&userUpdateArgs.Admin,
		"admin",
		false,
		"Mark user as admin",
	)

	userUpdateCmd.Flags().BoolVar(
		&userUpdateArgs.Regular,
		"regular",
		false,
		"Mark user as regular",
	)

	userUpdateCmd.Flags().StringVar(
		&userUpdateArgs.Format,
		"format",
		tmplUserShow,
		"Format for successful output",
	)
}

func userUpdateAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if userUpdateArgs.UserID == "" {
		return fmt.Errorf("you must provide a user ID or a slug")
	}

	body := v1.UpdateUserJSONRequestBody{}
	changed := false

	if val := userUpdateArgs.Username; val != "" {
		body.Username = v1.ToPtr(val)
		changed = true
	}

	if val := userUpdateArgs.Password; val != "" {
		body.Password = v1.ToPtr(val)
		changed = true
	}

	if val := userUpdateArgs.Email; val != "" {
		body.Email = v1.ToPtr(val)
		changed = true
	}

	if val := userUpdateArgs.Fullname; val != "" {
		body.Fullname = v1.ToPtr(val)
		changed = true
	}

	if val := userUpdateArgs.Active; val {
		body.Active = v1.ToPtr(true)
		changed = true
	}

	if val := userUpdateArgs.Inactive; val {
		body.Active = v1.ToPtr(false)
		changed = true
	}

	if val := userUpdateArgs.Admin; val {
		body.Admin = v1.ToPtr(true)
		changed = true
	}

	if val := userUpdateArgs.Regular; val {
		body.Admin = v1.ToPtr(false)
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
		fmt.Sprintln(userUpdateArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	resp, err := client.UpdateUserWithResponse(
		ccmd.Context(),
		userUpdateArgs.UserID,
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
