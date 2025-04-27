package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectScheduleDeleteBind struct {
	ProjectID  string
	ScheduleID string
}

var (
	projectScheduleDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project schedule",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectScheduleDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectScheduleDeleteArgs = projectScheduleDeleteBind{}
)

func init() {
	projectScheduleCmd.AddCommand(projectScheduleDeleteCmd)

	projectScheduleDeleteCmd.Flags().StringVar(
		&projectScheduleDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectScheduleDeleteCmd.Flags().StringVar(
		&projectScheduleDeleteArgs.ScheduleID,
		"schedule-id",
		"",
		"Schedule ID or slug",
	)
}

func projectScheduleDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectScheduleDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectScheduleDeleteArgs.ScheduleID == "" {
		return fmt.Errorf("you must provide a schedule ID or a slug")
	}

	resp, err := client.DeleteProjectScheduleWithResponse(
		ccmd.Context(),
		projectScheduleDeleteArgs.ProjectID,
		projectScheduleDeleteArgs.ScheduleID,
	)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Fprintln(os.Stderr, "Successfully deleted")
	case http.StatusForbidden:
		if resp.JSON403 != nil {
			return errors.New(v1.FromPtr(resp.JSON403.Message))
		}

		return errors.New(http.StatusText(http.StatusForbidden))
	case http.StatusBadRequest:
		if resp.JSON400 != nil {
			return errors.New(v1.FromPtr(resp.JSON400.Message))
		}

		return errors.New(http.StatusText(http.StatusBadRequest))
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
