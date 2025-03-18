package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/store"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// ListUsers implements the v1.ServerInterface.
func (a *API) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
	ctx := r.Context()
	sort, order, limit, offset, search := listUsersSorting(params)

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

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load users"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]User, len(records))
	for id, record := range records {
		payload[id] = a.convertUser(record)
	}

	render.JSON(w, r, UsersResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Users:  payload,
	})
}

// ShowUser implements the v1.ServerInterface.
func (a *API) ShowUser(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)

	render.JSON(w, r, UserResponse(
		a.convertUser(record),
	))
}

// CreateUser implements the v1.ServerInterface.
func (a *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &CreateUserBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateUser").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.User{}

	if body.Username != nil {
		record.Username = FromPtr(body.Username)
	}

	if body.Password != nil {
		record.Password = FromPtr(body.Password)
	}

	if body.Email != nil {
		record.Email = FromPtr(body.Email)
	}

	if body.Fullname != nil {
		record.Fullname = FromPtr(body.Fullname)
	}

	if body.Admin != nil {
		record.Admin = FromPtr(body.Admin)
	}

	if body.Active != nil {
		record.Active = FromPtr(body.Active)
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

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate user"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("action", "CreateUser").
			Msg("Failed to create user")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create user"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, UserResponse(
		a.convertUser(record),
	))
}

// UpdateUser implements the v1.ServerInterface.
func (a *API) UpdateUser(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UpdateUserBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "UpdateUser").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if body.Username != nil {
		record.Username = FromPtr(body.Username)
	}

	if body.Password != nil {
		record.Password = FromPtr(body.Password)
	}

	if body.Email != nil {
		record.Email = FromPtr(body.Email)
	}

	if body.Fullname != nil {
		record.Fullname = FromPtr(body.Fullname)
	}

	if body.Admin != nil {
		record.Admin = FromPtr(body.Admin)
	}

	if body.Active != nil {
		record.Active = FromPtr(body.Active)
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

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate user"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "UpdateUser").
			Msg("Failed to update user")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update user"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, UserResponse(
		a.convertUser(record),
	))
}

// DeleteUser implements the v1.ServerInterface.
func (a *API) DeleteUser(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.Delete(
		ctx,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "DeleteUser").
			Msg("Failed to delete user")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete user"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted user"),
		Status:  ToPtr(http.StatusOK),
	})
}

// ListUserGroups implements the v1.ServerInterface.
func (a *API) ListUserGroups(w http.ResponseWriter, r *http.Request, _ UserID, params ListUserGroupsParams) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	sort, order, limit, offset, search := listUserGroupsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.ListGroups(
		ctx,
		model.UserGroupParams{
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
			Str("user", record.ID).
			Str("action", "ListUserGroups").
			Msg("Failed to load user groups")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load user groups"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]UserGroup, len(records))
	for id, record := range records {
		payload[id] = a.convertUserGroup(record)
	}

	render.JSON(w, r, UserGroupsResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		User:   ToPtr(a.convertUser(record)),
		Groups: payload,
	})
}

// AttachUserToGroup implements the v1.ServerInterface.
func (a *API) AttachUserToGroup(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "AttachUserToGroup").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.AttachGroup(
		ctx,
		model.UserGroupParams{
			UserID:  record.ID,
			GroupID: body.Group,
			Perm:    body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Group is already attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
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

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate user group"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("group", body.Group).
			Str("action", "AttachUserToGroup").
			Msg("Failed to attach user to group")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to attach user to group"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached user to group"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitUserGroup implements the v1.ServerInterface.
func (a *API) PermitUserGroup(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "PermitUserGroup").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.PermitGroup(
		ctx,
		model.UserGroupParams{
			UserID:  record.ID,
			GroupID: body.Group,
			Perm:    body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Group is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
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

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate user group"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("group", body.Group).
			Str("action", "PermitUserGroup").
			Msg("Failed to update user group perms")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update user group perms"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated user group perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteUserFromGroup implements the v1.ServerInterface.
func (a *API) DeleteUserFromGroup(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "DeleteUserFromGroup").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.DropGroup(
		ctx,
		model.UserGroupParams{
			UserID:  record.ID,
			GroupID: body.Group,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Group is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("group", body.Group).
			Str("action", "DeleteUserFromGroup").
			Msg("Failed to drop user from group")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop user from group"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped user from group"),
		Status:  ToPtr(http.StatusOK),
	})
}

// ListUserProjects implements the v1.ServerInterface.
func (a *API) ListUserProjects(w http.ResponseWriter, r *http.Request, _ UserID, params ListUserProjectsParams) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	sort, order, limit, offset, search := listUserProjectsSorting(params)

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
			Str("user", record.ID).
			Str("action", "ListUserProjects").
			Msg("Failed to load user projects")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load user projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]UserProject, len(records))
	for id, record := range records {
		payload[id] = a.convertUserProject(record)
	}

	render.JSON(w, r, UserProjectsResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		User:     ToPtr(a.convertUser(record)),
		Projects: payload,
	})
}

// AttachUserToProject implements the v1.ServerInterface.
func (a *API) AttachUserToProject(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "AttachUserToProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.AttachProject(
		ctx,
		model.UserProjectParams{
			UserID:    record.ID,
			ProjectID: body.Project,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Project is already attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
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

			a.RenderNotify(w, r, Notification{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("project", body.Project).
			Str("action", "AttachUserToProject").
			Msg("Failed to attach user to project")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach user to project"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached user to project"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitUserProject implements the v1.ServerInterface.
func (a *API) PermitUserProject(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "PermitUserProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.PermitProject(
		ctx,
		model.UserProjectParams{
			UserID:    record.ID,
			ProjectID: body.Project,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Project is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
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

			a.RenderNotify(w, r, Notification{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate user project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("project", body.Project).
			Str("action", "PermitUserProject").
			Msg("Failed to update user project perms")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update user project perms"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated user project perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteUserFromProject implements the v1.ServerInterface.
func (a *API) DeleteUserFromProject(w http.ResponseWriter, r *http.Request, _ UserID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)
	body := &UserProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("action", "DeleteUserFromProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Users.DropProject(
		ctx,
		model.UserProjectParams{
			UserID:    record.ID,
			ProjectID: body.Project,
		},
	); err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Project is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		log.Error().
			Err(err).
			Str("user", record.ID).
			Str("project", body.Project).
			Str("action", "DeleteUserFromProject").
			Msg("Failed to drop user from project")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop user from project"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped user from project"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertUser(record *model.User) User {
	result := User{
		ID:        ToPtr(record.ID),
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

func (a *API) convertUserGroup(record *model.UserGroup) UserGroup {
	result := UserGroup{
		UserID:    record.UserID,
		GroupID:   record.GroupID,
		Group:     ToPtr(a.convertGroup(record.Group)),
		Perm:      ToPtr(UserGroupPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertUserProject(record *model.UserProject) UserProject {
	result := UserProject{
		UserID:    record.UserID,
		ProjectID: record.ProjectID,
		Project:   ToPtr(a.convertProject(record.Project)),
		Perm:      ToPtr(UserProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func listUsersSorting(request ListUsersParams) (string, string, int64, int64, string) {
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

func listUserGroupsSorting(request ListUserGroupsParams) (string, string, int64, int64, string) {
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

func listUserProjectsSorting(request ListUserProjectsParams) (string, string, int64, int64, string) {
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
