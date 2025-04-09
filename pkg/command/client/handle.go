package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// ErrMissingRequiredCredentials defines the error for missing credentials.
	ErrMissingRequiredCredentials = errors.New("missing credentials")

	// ErrUnknownServerResponse defines the error fot unknown server responses.
	ErrUnknownServerResponse = errors.New("unknown response from api")
)

// Client simply wraps the openapi client including authentication.
type Client struct {
	*v1.ClientWithResponses
}

// HandleFunc is the real handle implementation.
type HandleFunc func(ccmd *cobra.Command, args []string, client *Client) error

// Handle wraps the command function handler.
func Handle(ccmd *cobra.Command, args []string, fn HandleFunc) {
	if viper.GetString("server.address") == "" {
		fmt.Fprintf(os.Stderr, "Error: You must provide the server address.\n")
		os.Exit(1)
	}

	serverAddress, err := url.JoinPath(
		viper.GetString("server.address"),
		"api",
		"v1",
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid server address, bad format?\n")
		os.Exit(1)
	}

	server, err := url.Parse(
		serverAddress,
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid server address, bad format?\n")
		os.Exit(1)
	}

	child, err := v1.NewClientWithResponses(
		server.String(),
		v1.WithRequestEditorFn(WithTokenAuth(
			viper.GetString("server.token"),
		)),
		v1.WithRequestEditorFn(WithBasicAuth(
			viper.GetString("server.username"),
			viper.GetString("server.password"),
		)),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize client library\n")
		os.Exit(1)
	}

	client := &Client{
		ClientWithResponses: child,
	}

	if err := fn(ccmd, args, client); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(2)
	}
}

// WithTokenAuth integrates token auth into the sdk.
func WithTokenAuth(token string) v1.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		if token != "" {
			req.Header.Set(
				"X-API-Key",
				token,
			)
		}
		return nil
	}
}

// WithBasicAuth integrates basic auth into the sdk.
func WithBasicAuth(username, password string) v1.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		if username != "" && password != "" {
			req.SetBasicAuth(
				username,
				password,
			)
		}
		return nil
	}
}

var tmplValidationError = `{{ .Message }}
{{ range $validation := .Errors }}
* {{ $validation.Field }}: {{ $validation.Message }}
{{ end }}
`

func validationError(notification *v1.Notification) error {
	tmpl, err := template.New(
		"_",
	).Funcs(
		globalFuncMap,
	).Funcs(
		basicFuncMap,
	).Parse(
		tmplValidationError,
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	message := bytes.NewBufferString("")

	if err := tmpl.Execute(
		message,
		notification,
	); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return errors.New(message.String())
}
