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

// ListProjectTemplates implements the v1.ServerInterface.
func (a *API) ListProjectTemplates(ctx context.Context, request ListProjectTemplatesRequestObject) (ListProjectTemplatesResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectTemplates404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectTemplates").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectTemplates500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectTemplates404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listTemplatesSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.List(
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
			Str("action", "ListProjectTemplates").
			Str("project", parent.ID).
			Msg("Failed to load templates")

		return ListProjectTemplates500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load templates"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Template, len(records))
	for id, record := range records {
		payload[id] = a.convertTemplate(record)
	}

	return ListProjectTemplates200JSONResponse{ProjectTemplatesResponseJSONResponse{
		Total:     count,
		Limit:     limit,
		Offset:    offset,
		Project:   ToPtr(a.convertProject(parent)),
		Templates: payload,
	}}, nil
}

// ShowProjectTemplate implements the v1.ServerInterface.
func (a *API) ShowProjectTemplate(ctx context.Context, request ShowProjectTemplateRequestObject) (ShowProjectTemplateResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectTemplate").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Show(
		ctx,
		parent.ID,
		request.TemplateId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTemplateNotFound) {
			return ShowProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectTemplate").
			Str("project", parent.ID).
			Str("template", request.TemplateId).
			Msg("Failed to load template")

		return ShowProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load template"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowTemplate(ctx, templatePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or template"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProjectTemplate200JSONResponse{ProjectTemplateResponseJSONResponse(
		a.convertTemplate(record),
	)}, nil
}

// CreateProjectTemplate implements the v1.ServerInterface.
func (a *API) CreateProjectTemplate(ctx context.Context, request CreateProjectTemplateRequestObject) (CreateProjectTemplateResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectTemplate").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateTemplate(ctx, templatePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Template{
		ProjectID: parent.ID,
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Create(
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

			return CreateProjectTemplate422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate template"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectTemplate").
			Str("project", parent.ID).
			Msg("Failed to create template")

		return CreateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create template"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectTemplate200JSONResponse{ProjectTemplateResponseJSONResponse(
		a.convertTemplate(record),
	)}, nil
}

// UpdateProjectTemplate implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplate(ctx context.Context, request UpdateProjectTemplateRequestObject) (UpdateProjectTemplateResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Show(
		ctx,
		parent.ID,
		request.TemplateId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTemplateNotFound) {
			return UpdateProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", parent.ID).
			Str("template", request.TemplateId).
			Msg("Failed to load template")

		return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load template"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageTemplate(ctx, templatePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or template"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	// TODO
	// if request.Body.Dummy != nil {
	// 	record.Dummy = FromPtr(request.Body.Dummy)
	// }

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Update(
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

			return UpdateProjectTemplate422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate template"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", parent.ID).
			Str("template", record.ID).
			Msg("Failed to update template")

		return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update template"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectTemplate200JSONResponse{ProjectTemplateResponseJSONResponse(
		a.convertTemplate(record),
	)}, nil
}

// DeleteProjectTemplate implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplate(ctx context.Context, request DeleteProjectTemplateRequestObject) (DeleteProjectTemplateResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectTemplate").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Show(
		ctx,
		parent.ID,
		request.TemplateId,
	)

	if err != nil {
		if errors.Is(err, store.ErrTemplateNotFound) {
			return DeleteProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or template"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectTemplate").
			Str("project", parent.ID).
			Str("template", request.TemplateId).
			Msg("Failed to load template")

		return DeleteProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load template"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageTemplate(ctx, templatePermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectTemplate404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or template"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectTemplate").
			Str("project", parent.ID).
			Str("template", record.ID).
			Msg("Failed to delete template")

		return DeleteProjectTemplate400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete template"),
		}}, nil
	}

	return DeleteProjectTemplate200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted template"),
	}}, nil
}

func (a *API) convertTemplate(record *model.Template) Template {
	result := Template{
		Id:        ToPtr(record.ID),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

type templatePermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Template
	OwnerOnly bool
}

func (a *API) permitCreateTemplate(ctx context.Context, definition templatePermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowTemplate(ctx context.Context, definition templatePermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageTemplate(ctx context.Context, definition templatePermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listTemplatesSorting(request ListProjectTemplatesRequestObject) (string, string, int64, int64, string) {
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
