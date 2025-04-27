package command

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/spf13/cobra"
)

type projectInventoryDeleteBind struct {
	ProjectID   string
	InventoryID string
}

var (
	projectInventoryDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a project inventory",
		Run: func(ccmd *cobra.Command, args []string) {
			Handle(ccmd, args, projectInventoryDeleteAction)
		},
		Args: cobra.NoArgs,
	}

	projectInventoryDeleteArgs = projectInventoryDeleteBind{}
)

func init() {
	projectInventoryCmd.AddCommand(projectInventoryDeleteCmd)

	projectInventoryDeleteCmd.Flags().StringVar(
		&projectInventoryDeleteArgs.ProjectID,
		"project-id",
		"",
		"Project ID or slug",
	)

	projectInventoryDeleteCmd.Flags().StringVar(
		&projectInventoryDeleteArgs.InventoryID,
		"inventory-id",
		"",
		"Inventory ID or slug",
	)
}

func projectInventoryDeleteAction(ccmd *cobra.Command, _ []string, client *Client) error {
	if projectInventoryDeleteArgs.ProjectID == "" {
		return fmt.Errorf("you must provide a project ID or a slug")
	}

	if projectInventoryDeleteArgs.InventoryID == "" {
		return fmt.Errorf("you must provide a inventory ID or a slug")
	}

	resp, err := client.DeleteProjectInventoryWithResponse(
		ccmd.Context(),
		projectInventoryDeleteArgs.ProjectID,
		projectInventoryDeleteArgs.InventoryID,
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
