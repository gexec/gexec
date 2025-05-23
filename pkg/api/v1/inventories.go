package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
)

// ListProjectInventories implements the v1.ServerInterface.
func (a *API) ListProjectInventories(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectInventoriesParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listInventoriesSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.List(
		ctx,
		project.ID,
		model.ListParams{
			Sort:   sort,
			Order:  order,
			Limit:  limit,
			Offset: offset,
			Search: search,
		},
	)

	if err != nil {
		slog.Error(
			"Failed to load inventories",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "ListProjectInventories"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load inventories"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Inventory, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			slog.Error(
				"Failed to decrypt secrets",
				slog.Any("error", err),
				slog.String("project", project.ID),
				slog.String("action", "ListProjectInventories"),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertInventory(record)
	}

	render.JSON(w, r, ProjectInventoriesResponse{
		Total:       count,
		Limit:       limit,
		Offset:      offset,
		Project:     ToPtr(a.convertProject(project)),
		Inventories: payload,
	})
}

// ShowProjectInventory implements the v1.ServerInterface.
func (a *API) ShowProjectInventory(w http.ResponseWriter, r *http.Request, _ ProjectID, _ InventoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectInventoryFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", project.ID),
			slog.String("action", "ShowProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectInventoryResponse(
		a.convertInventory(record),
	))
}

// CreateProjectInventory implements the v1.ServerInterface.
func (a *API) CreateProjectInventory(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectInventoryBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	incoming := &model.Inventory{
		ProjectID: project.ID,
	}

	if body.RepositoryID != nil {
		incoming.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.CredentialID != nil {
		incoming.CredentialID = FromPtr(body.CredentialID)
	}

	if body.BecomeID != nil {
		incoming.BecomeID = FromPtr(body.BecomeID)
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Kind != nil {
		incoming.Kind = FromPtr(body.Kind)
	}

	if body.Content != nil {
		incoming.Content = FromPtr(body.Content)
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectIntenvory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Create(
		ctx,
		project,
		incoming,
	)

	if err != nil {
		if v, ok := err.(validate.Errors); ok {
			errors := make([]Validation, 0)

			for _, verr := range v.Errors {
				errors = append(
					errors,
					Validation{
						Field:   ToPtr(verr.Field),
						Message: ToPtr(verr.Error.Error()),
					},
				)
			}

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate inventory"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to create inventory",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectInventoryResponse(
		a.convertInventory(record),
	))
}

// UpdateProjectInventory implements the v1.ServerInterface.
func (a *API) UpdateProjectInventory(w http.ResponseWriter, r *http.Request, _ ProjectID, _ InventoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	incoming := a.ProjectInventoryFromContext(ctx)
	body := &UpdateProjectInventoryBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", incoming.ID),
			slog.String("action", "UpdateProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := incoming.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", project.ID),
			slog.String("action", "UpdateProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.RepositoryID != nil {
		incoming.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.CredentialID != nil {
		incoming.CredentialID = FromPtr(body.CredentialID)
	}

	if body.BecomeID != nil {
		incoming.BecomeID = FromPtr(body.BecomeID)
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Kind != nil {
		incoming.Kind = FromPtr(body.Kind)
	}

	if body.Content != nil {
		incoming.Content = FromPtr(body.Content)
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", incoming.ID),
			slog.String("action", "UpdateProjectIntenvory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Update(
		ctx,
		project,
		incoming,
	)

	if err != nil {
		if v, ok := err.(validate.Errors); ok {
			errors := make([]Validation, 0)

			for _, verr := range v.Errors {
				errors = append(
					errors,
					Validation{
						Field:   ToPtr(verr.Field),
						Message: ToPtr(verr.Error.Error()),
					},
				)
			}

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate inventory"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update inventory",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", record.ID),
			slog.String("action", "UpdateProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectInventoryResponse(
		a.convertInventory(record),
	))
}

// DeleteProjectInventory implements the v1.ServerInterface.
func (a *API) DeleteProjectInventory(w http.ResponseWriter, r *http.Request, _ ProjectID, _ InventoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectInventoryFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete inventory",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("inventory", record.ID),
			slog.String("action", "DeletProjectInventory"),
		)

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete inventory"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted inventory"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertInventory(record *model.Inventory) Inventory {
	result := Inventory{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Kind:      ToPtr(InventoryKind(record.Kind)),
		Content:   ToPtr(record.Content),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.Repository != nil {
		result.RepositoryID = ToPtr(record.RepositoryID)

		result.Repository = ToPtr(
			a.convertRepository(
				record.Repository,
			),
		)
	}

	if record.Credential != nil {
		result.CredentialID = ToPtr(record.CredentialID)

		result.Credential = ToPtr(
			a.convertCredential(
				record.Credential,
			),
		)
	}

	if record.Become != nil {
		result.BecomeID = ToPtr(record.BecomeID)

		result.Become = ToPtr(
			a.convertCredential(
				record.Become,
			),
		)
	}

	return result
}

// AllowShowProjectInventory defines a middleware to check permissions.
func (a *API) AllowShowProjectInventory(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectInventory defines a middleware to check permissions.
func (a *API) AllowManageProjectInventory(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listInventoriesSorting(request ListProjectInventoriesParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	if request.Order != nil {
		order = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}
