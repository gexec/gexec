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

// ListProjectTasks implements the v1.ServerInterface.
func (a *API) ListProjectTasks(ctx context.Context, request ListProjectTasksRequestObject) (ListProjectTasksResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectTasks404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectTasks").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectTasks500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectTasks404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listTasksSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Tasks.List(
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
			Str("action", "ListProjectTasks").
			Str("project", parent.ID).
			Msg("Failed to load tasks")

		return ListProjectTasks500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load tasks"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Task, len(records))
	for id, record := range records {
		payload[id] = a.convertTask(record)
	}

	return ListProjectTasks200JSONResponse{ProjectTasksResponseJSONResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(parent)),
		Tasks:   payload,
	}}, nil
}

// ShowProjectTask implements the v1.ServerInterface.
func (a *API) ShowProjectTask(ctx context.Context, request ShowProjectTaskRequestObject) (ShowProjectTaskResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectTask404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or task"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectTask").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Tasks.Show(
		ctx,
		parent.ID,
		request.TaskId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTaskNotFound) {
			return ShowProjectTask404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or task"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectTask").
			Str("project", record.ID).
			Str("task", request.TaskId).
			Msg("Failed to load task")

		return ShowProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load task"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowTask(ctx, taskPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectTask404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or task"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProjectTask200JSONResponse{ProjectTaskResponseJSONResponse(
		a.convertTask(record),
	)}, nil
}

// CreateProjectTask implements the v1.ServerInterface.
func (a *API) CreateProjectTask(ctx context.Context, request CreateProjectTaskRequestObject) (CreateProjectTaskResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectTask404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectTask").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateTask(ctx, taskPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectTask404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Task{
		ProjectID: parent.ID,
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Tasks.Create(
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

			return CreateProjectTask422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate task"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectTask").
			Str("project", parent.ID).
			Msg("Failed to create task")

		return CreateProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create task"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectTask200JSONResponse{ProjectTaskResponseJSONResponse(
		a.convertTask(record),
	)}, nil
}

// DeleteProjectTask implements the v1.ServerInterface.
func (a *API) DeleteProjectTask(ctx context.Context, request DeleteProjectTaskRequestObject) (DeleteProjectTaskResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectTask404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or task"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectTask").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Tasks.Show(
		ctx,
		parent.ID,
		request.TaskId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTaskNotFound) {
			return DeleteProjectTask404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or task"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectTask").
			Str("project", parent.ID).
			Str("task", request.TaskId).
			Msg("Failed to load task")

		return DeleteProjectTask500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load task"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageTask(ctx, taskPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectTask404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or task"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Tasks.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectTask").
			Str("project", parent.ID).
			Str("task", record.ID).
			Msg("Failed to delete task")

		return DeleteProjectTask400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete task"),
		}}, nil
	}

	return DeleteProjectTask200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted task"),
	}}, nil
}

// StopProjectTask implements the v1.ServerInterface.
func (a *API) StopProjectTask(_ context.Context, _ StopProjectTaskRequestObject) (StopProjectTaskResponseObject, error) {
	return StopProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// OutputProjectTask implements the v1.ServerInterface.
func (a *API) OutputProjectTask(_ context.Context, _ OutputProjectTaskRequestObject) (OutputProjectTaskResponseObject, error) {
	return OutputProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

func (a *API) convertTask(record *model.Task) Task {
	result := Task{
		Id:        ToPtr(record.ID),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

type taskPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Task
	OwnerOnly bool
}

func (a *API) permitCreateTask(ctx context.Context, definition taskPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowTask(ctx context.Context, definition taskPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageTask(ctx context.Context, definition taskPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listTasksSorting(request ListProjectTasksRequestObject) (string, string, int64, int64, string) {
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
