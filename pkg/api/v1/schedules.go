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

// ListProjectSchedules implements the v1.ServerInterface.
func (a *API) ListProjectSchedules(ctx context.Context, request ListProjectSchedulesRequestObject) (ListProjectSchedulesResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectSchedules404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectSchedules").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectSchedules500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectSchedules404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listSchedulesSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.List(
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
			Str("action", "ListProjectSchedules").
			Str("project", parent.ID).
			Msg("Failed to load schedules")

		return ListProjectSchedules500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load schedules"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Schedule, len(records))
	for id, record := range records {
		payload[id] = a.convertSchedule(record)
	}

	return ListProjectSchedules200JSONResponse{ProjectSchedulesResponseJSONResponse{
		Total:     count,
		Limit:     limit,
		Offset:    offset,
		Project:   ToPtr(a.convertProject(parent)),
		Schedules: payload,
	}}, nil
}

// ShowProjectSchedule implements the v1.ServerInterface.
func (a *API) ShowProjectSchedule(ctx context.Context, request ShowProjectScheduleRequestObject) (ShowProjectScheduleResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectSchedule").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Show(
		ctx,
		parent.ID,
		request.ScheduleId,
	)

	if err != nil {
		if errors.Is(err, store.ErrScheduleNotFound) {
			return ShowProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectSchedule").
			Str("project", parent.ID).
			Str("schedule", request.ScheduleId).
			Msg("Failed to load schedule")

		return ShowProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowSchedule(ctx, schedulePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or schedule"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProjectSchedule200JSONResponse{ProjectScheduleResponseJSONResponse(
		a.convertSchedule(record),
	)}, nil
}

// CreateProjectSchedule implements the v1.ServerInterface.
func (a *API) CreateProjectSchedule(ctx context.Context, request CreateProjectScheduleRequestObject) (CreateProjectScheduleResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectSchedule").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateSchedule(ctx, schedulePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Schedule{
		ProjectID: parent.ID,
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Create(
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

			return CreateProjectSchedule422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate schedule"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectSchedule").
			Str("project", parent.ID).
			Msg("Failed to create schedule")

		return CreateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectSchedule200JSONResponse{ProjectScheduleResponseJSONResponse(
		a.convertSchedule(record),
	)}, nil
}

// UpdateProjectSchedule implements the v1.ServerInterface.
func (a *API) UpdateProjectSchedule(ctx context.Context, request UpdateProjectScheduleRequestObject) (UpdateProjectScheduleResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectSchedule").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Show(
		ctx,
		parent.ID,
		request.ScheduleId,
	)

	if err != nil {
		if errors.Is(err, store.ErrScheduleNotFound) {
			return UpdateProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectSchedule").
			Str("project", parent.ID).
			Str("schedule", request.ScheduleId).
			Msg("Failed to load schedule")

		return UpdateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageSchedule(ctx, schedulePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or schedule"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Update(
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

			return UpdateProjectSchedule422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate schedule"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectSchedule").
			Str("project", parent.ID).
			Str("schedule", record.ID).
			Msg("Failed to update schedule")

		return UpdateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectSchedule200JSONResponse{ProjectScheduleResponseJSONResponse(
		a.convertSchedule(record),
	)}, nil
}

// DeleteProjectSchedule implements the v1.ServerInterface.
func (a *API) DeleteProjectSchedule(ctx context.Context, request DeleteProjectScheduleRequestObject) (DeleteProjectScheduleResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectSchedule").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Show(
		ctx,
		parent.ID,
		request.ScheduleId,
	)

	if err != nil {
		if errors.Is(err, store.ErrScheduleNotFound) {
			return DeleteProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or schedule"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectSchedule").
			Str("project", parent.ID).
			Str("schedule", request.ScheduleId).
			Msg("Failed to load schedule")

		return DeleteProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageSchedule(ctx, schedulePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectSchedule404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or schedule"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectSchedule").
			Str("project", parent.ID).
			Str("schedule", record.ID).
			Msg("Failed to delete schedule")

		return DeleteProjectSchedule400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete schedule"),
		}}, nil
	}

	return DeleteProjectSchedule200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted schedule"),
	}}, nil
}

func (a *API) convertSchedule(record *model.Schedule) Schedule {
	result := Schedule{
		Id:        ToPtr(record.ID),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

type schedulePermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Schedule
	OwnerOnly bool
}

func (a *API) permitCreateSchedule(ctx context.Context, definition schedulePermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowSchedule(ctx context.Context, definition schedulePermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageSchedule(ctx context.Context, definition schedulePermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listSchedulesSorting(request ListProjectSchedulesRequestObject) (string, string, int64, int64, string) {
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
