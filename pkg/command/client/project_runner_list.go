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

type projectRunnerListBind struct {
	Project string
	Format  string
}

// tmplProjectRunnerList represents a row within project runner listing.
var tmplProjectRunnerList = "{{ range . }}Slug: \x1b[33m{{ .Slug }} \x1b[0m" + `
ID: {{ .ID }}
Name: {{ .Name }}

{{ end -}}`

var (
	projectRunnerListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all runners for a project",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectRunnerListAction)
		},
		Args: cobra.NoArgs,
	}

	projectRunnerListArgs = projectRunnerListBind{}
)

func init() {
	projectRunnerCmd.AddCommand(projectRunnerListCmd)

	projectRunnerListCmd.Flags().StringVarP(
		&projectRunnerListArgs.Project,
		"project",
		"p",
		"",
		"Project ID or slug",
	)

	projectRunnerListCmd.Flags().StringVar(
		&projectRunnerListArgs.Format,
		"format",
		tmplProjectRunnerList,
		"Custom output format",
	)
}

func projectRunnerListAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectRunnerListArgs.Project == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	resp, err := client.ListProjectRunnersWithResponse(
		ccmd.Context(),
		projectRunnerListArgs.Project,
		&v1.ListProjectRunnersParams{
			Limit:  v1.ToPtr(10000),
			Offset: v1.ToPtr(0),
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
		fmt.Sprintln(projectRunnerListArgs.Format),
	)

	if err != nil {
		return fmt.Errorf("failed to process template: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		records := resp.JSON200.Runners

		if len(records) == 0 {
			fmt.Fprintln(os.Stderr, "Empty result")
			return nil
		}

		if err := tmpl.Execute(
			os.Stdout,
			records,
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
