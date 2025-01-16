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

// ListProjectExecutions implements the v1.ServerInterface.
func (a *API) ListProjectExecutions(ctx context.Context, request ListProjectExecutionsRequestObject) (ListProjectExecutionsResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectExecutions404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectExecutions").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectExecutions500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectExecutions404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listExecutionsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.List(
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
			Str("action", "ListProjectExecutions").
			Str("project", parent.ID).
			Msg("Failed to load executions")

		return ListProjectExecutions500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load executions"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Execution, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
			log.Error().
				Err(err).
				Str("action", "ListProjectExecutions").
				Str("project", parent.ID).
				Msg("Failed to decrypt secrets")

			return ListProjectExecutions500JSONResponse{InternalServerErrorJSONResponse{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			}}, nil
		}

		payload[id] = a.convertExecution(record)
	}

	return ListProjectExecutions200JSONResponse{ProjectExecutionsResponseJSONResponse{
		Total:      count,
		Limit:      limit,
		Offset:     offset,
		Project:    ToPtr(a.convertProject(parent)),
		Executions: payload,
	}}, nil
}

// ShowProjectExecution implements the v1.ServerInterface.
func (a *API) ShowProjectExecution(ctx context.Context, request ShowProjectExecutionRequestObject) (ShowProjectExecutionResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or execution"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectExecution").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Show(
		ctx,
		parent.ID,
		request.ExecutionId,
	)

	if err != nil {
		if errors.Is(err, store.ErrExecutionNotFound) {
			return ShowProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or execution"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectExecution").
			Str("project", record.ID).
			Str("execution", request.ExecutionId).
			Msg("Failed to load execution")

		return ShowProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load execution"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowExecution(ctx, executionPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or execution"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
		log.Error().
			Err(err).
			Str("action", "ShowProjectExecution").
			Str("project", parent.ID).
			Str("execution", record.ID).
			Msg("Failed to decrypt secrets")

		return ShowProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return ShowProjectExecution200JSONResponse{ProjectExecutionResponseJSONResponse(
		a.convertExecution(record),
	)}, nil
}

// CreateProjectExecution implements the v1.ServerInterface.
func (a *API) CreateProjectExecution(ctx context.Context, request CreateProjectExecutionRequestObject) (CreateProjectExecutionResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectExecution").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateExecution(ctx, executionPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Execution{
		ProjectID: parent.ID,
	}

	if request.Body.TemplateId != nil {
		record.TemplateID = FromPtr(request.Body.TemplateId)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateProjectExecution").
			Str("project", parent.ID).
			Msg("Failed to encrypt secrets")

		return CreateProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Create(
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

			return CreateProjectExecution422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate execution"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectExecution").
			Str("project", parent.ID).
			Msg("Failed to create execution")

		return CreateProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create execution"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectExecution200JSONResponse{ProjectExecutionResponseJSONResponse(
		a.convertExecution(record),
	)}, nil
}

// DeleteProjectExecution implements the v1.ServerInterface.
func (a *API) DeleteProjectExecution(ctx context.Context, request DeleteProjectExecutionRequestObject) (DeleteProjectExecutionResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or execution"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectExecution").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Show(
		ctx,
		parent.ID,
		request.ExecutionId,
	)

	if err != nil {
		if errors.Is(err, store.ErrExecutionNotFound) {
			return DeleteProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or execution"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectExecution").
			Str("project", parent.ID).
			Str("execution", request.ExecutionId).
			Msg("Failed to load execution")

		return DeleteProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load execution"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageExecution(ctx, executionPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectExecution404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or execution"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectExecution").
			Str("project", parent.ID).
			Str("execution", record.ID).
			Msg("Failed to delete execution")

		return DeleteProjectExecution400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete execution"),
		}}, nil
	}

	return DeleteProjectExecution200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted execution"),
	}}, nil
}

// PurgeProjectExecution implements the v1.ServerInterface.
func (a *API) PurgeProjectExecution(_ context.Context, _ PurgeProjectExecutionRequestObject) (PurgeProjectExecutionResponseObject, error) {
	return PurgeProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// StopProjectExecution implements the v1.ServerInterface.
func (a *API) StopProjectExecution(_ context.Context, _ StopProjectExecutionRequestObject) (StopProjectExecutionResponseObject, error) {
	return StopProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// OutputProjectExecution implements the v1.ServerInterface.
func (a *API) OutputProjectExecution(_ context.Context, _ OutputProjectExecutionRequestObject) (OutputProjectExecutionResponseObject, error) {
	return OutputProjectExecution500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

func (a *API) convertExecution(record *model.Execution) Execution {
	result := Execution{
		Id:        ToPtr(record.ID),
		Status:    ToPtr(record.Status),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.Template != nil {
		result.TemplateId = ToPtr(record.TemplateID)

		result.Template = ToPtr(
			a.convertTemplate(
				record.Template,
			),
		)
	}

	return result
}

type executionPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Execution
	OwnerOnly bool
}

func (a *API) permitCreateExecution(ctx context.Context, definition executionPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowExecution(ctx context.Context, definition executionPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageExecution(ctx context.Context, definition executionPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listExecutionsSorting(request ListProjectExecutionsRequestObject) (string, string, int64, int64, string) {
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
