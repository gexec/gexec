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

// ListGroups implements the v1.ServerInterface.
func (a *API) ListGroups(w http.ResponseWriter, r *http.Request, params ListGroupsParams) {
	ctx := r.Context()
	sort, order, limit, offset, search := listGroupsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.List(
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
			Str("action", "ListGroups").
			Msg("Failed to load groups")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load groups"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Group, len(records))
	for id, record := range records {
		payload[id] = a.convertGroup(record)
	}

	render.JSON(w, r, GroupsResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Groups: payload,
	})
}

// ShowGroup implements the v1.ServerInterface.
func (a *API) ShowGroup(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)

	render.JSON(w, r, GroupResponse(
		a.convertGroup(record),
	))
}

// CreateGroup implements the v1.ServerInterface.
func (a *API) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &CreateGroupBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateGroup").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Group{}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.Create(
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
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate group"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("action", "CreateGroup").
			Msg("Failed to create group")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create group"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, GroupResponse(
		a.convertGroup(record),
	))
}

// UpdateGroup implements the v1.ServerInterface.
func (a *API) UpdateGroup(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &CreateGroupBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "UpdateGroup").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.Update(
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
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate group"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "UpdateGroup").
			Msg("Failed to update group")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update group"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, GroupResponse(
		a.convertGroup(record),
	))
}

// DeleteGroup implements the v1.ServerInterface.
func (a *API) DeleteGroup(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.Delete(
		ctx,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "DeleteGroup").
			Msg("Failed to delete group")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete group"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted group"),
	})
}

// ListGroupUsers implements the v1.ServerInterface.
func (a *API) ListGroupUsers(w http.ResponseWriter, r *http.Request, _ GroupID, params ListGroupUsersParams) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	sort, order, limit, offset, search := listGroupUsersSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.ListUsers(
		ctx,
		model.UserGroupParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			GroupID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "ListGroupUsers").
			Msg("Failed to load group users")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load group users"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]UserGroup, len(records))
	for id, record := range records {
		payload[id] = a.convertGroupUser(record)
	}

	render.JSON(w, r, GroupUsersResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Group:  ToPtr(a.convertGroup(record)),
		Users:  payload,
	})
}

// AttachGroupToUser implements the v1.ServerInterface.
func (a *API) AttachGroupToUser(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "AttachGroupToUser").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.AttachUser(
		ctx,
		model.UserGroupParams{
			GroupID: record.ID,
			UserID:  body.User,
			Perm:    body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrAlreadyAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("User is already attached"),
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
				Message: ToPtr("Failed to validate group user"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("user", body.User).
			Str("action", "AttachGroupToUser").
			Msg("Failed to attach group to user")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach group to user"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached group to user"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitGroupUser implements the v1.ServerInterface.
func (a *API) PermitGroupUser(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "PermitGroupUser").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.PermitUser(
		ctx,
		model.UserGroupParams{
			GroupID: record.ID,
			UserID:  body.User,
			Perm:    body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("User is not attached"),
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
				Message: ToPtr("Failed to validate group user"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("user", body.User).
			Str("action", "PermitGroupUser").
			Msg("Failed to update group user perms")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update group user perms"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated group user perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteGroupFromUser implements the v1.ServerInterface.
func (a *API) DeleteGroupFromUser(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "DeleteGroupFromUser").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.DropUser(
		ctx,
		model.UserGroupParams{
			GroupID: record.ID,
			UserID:  body.User,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find user"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		if errors.Is(err, store.ErrNotAssigned) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("User is not attached"),
				Status:  ToPtr(http.StatusPreconditionFailed),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("user", body.User).
			Str("action", "DeleteGroupFromUser").
			Msg("Failed to drop group from user")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop group from user"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped group from user"),
		Status:  ToPtr(http.StatusOK),
	})
}

// ListGroupProjects implements the v1.ServerInterface.
func (a *API) ListGroupProjects(w http.ResponseWriter, r *http.Request, _ GroupID, params ListGroupProjectsParams) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	sort, order, limit, offset, search := listGroupProjectsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.ListProjects(
		ctx,
		model.GroupProjectParams{
			ListParams: model.ListParams{
				Sort:   sort,
				Order:  order,
				Limit:  limit,
				Offset: offset,
				Search: search,
			},
			GroupID: record.ID,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "ListGroupProjects").
			Msg("Failed to load group projects")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load group projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]GroupProject, len(records))
	for id, record := range records {
		payload[id] = a.convertGroupProject(record)
	}

	render.JSON(w, r, GroupProjectsResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		Group:    ToPtr(a.convertGroup(record)),
		Projects: payload,
	})
}

// AttachGroupToProject implements the v1.ServerInterface.
func (a *API) AttachGroupToProject(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "AttachGroupToProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.AttachProject(
		ctx,
		model.GroupProjectParams{
			GroupID:   record.ID,
			ProjectID: body.Project,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
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
				Message: ToPtr("Failed to validate group project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("project", body.Project).
			Str("action", "AttachGroupToProject").
			Msg("Failed to attach group to project")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to attach group to project"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached group to project"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitGroupProject implements the v1.ServerInterface.
func (a *API) PermitGroupProject(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "PermitGroupProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.PermitProject(
		ctx,
		model.GroupProjectParams{
			GroupID:   record.ID,
			ProjectID: body.Project,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
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
				Message: ToPtr("Failed to validate group project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("project", body.Project).
			Str("action", "PermitGroupProject").
			Msg("Failed to update group project perms")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to update group project perms"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated group project perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteGroupFromProject implements the v1.ServerInterface.
func (a *API) DeleteGroupFromProject(w http.ResponseWriter, r *http.Request, _ GroupID) {
	ctx := r.Context()
	record := a.GroupFromContext(ctx)
	body := &GroupProjectPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("group", record.ID).
			Str("action", "DeleteGroupFromProject").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Groups.DropProject(
		ctx,
		model.GroupProjectParams{
			GroupID:   record.ID,
			ProjectID: body.Project,
		},
	); err != nil {
		if errors.Is(err, store.ErrGroupNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find group"),
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
			Str("group", record.ID).
			Str("project", body.Project).
			Str("action", "DeleteGroupFromProject").
			Msg("Failed to drop group from project")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop group from project"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped group from project"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertGroup(record *model.Group) Group {
	result := Group{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertGroupUser(record *model.UserGroup) UserGroup {
	result := UserGroup{
		UserID:    record.UserID,
		User:      ToPtr(a.convertUser(record.User)),
		GroupID:   record.GroupID,
		Perm:      ToPtr(UserGroupPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertGroupProject(record *model.GroupProject) GroupProject {
	result := GroupProject{
		GroupID:   record.GroupID,
		ProjectID: record.ProjectID,
		Project:   ToPtr(a.convertProject(record.Project)),
		Perm:      ToPtr(GroupProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func listGroupsSorting(request ListGroupsParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	if request.Order != nil {
		sort = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}

func listGroupUsersSorting(request ListGroupUsersParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	if request.Order != nil {
		sort = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}

func listGroupProjectsSorting(request ListGroupProjectsParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	if request.Order != nil {
		sort = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}
