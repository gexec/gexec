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

// ListProjectEnvironments implements the v1.ServerInterface.
func (a *API) ListProjectEnvironments(ctx context.Context, request ListProjectEnvironmentsRequestObject) (ListProjectEnvironmentsResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectEnvironments404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectEnvironments").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectEnvironments500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectEnvironments404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listEnvironmentsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.List(
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
			Str("action", "ListProjectEnvironments").
			Str("project", parent.ID).
			Msg("Failed to load environments")

		return ListProjectEnvironments500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load environments"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Environment, len(records))
	for id, record := range records {
		payload[id] = a.convertEnvironment(record)
	}

	return ListProjectEnvironments200JSONResponse{ProjectEnvironmentsResponseJSONResponse{
		Total:        count,
		Limit:        limit,
		Offset:       offset,
		Project:      ToPtr(a.convertProject(parent)),
		Environments: payload,
	}}, nil
}

// ShowProjectEnvironment implements the v1.ServerInterface.
func (a *API) ShowProjectEnvironment(ctx context.Context, request ShowProjectEnvironmentRequestObject) (ShowProjectEnvironmentResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectEnvironment").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Show(
		ctx,
		parent.ID,
		request.EnvironmentId,
	)

	if err != nil {
		if errors.Is(err, store.ErrEnvironmentNotFound) {
			return ShowProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectEnvironment").
			Str("project", parent.ID).
			Str("project", request.EnvironmentId).
			Msg("Failed to load environment")

		return ShowProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowEnvironment(ctx, environmentPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or environment"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProjectEnvironment200JSONResponse{ProjectEnvironmentResponseJSONResponse(
		a.convertEnvironment(record),
	)}, nil
}

// CreateProjectEnvironment implements the v1.ServerInterface.
func (a *API) CreateProjectEnvironment(ctx context.Context, request CreateProjectEnvironmentRequestObject) (CreateProjectEnvironmentResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectEnvironment").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateEnvironment(ctx, environmentPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Environment{
		ProjectID: parent.ID,
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Create(
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

			return CreateProjectEnvironment422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate environment"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectEnvironment").
			Str("project", parent.ID).
			Msg("Failed to create environment")

		return CreateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectEnvironment200JSONResponse{ProjectEnvironmentResponseJSONResponse(
		a.convertEnvironment(record),
	)}, nil
}

// UpdateProjectEnvironment implements the v1.ServerInterface.
func (a *API) UpdateProjectEnvironment(ctx context.Context, request UpdateProjectEnvironmentRequestObject) (UpdateProjectEnvironmentResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectEnvironment").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Show(
		ctx,
		parent.ID,
		request.EnvironmentId,
	)

	if err != nil {
		if errors.Is(err, store.ErrEnvironmentNotFound) {
			return UpdateProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectEnvironment").
			Str("project", parent.ID).
			Str("environment", request.EnvironmentId).
			Msg("Failed to load environment")

		return UpdateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageEnvironment(ctx, environmentPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or environment"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Update(
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

			return UpdateProjectEnvironment422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate environment"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectEnvironment").
			Str("project", parent.ID).
			Str("environment", record.ID).
			Msg("Failed to update environment")

		return UpdateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectEnvironment200JSONResponse{ProjectEnvironmentResponseJSONResponse(
		a.convertEnvironment(record),
	)}, nil
}

// DeleteProjectEnvironment implements the v1.ServerInterface.
func (a *API) DeleteProjectEnvironment(ctx context.Context, request DeleteProjectEnvironmentRequestObject) (DeleteProjectEnvironmentResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectEnvironment").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Show(
		ctx,
		parent.ID,
		request.EnvironmentId,
	)

	if err != nil {
		if errors.Is(err, store.ErrEnvironmentNotFound) {
			return DeleteProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or environment"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectEnvironment").
			Str("project", parent.ID).
			Str("environment", request.EnvironmentId).
			Msg("Failed to load environment")

		return DeleteProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageEnvironment(ctx, environmentPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectEnvironment404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or environment"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectEnvironment").
			Str("project", parent.ID).
			Str("environment", record.ID).
			Msg("Failed to delete environment")

		return DeleteProjectEnvironment400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete environment"),
		}}, nil
	}

	return DeleteProjectEnvironment200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted environment"),
	}}, nil
}

func (a *API) convertEnvironment(record *model.Environment) Environment {
	result := Environment{
		Id:        ToPtr(record.ID),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

type environmentPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Environment
	OwnerOnly bool
}

func (a *API) permitCreateEnvironment(ctx context.Context, definition environmentPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowEnvironment(ctx context.Context, definition environmentPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageEnvironment(ctx context.Context, definition environmentPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listEnvironmentsSorting(request ListProjectEnvironmentsRequestObject) (string, string, int64, int64, string) {
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
