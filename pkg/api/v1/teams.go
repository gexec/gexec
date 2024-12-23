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

// ListTeams implements the v1.ServerInterface.
func (a *API) ListTeams(ctx context.Context, request ListTeamsRequestObject) (ListTeamsResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListTeams403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	sort, order, limit, offset, search := listTeamsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.List(
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
			Str("action", "ListTeams").
			Msg("Failed to load teams")

		return ListTeams500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load teams"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Team, len(records))
	for id, record := range records {
		payload[id] = a.convertTeam(record)
	}

	return ListTeams200JSONResponse{TeamsResponseJSONResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Teams:  payload,
	}}, nil
}

// ShowTeam implements the v1.ServerInterface.
func (a *API) ShowTeam(ctx context.Context, request ShowTeamRequestObject) (ShowTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ShowTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Show(
		ctx,
		request.TeamId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return ShowTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowTeam").
			Msg("Failed to load team")

		return ShowTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return ShowTeam200JSONResponse{TeamResponseJSONResponse(
		a.convertTeam(record),
	)}, nil
}

// CreateTeam implements the v1.ServerInterface.
func (a *API) CreateTeam(ctx context.Context, request CreateTeamRequestObject) (CreateTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return CreateTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record := &model.Team{
		Name: request.Body.Name,
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Create(
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

			return CreateTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateTeam").
			Msg("Failed to create team")

		return CreateTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateTeam200JSONResponse{TeamResponseJSONResponse(
		a.convertTeam(record),
	)}, nil
}

// UpdateTeam implements the v1.ServerInterface.
func (a *API) UpdateTeam(ctx context.Context, request UpdateTeamRequestObject) (UpdateTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return UpdateTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Show(
		ctx,
		request.TeamId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return UpdateTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateTeam").
			Msg("Failed to load team")

		return UpdateTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team"),
			Status:  ToPtr(http.StatusInternalServerError),
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
	).Teams.Update(
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

			return UpdateTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateTeam").
			Msg("Failed to update team")

		return UpdateTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateTeam200JSONResponse{TeamResponseJSONResponse(
		a.convertTeam(record),
	)}, nil
}

// DeleteTeam implements the v1.ServerInterface.
func (a *API) DeleteTeam(ctx context.Context, request DeleteTeamRequestObject) (DeleteTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Show(
		ctx,
		request.TeamId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return DeleteTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteTeam").
			Msg("Failed to load team")

		return DeleteTeam500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Delete(
		ctx,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeleteTeam").
			Msg("Failed to delete team")

		return DeleteTeam400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete team"),
		}}, nil
	}

	return DeleteTeam200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted team"),
	}}, nil
}

// ListTeamUsers implements the v1.ServerInterface.
func (a *API) ListTeamUsers(ctx context.Context, request ListTeamUsersRequestObject) (ListTeamUsersResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListTeamUsers403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Show(
		ctx,
		request.TeamId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return ListTeamUsers404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListTeamUsers").
			Msg("Failed to load team")

		return ListTeamUsers500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	sort, order, limit, offset, search := listTeamUsersSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.ListUsers(
		ctx,
		model.UserTeamParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			TeamID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListTeamUsers").
			Msg("Failed to load team users")

		return ListTeamUsers500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team users"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]UserTeam, len(records))
	for id, record := range records {
		payload[id] = a.convertTeamUser(record)
	}

	return ListTeamUsers200JSONResponse{TeamUsersResponseJSONResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Team:   ToPtr(a.convertTeam(record)),
		Users:  payload,
	}}, nil
}

// AttachTeamToUser implements the v1.ServerInterface.
func (a *API) AttachTeamToUser(ctx context.Context, request AttachTeamToUserRequestObject) (AttachTeamToUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return AttachTeamToUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.AttachUser(
		ctx,
		model.UserTeamParams{
			TeamID: request.TeamId,
			UserID: request.Body.User,
			Perm:   request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return AttachTeamToUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrUserNotFound) {
			return AttachTeamToUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachTeamToUser412JSONResponse{AlreadyAttachedErrorJSONResponse{
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

			return AttachTeamToUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachTeamToUser").
			Msg("Failed to attach team to user")

		return AttachTeamToUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach team to user"),
		}}, nil
	}

	return AttachTeamToUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached team to user"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitTeamUser implements the v1.ServerInterface.
func (a *API) PermitTeamUser(ctx context.Context, request PermitTeamUserRequestObject) (PermitTeamUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return PermitTeamUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.PermitUser(
		ctx,
		model.UserTeamParams{
			TeamID: request.TeamId,
			UserID: request.Body.User,
			Perm:   request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return PermitTeamUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrUserNotFound) {
			return PermitTeamUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitTeamUser412JSONResponse{NotAttachedErrorJSONResponse{
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

			return PermitTeamUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitTeamUser").
			Msg("Failed to update team user perms")

		return PermitTeamUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update team user perms"),
		}}, nil
	}

	return PermitTeamUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated team user perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteTeamFromUser implements the v1.ServerInterface.
func (a *API) DeleteTeamFromUser(ctx context.Context, request DeleteTeamFromUserRequestObject) (DeleteTeamFromUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteTeamFromUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.DropUser(
		ctx,
		model.UserTeamParams{
			TeamID: request.TeamId,
			UserID: request.Body.User,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return DeleteTeamFromUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrUserNotFound) {
			return DeleteTeamFromUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteTeamFromUser412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("User is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteTeamFromUser").
			Msg("Failed to drop team from user")

		return DeleteTeamFromUser500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop team from user"),
		}}, nil
	}

	return DeleteTeamFromUser200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped team from user"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// ListTeamProjects implements the v1.ServerInterface.
func (a *API) ListTeamProjects(ctx context.Context, request ListTeamProjectsRequestObject) (ListTeamProjectsResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListTeamProjects403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.Show(
		ctx,
		request.TeamId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return ListTeamProjects404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListTeamProjects").
			Msg("Failed to load team")

		return ListTeamProjects500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	sort, order, limit, offset, search := listTeamProjectsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.ListProjects(
		ctx,
		model.TeamProjectParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			TeamID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListTeamProjects").
			Msg("Failed to load team projects")

		return ListTeamProjects500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load team projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]TeamProject, len(records))
	for id, record := range records {
		payload[id] = a.convertTeamProject(record)
	}

	return ListTeamProjects200JSONResponse{TeamProjectsResponseJSONResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		Team:     ToPtr(a.convertTeam(record)),
		Projects: payload,
	}}, nil
}

// AttachTeamToProject implements the v1.ServerInterface.
func (a *API) AttachTeamToProject(ctx context.Context, request AttachTeamToProjectRequestObject) (AttachTeamToProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return AttachTeamToProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.AttachProject(
		ctx,
		model.TeamProjectParams{
			TeamID:    request.TeamId,
			ProjectID: request.Body.Project,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return AttachTeamToProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return AttachTeamToProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachTeamToProject412JSONResponse{AlreadyAttachedErrorJSONResponse{
				Message: ToPtr("Project is already attached"),
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

			return AttachTeamToProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachTeamToProject").
			Msg("Failed to attach team to project")

		return AttachTeamToProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach team to project"),
		}}, nil
	}

	return AttachTeamToProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached team to project"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitTeamProject implements the v1.ServerInterface.
func (a *API) PermitTeamProject(ctx context.Context, request PermitTeamProjectRequestObject) (PermitTeamProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return PermitTeamProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.PermitProject(
		ctx,
		model.TeamProjectParams{
			TeamID:    request.TeamId,
			ProjectID: request.Body.Project,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return PermitTeamProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return PermitTeamProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitTeamProject412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Project is not attached"),
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

			return PermitTeamProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate team project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitTeamProject").
			Msg("Failed to update team project perms")

		return PermitTeamProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update team project perms"),
		}}, nil
	}

	return PermitTeamProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated team project perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteTeamFromProject implements the v1.ServerInterface.
func (a *API) DeleteTeamFromProject(ctx context.Context, request DeleteTeamFromProjectRequestObject) (DeleteTeamFromProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteTeamFromProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Teams.DropProject(
		ctx,
		model.TeamProjectParams{
			TeamID:    request.TeamId,
			ProjectID: request.Body.Project,
		},
	); err != nil {
		if errors.Is(err, store.ErrTeamNotFound) {
			return DeleteTeamFromProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteTeamFromProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteTeamFromProject412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Project is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteTeamFromProject").
			Msg("Failed to drop team from project")

		return DeleteTeamFromProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop team from project"),
		}}, nil
	}

	return DeleteTeamFromProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped team from project"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

func (a *API) convertTeam(record *model.Team) Team {
	result := Team{
		Id:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	auths := make([]TeamAuth, 0)

	for _, auth := range record.Auths {
		auths = append(
			auths,
			a.convertTeamAuth(auth),
		)
	}

	result.Auths = ToPtr(auths)

	return result
}

func (a *API) convertTeamAuth(record *model.TeamAuth) TeamAuth {
	result := TeamAuth{
		Provider:  ToPtr(record.Provider),
		Ref:       ToPtr(record.Ref),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertTeamUser(record *model.UserTeam) UserTeam {
	result := UserTeam{
		UserId:    record.UserID,
		User:      ToPtr(a.convertUser(record.User)),
		TeamId:    record.TeamID,
		Perm:      ToPtr(UserTeamPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertTeamProject(record *model.TeamProject) TeamProject {
	result := TeamProject{
		TeamId:    record.TeamID,
		ProjectId: record.ProjectID,
		Project:   ToPtr(a.convertProject(record.Project)),
		Perm:      ToPtr(TeamProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func listTeamsSorting(request ListTeamsRequestObject) (string, string, int64, int64, string) {
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

func listTeamUsersSorting(request ListTeamUsersRequestObject) (string, string, int64, int64, string) {
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

func listTeamProjectsSorting(request ListTeamProjectsRequestObject) (string, string, int64, int64, string) {
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
