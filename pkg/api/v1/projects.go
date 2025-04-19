package v1

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/store"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
)

// ListProjects implements the v1.ServerInterface.
func (a *API) ListProjects(w http.ResponseWriter, r *http.Request, params ListProjectsParams) {
	ctx := r.Context()
	sort, order, limit, offset, search := listProjectsSorting(params)

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
		slog.Error(
			"Failed to load projects",
			slog.Any("error", err),
			slog.String("action", "ListProjects"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load projects"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Project, len(records))
	for id, record := range records {
		payload[id] = a.convertProject(record)
	}

	render.JSON(w, r, ProjectsResponse{
		Total:    count,
		Limit:    limit,
		Offset:   offset,
		Projects: payload,
	})
}

// ShowProject implements the v1.ServerInterface.
func (a *API) ShowProject(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)

	render.JSON(w, r, ProjectResponse(
		a.convertProject(record),
	))
}

// CreateProject implements the v1.ServerInterface.
func (a *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &CreateProjectBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("action", "CreateProject"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Project{}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Demo != nil {
		record.Demo = FromPtr(body.Demo)
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

			a.RenderNotify(w, r, Notification{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to create project",
			slog.Any("error", err),
			slog.String("action", "CreateProject"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create project"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectResponse(
		a.convertProject(record),
	))
}

// UpdateProject implements the v1.ServerInterface.
func (a *API) UpdateProject(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &UpdateProjectBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "UpdateProject"),
		)

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

			a.RenderNotify(w, r, Notification{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project"),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update project",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "UpdateProject"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update project"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectResponse(
		a.convertProject(record),
	))
}

// DeleteProject implements the v1.ServerInterface.
func (a *API) DeleteProject(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Delete(
		ctx,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete project",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "DeleteProject"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete project"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted project"),
		Status:  ToPtr(http.StatusOK),
	})
}

// ListProjectGroups implements the v1.ServerInterface.
func (a *API) ListProjectGroups(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectGroupsParams) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listProjectGroupsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.ListGroups(
		ctx,
		model.GroupProjectParams{
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
		slog.Error(
			"Failed to load project groups",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "ListProjectGroups"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load project groups"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]GroupProject, len(records))
	for id, record := range records {
		payload[id] = a.convertProjectGroup(record)
	}

	render.JSON(w, r, ProjectGroupsResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(record)),
		Groups:  payload,
	})
}

// AttachProjectToGroup implements the v1.ServerInterface.
func (a *API) AttachProjectToGroup(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "AttachProjectToGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.AttachGroup(
		ctx,
		model.GroupProjectParams{
			ProjectID: record.ID,
			GroupID:   body.Group,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
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
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project group"),
				Errors:  ToPtr(errors),
			})
		}

		slog.Error(
			"Failed to attach project to group",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("group", body.Group),
			slog.String("action", "AttachProjectToGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to attach project to group"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached project to group"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitProjectGroup implements the v1.ServerInterface.
func (a *API) PermitProjectGroup(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "PermitProjectGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.PermitGroup(
		ctx,
		model.GroupProjectParams{
			ProjectID: record.ID,
			GroupID:   body.Group,
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
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate project group"),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update project group perms",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("group", body.Group),
			slog.String("action", "PermitProjectGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Successfully updated project group perms"),
			Status:  ToPtr(http.StatusOK),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated project group perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteProjectFromGroup implements the v1.ServerInterface.
func (a *API) DeleteProjectFromGroup(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectGroupPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "DeleteProjectFromGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.DropGroup(
		ctx,
		model.GroupProjectParams{
			ProjectID: record.ID,
			GroupID:   body.Group,
		},
	); err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
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

		slog.Error(
			"Failed to drop project from group",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("group", body.Group),
			slog.String("action", "DeleteProjectFromGroup"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to drop project from group"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped project from group"),
		Status:  ToPtr(http.StatusOK),
	})
}

// ListProjectUsers implements the v1.ServerInterface.
func (a *API) ListProjectUsers(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectUsersParams) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listProjectUsersSorting(params)

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
		slog.Error(
			"Failed to load project users",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "ListProjectUsers"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load project users"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]UserProject, len(records))
	for id, record := range records {
		payload[id] = a.convertProjectUser(record)
	}

	render.JSON(w, r, ProjectUsersResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(record)),
		Users:   payload,
	})
}

// AttachProjectToUser implements the v1.ServerInterface.
func (a *API) AttachProjectToUser(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "AttachProjectToUser"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.AttachUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    body.User,
			Perm:      body.Perm,
		},
	); err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
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
				Message: ToPtr("Failed to validate project user"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to attach project to user",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("user", body.User),
			slog.String("action", "AttachProjectToUser"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to attach project to user"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully attached project to user"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PermitProjectUser implements the v1.ServerInterface.
func (a *API) PermitProjectUser(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "PermitProjectUser"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.PermitUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    body.User,
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
				Message: ToPtr("Failed to validate project user"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update project user perms",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("user", body.User),
			slog.String("action", "PermitProjectUser"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update project user perms"),
			Status:  ToPtr(http.StatusUnprocessableEntity),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully updated project user perms"),
		Status:  ToPtr(http.StatusOK),
	})
}

// DeleteProjectFromUser implements the v1.ServerInterface.
func (a *API) DeleteProjectFromUser(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	record := a.ProjectFromContext(ctx)
	body := &ProjectUserPermBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("action", "DeleteProjectFromUser"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.DropUser(
		ctx,
		model.UserProjectParams{
			ProjectID: record.ID,
			UserID:    body.User,
		},
	); err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
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

		slog.Error(
			"Failed to drop project from user",
			slog.Any("error", err),
			slog.String("project", record.ID),
			slog.String("user", body.User),
			slog.String("action", "DeleteProjectFromUser"),
		)

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusUnprocessableEntity),
			Message: ToPtr("Failed to drop project from user"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully dropped project from user"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertProject(record *model.Project) Project {
	result := Project{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertProjectGroup(record *model.GroupProject) GroupProject {
	result := GroupProject{
		ProjectID: record.ProjectID,
		GroupID:   record.GroupID,
		Group:     ToPtr(a.convertGroup(record.Group)),
		Perm:      ToPtr(GroupProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

func (a *API) convertProjectUser(record *model.UserProject) UserProject {
	result := UserProject{
		ProjectID: record.ProjectID,
		UserID:    record.UserID,
		User:      ToPtr(a.convertUser(record.User)),
		Perm:      ToPtr(UserProjectPerm(record.Perm)),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

// AllowShowProject defines a middleware to check permissions.
func (a *API) AllowShowProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		principal := current.GetUser(ctx)

		if principal == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		if principal.Admin {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		project := a.ProjectFromContext(ctx)

		if project == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		for _, p := range principal.Projects {
			if p.ProjectID == project.ID {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		for _, t := range principal.Groups {
			for _, p := range t.Group.Projects {
				if p.ProjectID == project.ID {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		render.JSON(w, r, Notification{
			Message: ToPtr("You are not allowd to access the resource"),
			Status:  ToPtr(http.StatusForbidden),
		})
	})
}

// AllowOwnerProject defines a middleware to check permissions.
func (a *API) AllowOwnerProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		principal := current.GetUser(ctx)

		if principal == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		if principal.Admin {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		project := a.ProjectFromContext(ctx)

		if project == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		for _, p := range principal.Projects {
			if p.ProjectID == project.ID &&
				p.Perm == model.UserProjectOwnerPerm {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		for _, t := range principal.Groups {
			for _, p := range t.Group.Projects {
				if p.ProjectID == project.ID &&
					p.Perm == model.GroupProjectOwnerPerm {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		render.JSON(w, r, Notification{
			Message: ToPtr("You are not allowd to access the resource"),
			Status:  ToPtr(http.StatusForbidden),
		})
	})
}

// AllowManageProject defines a middleware to check permissions.
func (a *API) AllowManageProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		principal := current.GetUser(ctx)

		if principal == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		if principal.Admin {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		project := a.ProjectFromContext(ctx)

		if project == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("You are not allowd to access the resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		for _, p := range principal.Projects {
			if p.ProjectID == project.ID &&
				(p.Perm == model.UserProjectAdminPerm ||
					p.Perm == model.UserProjectOwnerPerm) {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		for _, t := range principal.Groups {
			for _, p := range t.Group.Projects {
				if p.ProjectID == project.ID &&
					(p.Perm == model.GroupProjectAdminPerm ||
						p.Perm == model.GroupProjectOwnerPerm) {
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		render.JSON(w, r, Notification{
			Message: ToPtr("You are not allowd to access the resource"),
			Status:  ToPtr(http.StatusForbidden),
		})
	})
}

func listProjectsSorting(request ListProjectsParams) (string, string, int64, int64, string) {
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

func listProjectGroupsSorting(request ListProjectGroupsParams) (string, string, int64, int64, string) {
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

func listProjectUsersSorting(request ListProjectUsersParams) (string, string, int64, int64, string) {
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
