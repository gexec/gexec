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

// ListUsers implements the v1.ServerInterface.
func (a *API) ListUsers(ctx context.Context, request ListUsersRequestObject) (ListUsersResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListUsers403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	sort, order, limit, offset, search := listUsersSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.List(
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
			Str("action", "ListUsers").
			Msg("Failed to load users")

		return ListUsers500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load users"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]User, len(records))
	for id, record := range records {
		payload[id] = a.convertUser(record)
	}

	return ListUsers200JSONResponse{UsersResponseJSONResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Users:  payload,
	}}, nil
}

// ShowUser implements the v1.ServerInterface.
func (a *API) ShowUser(ctx context.Context, request ShowUserRequestObject) (ShowUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ShowUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Show(
		ctx,
		request.UserId,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return ShowUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowUser").
			Msg("Failed to load user")

		return ShowUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return ShowUser200JSONResponse{UserResponseJSONResponse(
		a.convertUser(record),
	)}, nil
}

// CreateUser implements the v1.ServerInterface.
func (a *API) CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return CreateUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record := &model.User{}

	if request.Body.Username != nil {
		record.Username = FromPtr(request.Body.Username)
	}

	if request.Body.Password != nil {
		record.Password = FromPtr(request.Body.Password)
	}

	if request.Body.Email != nil {
		record.Email = FromPtr(request.Body.Email)
	}

	if request.Body.Fullname != nil {
		record.Fullname = FromPtr(request.Body.Fullname)
	}

	if request.Body.Admin != nil {
		record.Admin = FromPtr(request.Body.Admin)
	}

	if request.Body.Active != nil {
		record.Active = FromPtr(request.Body.Active)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Create(
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

			return CreateUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateUser").
			Msg("Failed to create user")

		return CreateUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateUser200JSONResponse{UserResponseJSONResponse(
		a.convertUser(record),
	)}, nil
}

// UpdateUser implements the v1.ServerInterface.
func (a *API) UpdateUser(ctx context.Context, request UpdateUserRequestObject) (UpdateUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return UpdateUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Show(
		ctx,
		request.UserId,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return UpdateUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateUser").
			Msg("Failed to load user")

		return UpdateUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if request.Body.Username != nil {
		record.Username = FromPtr(request.Body.Username)
	}

	if request.Body.Password != nil {
		record.Password = FromPtr(request.Body.Password)
	}

	if request.Body.Email != nil {
		record.Email = FromPtr(request.Body.Email)
	}

	if request.Body.Fullname != nil {
		record.Fullname = FromPtr(request.Body.Fullname)
	}

	if request.Body.Admin != nil {
		record.Admin = FromPtr(request.Body.Admin)
	}

	if request.Body.Active != nil {
		record.Active = FromPtr(request.Body.Active)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Update(
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

			return UpdateUser422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateUser").
			Msg("Failed to update user")

		return UpdateUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateUser200JSONResponse{UserResponseJSONResponse(
		a.convertUser(record),
	)}, nil
}

// DeleteUser implements the v1.ServerInterface.
func (a *API) DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteUser403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Show(
		ctx,
		request.UserId,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return DeleteUser404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteUser").
			Msg("Failed to load user")

		return DeleteUser500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Delete(
		ctx,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeleteUser").
			Msg("Failed to delete user")

		return DeleteUser400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete user"),
		}}, nil
	}

	return DeleteUser200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted user"),
	}}, nil
}

// ListUserTeams implements the v1.ServerInterface.
func (a *API) ListUserTeams(ctx context.Context, request ListUserTeamsRequestObject) (ListUserTeamsResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListUserTeams403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Show(
		ctx,
		request.UserId,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return ListUserTeams404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListUserTeams").
			Msg("Failed to load user")

		return ListUserTeams500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	sort, order, limit, offset, search := listUserTeamsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.ListTeams(
		ctx,
		model.UserTeamParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			UserID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListUserTeams").
			Msg("Failed to load user teams")

		return ListUserTeams500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user teams"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]UserTeam, len(records))
	for id, record := range records {
		payload[id] = a.convertUserTeam(record)
	}

	return ListUserTeams200JSONResponse{UserTeamsResponseJSONResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		User:   ToPtr(a.convertUser(record)),
		Teams:  payload,
	}}, nil
}

// AttachUserToTeam implements the v1.ServerInterface.
func (a *API) AttachUserToTeam(ctx context.Context, request AttachUserToTeamRequestObject) (AttachUserToTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return AttachUserToTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.AttachTeam(
		ctx,
		model.UserTeamParams{
			UserID: request.UserId,
			TeamID: request.Body.Team,
			Perm:   request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return AttachUserToTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrTeamNotFound) {
			return AttachUserToTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachUserToTeam412JSONResponse{AlreadyAttachedErrorJSONResponse{
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

			return AttachUserToTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachUserToTeam").
			Msg("Failed to attach user to team")

		return AttachUserToTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach user to team"),
		}}, nil
	}

	return AttachUserToTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached user to team"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitUserTeam implements the v1.ServerInterface.
func (a *API) PermitUserTeam(ctx context.Context, request PermitUserTeamRequestObject) (PermitUserTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return PermitUserTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.PermitTeam(
		ctx,
		model.UserTeamParams{
			UserID: request.UserId,
			TeamID: request.Body.Team,
			Perm:   request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return PermitUserTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrTeamNotFound) {
			return PermitUserTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitUserTeam412JSONResponse{NotAttachedErrorJSONResponse{
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

			return PermitUserTeam422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user team"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitUserTeam").
			Msg("Failed to update user team perms")

		return PermitUserTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update user team perms"),
		}}, nil
	}

	return PermitUserTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated user team perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteUserFromTeam implements the v1.ServerInterface.
func (a *API) DeleteUserFromTeam(ctx context.Context, request DeleteUserFromTeamRequestObject) (DeleteUserFromTeamResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteUserFromTeam403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.DropTeam(
		ctx,
		model.UserTeamParams{
			UserID: request.UserId,
			TeamID: request.Body.Team,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return DeleteUserFromTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrTeamNotFound) {
			return DeleteUserFromTeam404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find team"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteUserFromTeam412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Team is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteUserFromTeam").
			Msg("Failed to drop user from team")

		return DeleteUserFromTeam500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop user from team"),
		}}, nil
	}

	return DeleteUserFromTeam200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped user from team"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// ListUserProjects implements the v1.ServerInterface.
func (a *API) ListUserProjects(ctx context.Context, request ListUserProjectsRequestObject) (ListUserProjectsResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return ListUserProjects403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Show(
		ctx,
		request.UserId,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return ListUserProjects404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListUserProjects").
			Msg("Failed to load user")

		return ListUserProjects500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	sort, order, limit, offset, search := listUserProjectsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.ListProjects(
		ctx,
		model.UserProjectParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			UserID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("action", "ListUserProjects").
			Msg("Failed to load user projects")

		return ListUserProjects500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load user projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]UserProject, len(records))
	for id, record := range records {
		payload[id] = a.convertUserProject(record)
	}

	return ListUserProjects200JSONResponse{UserProjectsResponseJSONResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		User:     ToPtr(a.convertUser(record)),
		Projects: payload,
	}}, nil
}

// AttachUserToProject implements the v1.ServerInterface.
func (a *API) AttachUserToProject(ctx context.Context, request AttachUserToProjectRequestObject) (AttachUserToProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return AttachUserToProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.AttachProject(
		ctx,
		model.UserProjectParams{
			UserID:    request.UserId,
			ProjectID: request.Body.Project,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return AttachUserToProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return AttachUserToProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			return AttachUserToProject412JSONResponse{AlreadyAttachedErrorJSONResponse{
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

			return AttachUserToProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "AttachUserToProject").
			Msg("Failed to attach user to project")

		return AttachUserToProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach user to project"),
		}}, nil
	}

	return AttachUserToProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully attached user to project"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// PermitUserProject implements the v1.ServerInterface.
func (a *API) PermitUserProject(ctx context.Context, request PermitUserProjectRequestObject) (PermitUserProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return PermitUserProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.PermitProject(
		ctx,
		model.UserProjectParams{
			UserID:    request.UserId,
			ProjectID: request.Body.Project,
			Perm:      request.Body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return PermitUserProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return PermitUserProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return PermitUserProject412JSONResponse{NotAttachedErrorJSONResponse{
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

			return PermitUserProject422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user project"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "PermitUserProject").
			Msg("Failed to update user project perms")

		return PermitUserProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update user project perms"),
		}}, nil
	}

	return PermitUserProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully updated user project perms"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

// DeleteUserFromProject implements the v1.ServerInterface.
func (a *API) DeleteUserFromProject(ctx context.Context, request DeleteUserFromProjectRequestObject) (DeleteUserFromProjectResponseObject, error) {
	if !a.permitAdmin(ctx, current.GetUser(ctx)) {
		return DeleteUserFromProject403JSONResponse{NotAuthorizedErrorJSONResponse{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.DropProject(
		ctx,
		model.UserProjectParams{
			UserID:    request.UserId,
			ProjectID: request.Body.Project,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return DeleteUserFromProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteUserFromProject404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		if errors.Is(err, store.ErrNotAssigned) {
			return DeleteUserFromProject412JSONResponse{NotAttachedErrorJSONResponse{
				Message: ToPtr("Project is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteUserFromProject").
			Msg("Failed to drop user from project")

		return DeleteUserFromProject500JSONResponse{InternalServerErrorJSONResponse{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop user from project"),
		}}, nil
	}

	return DeleteUserFromProject200JSONResponse{SuccessMessageJSONResponse{
		Message: ToPtr("Successfully dropped user from project"),
		Status:  ToPtr(http.StatusOK),
	}}, nil
}

func (a *API) convertUser(record *model.User) User {
	result := User{
		Id:        ToPtr(record.ID),
		Username:  ToPtr(record.Username),
		Email:     ToPtr(record.Email),
		Fullname:  ToPtr(record.Fullname),
		Profile:   ToPtr(gravatarFor(record.Email)),
		Active:    ToPtr(record.Active),
		Admin:     ToPtr(record.Admin),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	auths := make([]UserAuth, 0)

	for _, auth := range record.Auths {
		auths = append(
			auths,
			a.convertUserAuth(auth),
		)
	}

	result.Auths = ToPtr(auths)

	return result
}

func (a *API) convertUserAuth(record *model.UserAuth) UserAuth {
	result := UserAuth{
		Provider:  ToPtr(record.Provider),
		Ref:       ToPtr(record.Ref),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertUserTeam(record *model.UserTeam) UserTeam {
	result := UserTeam{
		UserId:    record.UserID,
		TeamId:    record.TeamID,
		Team:      ToPtr(a.convertTeam(record.Team)),
		Perm:      ToPtr(UserTeamPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertUserProject(record *model.UserProject) UserProject {
	result := UserProject{
		UserId:    record.UserID,
		ProjectId: record.ProjectID,
		Project:   ToPtr(a.convertProject(record.Project)),
		Perm:      ToPtr(UserProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func listUsersSorting(request ListUsersRequestObject) (string, string, int64, int64, string) {
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

func listUserTeamsSorting(request ListUserTeamsRequestObject) (string, string, int64, int64, string) {
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

func listUserProjectsSorting(request ListUserProjectsRequestObject) (string, string, int64, int64, string) {
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
