package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/genexec/genexec/pkg/middleware/current"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/store"
	"github.com/genexec/genexec/pkg/validate"
	"github.com/rs/zerolog/log"
)

// ListProjectInventories implements the v1.ServerInterface.
func (a *API) ListProjectInventories(ctx context.Context, request ListProjectInventoriesRequestObject) (ListProjectInventoriesResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectInventories404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectInventories").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectInventories500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectInventories404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listInventoriesSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.List(
		ctx,
		parent.ID,
		model.ListParams{
			Sort:   sort,
			Order:  order,
			Limit:  limit,
			Offset: offset,
			Search: search,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListProjectInventories").
			Str("project", parent.ID).
			Msg("Failed to load inventories")

		return ListProjectInventories500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load inventories"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Inventory, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
			log.Error().
				Err(err).
				Str("action", "ListProjectInventories").
				Str("project", parent.ID).
				Msg("Failed to decrypt secrets")

			return ListProjectInventories500JSONResponse{InternalServerErrorJSONResponse{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			}}, nil
		}

		payload[id] = a.convertInventory(record)
	}

	return ListProjectInventories200JSONResponse{ProjectInventoriesResponseJSONResponse{
		Total:       count,
		Limit:       limit,
		Offset:      offset,
		Project:     ToPtr(a.convertProject(parent)),
		Inventories: payload,
	}}, nil
}

// ShowProjectInventory implements the v1.ServerInterface.
func (a *API) ShowProjectInventory(ctx context.Context, request ShowProjectInventoryRequestObject) (ShowProjectInventoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectInventory").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Show(
		ctx,
		parent.ID,
		request.InventoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrInventoryNotFound) {
			return ShowProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectInventory").
			Str("project", record.ID).
			Str("inventory", request.InventoryId).
			Msg("Failed to load inventory")

		return ShowProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowInventory(ctx, inventoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or inventory"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
		log.Error().
			Err(err).
			Str("action", "ShowProjectInventory").
			Str("project", parent.ID).
			Str("inventory", record.ID).
			Msg("Failed to decrypt secrets")

		return ShowProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return ShowProjectInventory200JSONResponse{ProjectInventoryResponseJSONResponse(
		a.convertInventory(record),
	)}, nil
}

// CreateProjectInventory implements the v1.ServerInterface.
func (a *API) CreateProjectInventory(ctx context.Context, request CreateProjectInventoryRequestObject) (CreateProjectInventoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectInventory").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateInventory(ctx, inventoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Inventory{
		ProjectID: parent.ID,
	}

	if request.Body.RepositoryId != nil {
		record.RepositoryID = FromPtr(request.Body.RepositoryId)
	}

	if request.Body.CredentialId != nil {
		record.CredentialID = FromPtr(request.Body.CredentialId)
	}

	if request.Body.BecomeId != nil {
		record.BecomeID = FromPtr(request.Body.BecomeId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Kind != nil {
		record.Kind = FromPtr(request.Body.Kind)
	}

	if request.Body.Content != nil {
		record.Content = FromPtr(request.Body.Content)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateProjectIntenvoryl").
			Str("project", parent.ID).
			Msg("Failed to encrypt secrets")

		return CreateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Create(
		ctx,
		parent.ID,
		record,
	); err != nil {
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

			return CreateProjectInventory422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate inventory"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectInventory").
			Str("project", parent.ID).
			Msg("Failed to create inventory")

		return CreateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectInventory200JSONResponse{ProjectInventoryResponseJSONResponse(
		a.convertInventory(record),
	)}, nil
}

// UpdateProjectInventory implements the v1.ServerInterface.
func (a *API) UpdateProjectInventory(ctx context.Context, request UpdateProjectInventoryRequestObject) (UpdateProjectInventoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectInventory").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Show(
		ctx,
		parent.ID,
		request.InventoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrInventoryNotFound) {
			return UpdateProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectInventory").
			Str("project", parent.ID).
			Str("inventory", request.InventoryId).
			Msg("Failed to load inventory")

		return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageInventory(ctx, inventoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or inventory"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectIntenvoryl").
			Str("project", parent.ID).
			Str("inventory", record.ID).
			Msg("Failed to decrypt secrets")

		return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if request.Body.RepositoryId != nil {
		record.RepositoryID = FromPtr(request.Body.RepositoryId)
	}

	if request.Body.CredentialId != nil {
		record.CredentialID = FromPtr(request.Body.CredentialId)
	}

	if request.Body.BecomeId != nil {
		record.BecomeID = FromPtr(request.Body.BecomeId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Kind != nil {
		record.Kind = FromPtr(request.Body.Kind)
	}

	if request.Body.Content != nil {
		record.Content = FromPtr(request.Body.Content)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectIntenvoryl").
			Str("project", parent.ID).
			Str("inventory", record.ID).
			Msg("Failed to encrypt secrets")

		return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Update(
		ctx,
		parent.ID,
		record,
	); err != nil {
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

			return UpdateProjectInventory422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate inventory"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectInventory").
			Str("project", parent.ID).
			Str("inventory", record.ID).
			Msg("Failed to update inventory")

		return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectInventory200JSONResponse{ProjectInventoryResponseJSONResponse(
		a.convertInventory(record),
	)}, nil
}

// DeleteProjectInventory implements the v1.ServerInterface.
func (a *API) DeleteProjectInventory(ctx context.Context, request DeleteProjectInventoryRequestObject) (DeleteProjectInventoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectInventory").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Show(
		ctx,
		parent.ID,
		request.InventoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrInventoryNotFound) {
			return DeleteProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or inventory"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectInventory").
			Str("project", parent.ID).
			Str("inventory", request.InventoryId).
			Msg("Failed to load inventory")

		return DeleteProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load inventory"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageInventory(ctx, inventoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectInventory404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or inventory"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Inventories.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectInventory").
			Str("project", parent.ID).
			Str("inventory", record.ID).
			Msg("Failed to delete inventory")

		return DeleteProjectInventory400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete inventory"),
		}}, nil
	}

	return DeleteProjectInventory200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted inventory"),
	}}, nil
}

func (a *API) convertInventory(record *model.Inventory) Inventory {
	result := Inventory{
		Id:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Kind:      ToPtr(InventoryKind(record.Kind)),
		Content:   ToPtr(record.Content),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.Repository != nil {
		result.RepositoryId = ToPtr(record.RepositoryID)

		result.Repository = ToPtr(
			a.convertRepository(
				record.Repository,
			),
		)
	}

	if record.Credential != nil {
		result.CredentialId = ToPtr(record.CredentialID)

		result.Credential = ToPtr(
			a.convertCredential(
				record.Credential,
			),
		)
	}

	if record.Become != nil {
		result.BecomeId = ToPtr(record.BecomeID)

		result.Become = ToPtr(
			a.convertCredential(
				record.Become,
			),
		)
	}

	return result
}

type inventoryPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Inventory
	OwnerOnly bool
}

func (a *API) permitCreateInventory(ctx context.Context, definition inventoryPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowInventory(ctx context.Context, definition inventoryPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageInventory(ctx context.Context, definition inventoryPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listInventoriesSorting(request ListProjectInventoriesRequestObject) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Params.Sort,
		request.Params.Limit,
		request.Params.Offset,
		request.Params.Search,
	)

	order := ""

	if request.Params.Order != nil {
		sort = string(FromPtr(request.Params.Order))
	}

	return sort, order, limit, offset, search
}
