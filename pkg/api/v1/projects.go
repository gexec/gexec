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

// ListProjects implements the v1.ServerInterface.
func (a *API) ListProjects(ctx context.Context, request ListProjectsRequestObject) (ListProjectsResponseObject, error) {
	sort, order, limit, offset, search := listProjectsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.List(
		ctx,
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
			Str("action", "ListProjects").
			Msg("Failed to load projects")

		return ListProjects500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Project, len(records))
	for id, record := range records {
		payload[id] = a.convertProject(record)
	}

	return ListProjects200JSONResponse{ProjectsResponseJSONResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		Projects: payload,
	}}, nil
}

// ShowProject implements the v1.ServerInterface.
func (a *API) ShowProject(ctx context.Context, request ShowProjectRequestObject) (ShowProjectResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProject").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProject500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: record,
	}) {
		return ShowProject404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProject200JSONResponse{ProjectResponseJSONResponse(
		a.convertProject(record),
	)}, nil
}

// CreateProject implements the v1.ServerInterface.
func (a *API) CreateProject(ctx context.Context, request CreateProjectRequestObject) (CreateProjectResponseObject, error) {
	if !a.permitCreateProject(ctx, projectPermissions{
		User: current.GetUser(ctx),
	}) {
		return CreateProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("You are not authorized to create projects"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record := &model.Project{}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Demo != nil {
		record.Demo = FromPtr(request.Body.Demo)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Create(
		ctx,
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

			return CreateProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProject").
			Msg("Failed to create project")

		return CreateProject500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProject200JSONResponse{ProjectResponseJSONResponse(
		a.convertProject(record),
	)}, nil
}

// UpdateProject implements the v1.ServerInterface.
func (a *API) UpdateProject(ctx context.Context, request UpdateProjectRequestObject) (UpdateProjectResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProject").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProject500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: record,
	}) {
		return UpdateProject404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Update(
		ctx,
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

			return UpdateProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProject").
			Str("project", record.ID).
			Msg("Failed to update project")

		return UpdateProject500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProject200JSONResponse{ProjectResponseJSONResponse(
		a.convertProject(record),
	)}, nil
}

// DeleteProject implements the v1.ServerInterface.
func (a *API) DeleteProject(ctx context.Context, request DeleteProjectRequestObject) (DeleteProjectResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProject").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProject500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return DeleteProject404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Delete(
		ctx,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProject").
			Str("project", record.ID).
			Msg("Failed to delete project")

		return DeleteProject400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete project"),
		}}, nil
	}

	return DeleteProject200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted project"),
	}}, nil
}

// ListProjectTeams implements the v1.ServerInterface.
func (a *API) ListProjectTeams(ctx context.Context, request ListProjectTeamsRequestObject) (ListProjectTeamsResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectTeams404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectTeams").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectTeams500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: record,
	}) {
		return ListProjectTeams404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listProjectTeamsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.ListTeams(
		ctx,
		model.TeamProjectParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			ProjectID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListProjectTeams").
			Str("project", record.ID).
			Msg("Failed to load project teams")

		return ListProjectTeams500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project teams"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]TeamProject, len(records))
	for id, record := range records {
		payload[id] = a.convertProjectTeam(record)
	}

	return ListProjectTeams200JSONResponse{ProjectTeamsResponseJSONResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(record)),
		Teams:   payload,
	}}, nil
}

// AttachProjectToTeam implements the v1.ServerInterface.
func (a *API) AttachProjectToTeam(ctx context.Context, request AttachProjectToTeamRequestObject) (AttachProjectToTeamResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return AttachProjectToTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachProjectToTeam").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return AttachProjectToTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return AttachProjectToTeam404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.AttachTeam(
		ctx,
		model.TeamProjectParams{
			ProjectID: record.ID,
			TeamID:    request.Body.Team,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return AttachProjectToTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachProjectToTeam412JSONResponse{AlreadyAttachedErrorJSONResponse{
				Message: ToPtr("Team is already attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

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

			return AttachProjectToTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachProjectToTeam").
			Str("project", record.ID).
			Str("team", request.Body.Team).
			Msg("Failed to attach project to team")

		return AttachProjectToTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach project to team"),
		}}, nil
	}

	return AttachProjectToTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached project to team"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitProjectTeam implements the v1.ServerInterface.
func (a *API) PermitProjectTeam(ctx context.Context, request PermitProjectTeamRequestObject) (PermitProjectTeamResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return PermitProjectTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitProjectTeam").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return PermitProjectTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return PermitProjectTeam404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.PermitTeam(
		ctx,
		model.TeamProjectParams{
			ProjectID: record.ID,
			TeamID:    request.Body.Team,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return PermitProjectTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitProjectTeam412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Team is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

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

			return PermitProjectTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitProjectTeam").
			Str("project", record.ID).
			Str("team", request.Body.Team).
			Msg("Failed to update project team perms")

		return PermitProjectTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update project team perms"),
		}}, nil
	}

	return PermitProjectTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated project team perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteProjectFromTeam implements the v1.ServerInterface.
func (a *API) DeleteProjectFromTeam(ctx context.Context, request DeleteProjectFromTeamRequestObject) (DeleteProjectFromTeamResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectFromTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectFromTeam").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectFromTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return DeleteProjectFromTeam404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.DropTeam(
		ctx,
		model.TeamProjectParams{
			ProjectID: record.ID,
			TeamID:    request.Body.Team,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return DeleteProjectFromTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteProjectFromTeam412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Team is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectFromTeam").
			Str("project", record.ID).
			Str("team", request.Body.Team).
			Msg("Failed to drop project from team")

		return DeleteProjectFromTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop project from team"),
		}}, nil
	}

	return DeleteProjectFromTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped project from team"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// ListProjectUsers implements the v1.ServerInterface.
func (a *API) ListProjectUsers(ctx context.Context, request ListProjectUsersRequestObject) (ListProjectUsersResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectUsers404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectUsers").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectUsers500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: record,
	}) {
		return ListProjectUsers404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listProjectUsersSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.ListUsers(
		ctx,
		model.UserProjectParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			ProjectID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListProjectUsers").
			Str("project", record.ID).
			Msg("Failed to load project users")

		return ListProjectUsers500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project users"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]UserProject, len(records))
	for id, record := range records {
		payload[id] = a.convertProjectUser(record)
	}

	return ListProjectUsers200JSONResponse{ProjectUsersResponseJSONResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(record)),
		Users:   payload,
	}}, nil
}

// AttachProjectToUser implements the v1.ServerInterface.
func (a *API) AttachProjectToUser(ctx context.Context, request AttachProjectToUserRequestObject) (AttachProjectToUserResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return AttachProjectToUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachProjectToUser").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return AttachProjectToUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return AttachProjectToUser404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.AttachUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    request.Body.User,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return AttachProjectToUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachProjectToUser412JSONResponse{AlreadyAttachedErrorJSONResponse{
				Message: ToPtr("User is already attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

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

			return AttachProjectToUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachProjectToUser").
			Str("project", record.ID).
			Str("user", request.Body.User).
			Msg("Failed to attach project to user")

		return AttachProjectToUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach project to user"),
		}}, nil
	}

	return AttachProjectToUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached project to user"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitProjectUser implements the v1.ServerInterface.
func (a *API) PermitProjectUser(ctx context.Context, request PermitProjectUserRequestObject) (PermitProjectUserResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return PermitProjectUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitProjectUser").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return PermitProjectUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return PermitProjectUser404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.PermitUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    request.Body.User,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return PermitProjectUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitProjectUser412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("User is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

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

			return PermitProjectUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitProjectUser").
			Str("project", record.ID).
			Str("user", request.Body.User).
			Msg("Failed to update project user perms")

		return PermitProjectUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update project user perms"),
		}}, nil
	}

	return PermitProjectUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated project user perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteProjectFromUser implements the v1.ServerInterface.
func (a *API) DeleteProjectFromUser(ctx context.Context, request DeleteProjectFromUserRequestObject) (DeleteProjectFromUserResponseObject, error) {
	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectFromUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectFromUser").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectFromUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageProject(ctx, projectPermissions{
		User:      current.GetUser(ctx),
		Record:    record,
		OwnerOnly: true,
	}) {
		return DeleteProjectFromUser404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.DropUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    request.Body.User,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return DeleteProjectFromUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteProjectFromUser412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("User is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectFromUser").
			Str("project", record.ID).
			Str("user", request.Body.User).
			Msg("Failed to drop project from user")

		return DeleteProjectFromUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop project from user"),
		}}, nil
	}

	return DeleteProjectFromUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped project from user"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

func (a *API) convertProject(record *model.Project) Project {
	result := Project{
		Id:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertProjectTeam(record *model.TeamProject) TeamProject {
	result := TeamProject{
		ProjectId: record.ProjectID,
		TeamId:    record.TeamID,
		Team:      ToPtr(a.convertTeam(record.Team)),
		Perm:      ToPtr(TeamProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertProjectUser(record *model.UserProject) UserProject {
	result := UserProject{
		ProjectId: record.ProjectID,
		UserId:    record.UserID,
		User:      ToPtr(a.convertUser(record.User)),
		Perm:      ToPtr(UserProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

type projectPermissions struct {
	User      *model.User
	Record    *model.Project
	OwnerOnly bool
}

func (a *API) permitCreateProject(_ context.Context, definition projectPermissions) bool {
	return definition.User != nil
}

func (a *API) permitShowProject(_ context.Context, definition projectPermissions) bool {
	if definition.User == nil {
		return false
	}

	if definition.User.Admin {
		return true
	}

	for _, p := range definition.User.Projects {
		if p.ProjectID == definition.Record.ID {
			return true
		}
	}

	for _, t := range definition.User.Teams {
		for _, p := range t.Team.Projects {
			if p.ProjectID == definition.Record.ID {
				return true
			}
		}
	}

	return false
}

func (a *API) permitManageProject(_ context.Context, definition projectPermissions) bool {
	if definition.User == nil {
		return false
	}

	if definition.User.Admin {
		return true
	}

	for _, p := range definition.User.Projects {
		if p.ProjectID == definition.Record.ID {
			if definition.OwnerOnly {
				if p.Perm == model.OwnerPerm {
					return true
				}

				break
			}

			if p.Perm == model.AdminPerm {
				return true
			}

			break
		}
	}

	for _, t := range definition.User.Teams {
		for _, p := range t.Team.Projects {
			if p.ProjectID == definition.Record.ID {
				if definition.OwnerOnly {
					if p.Perm == model.OwnerPerm {
						return true
					}

					break
				}

				if p.Perm == model.AdminPerm {
					return true
				}

				break
			}
		}
	}

	return false
}

func listProjectsSorting(request ListProjectsRequestObject) (string, string, int64, int64, string) {
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

func listProjectTeamsSorting(request ListProjectTeamsRequestObject) (string, string, int64, int64, string) {
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

func listProjectUsersSorting(request ListProjectUsersRequestObject) (string, string, int64, int64, string) {
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
