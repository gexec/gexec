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
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
			log.Error().
				Err(err).
				Str("action", "ListProjectTemplates").
				Str("project", parent.ID).
				Msg("Failed to decrypt secrets")

			return ListProjectTemplates500JSONResponse{InternalServerErrorJSONResponse{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			}}, nil
		}

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

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
		log.Error().
			Err(err).
			Str("action", "ShowProjectTemplate").
			Str("project", parent.ID).
			Str("template", record.ID).
			Msg("Failed to decrypt secrets")

		return ShowProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
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

	if request.Body.Executor != nil {
		record.Executor = FromPtr(request.Body.Executor)
	}

	if request.Body.RepositoryId != nil {
		record.RepositoryID = FromPtr(request.Body.RepositoryId)
	}

	if request.Body.InventoryId != nil {
		record.InventoryID = FromPtr(request.Body.InventoryId)
	}

	if request.Body.EnvironmentId != nil {
		record.EnvironmentID = FromPtr(request.Body.EnvironmentId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Description != nil {
		record.Description = FromPtr(request.Body.Description)
	}

	if request.Body.Playbook != nil {
		record.Playbook = FromPtr(request.Body.Playbook)
	}

	if request.Body.Arguments != nil {
		record.Arguments = FromPtr(request.Body.Arguments)
	}

	if request.Body.Limit != nil {
		record.Limit = FromPtr(request.Body.Limit)
	}

	if request.Body.Branch != nil {
		record.Branch = FromPtr(request.Body.Branch)
	}

	if request.Body.AllowOverride != nil {
		record.Override = FromPtr(request.Body.AllowOverride)
	}

	if request.Body.Surveys != nil {
		record.Surveys = make([]*model.TemplateSurvey, 0)

		for _, row := range FromPtr(request.Body.Surveys) {
			survey := &model.TemplateSurvey{}

			if row.Kind != nil {
				survey.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				survey.Name = FromPtr(row.Name)
			}

			if row.Title != nil {
				survey.Title = FromPtr(row.Title)
			}

			if row.Description != nil {
				survey.Description = FromPtr(row.Description)
			}

			if row.Required != nil {
				survey.Required = FromPtr(row.Required)
			}

			if row.Values != nil {
				survey.Values = make([]*model.TemplateValue, 0)

				for _, val := range FromPtr(row.Values) {
					value := &model.TemplateValue{}

					if val.Name != nil {
						value.Name = FromPtr(val.Name)
					}

					if val.Value != nil {
						value.Value = FromPtr(val.Value)
					}

					survey.Values = append(survey.Values, value)
				}
			}

			record.Surveys = append(record.Surveys, survey)
		}
	}

	if request.Body.Vaults != nil {
		record.Vaults = make([]*model.TemplateVault, 0)

		for _, row := range FromPtr(request.Body.Vaults) {
			vault := &model.TemplateVault{}

			if row.CredentialId != nil {
				vault.CredentialID = FromPtr(row.CredentialId)
			}

			if row.Kind != nil {
				vault.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				vault.Name = FromPtr(row.Name)
			}

			if row.Script != nil {
				vault.Script = FromPtr(row.Script)
			}

			record.Vaults = append(record.Vaults, vault)
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateProjectTemplate").
			Str("project", parent.ID).
			Msg("Failed to encrypt secrets")

		return CreateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

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

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", parent.ID).
			Str("template", record.ID).
			Msg("Failed to decrypt secrets")

		return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if request.Body.RepositoryId != nil {
		record.RepositoryID = FromPtr(request.Body.RepositoryId)
	}

	if request.Body.InventoryId != nil {
		record.InventoryID = FromPtr(request.Body.InventoryId)
	}

	if request.Body.EnvironmentId != nil {
		record.EnvironmentID = FromPtr(request.Body.EnvironmentId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Description != nil {
		record.Description = FromPtr(request.Body.Description)
	}

	if request.Body.Playbook != nil {
		record.Playbook = FromPtr(request.Body.Playbook)
	}

	if request.Body.Arguments != nil {
		record.Arguments = FromPtr(request.Body.Arguments)
	}

	if request.Body.Limit != nil {
		record.Limit = FromPtr(request.Body.Limit)
	}

	if request.Body.Branch != nil {
		record.Branch = FromPtr(request.Body.Branch)
	}

	if request.Body.AllowOverride != nil {
		record.Override = FromPtr(request.Body.AllowOverride)
	}

	if request.Body.Surveys != nil {
		record.Surveys = make([]*model.TemplateSurvey, 0)

		for _, row := range FromPtr(request.Body.Surveys) {
			survey := &model.TemplateSurvey{}

			if row.Id != nil {
				survey.ID = FromPtr(row.Id)
			}

			if row.Kind != nil {
				survey.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				survey.Name = FromPtr(row.Name)
			}

			if row.Title != nil {
				survey.Title = FromPtr(row.Title)
			}

			if row.Description != nil {
				survey.Description = FromPtr(row.Description)
			}

			if row.Required != nil {
				survey.Required = FromPtr(row.Required)
			}

			if row.Values != nil {
				survey.Values = make([]*model.TemplateValue, 0)

				for _, val := range FromPtr(row.Values) {
					value := &model.TemplateValue{}

					if val.Id != nil {
						value.ID = FromPtr(val.Id)
					}

					if val.Name != nil {
						value.Name = FromPtr(val.Name)
					}

					if val.Value != nil {
						value.Value = FromPtr(val.Value)
					}

					survey.Values = append(survey.Values, value)
				}
			}

			record.Surveys = append(record.Surveys, survey)
		}
	}

	if request.Body.Vaults != nil {
		record.Vaults = make([]*model.TemplateVault, 0)

		for _, row := range FromPtr(request.Body.Vaults) {
			vault := &model.TemplateVault{}

			if row.Id != nil {
				vault.ID = FromPtr(row.Id)
			}

			if row.CredentialId != nil {
				vault.CredentialID = FromPtr(row.CredentialId)
			}

			if row.Kind != nil {
				vault.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				vault.Name = FromPtr(row.Name)
			}

			if row.Script != nil {
				vault.Script = FromPtr(row.Script)
			}

			record.Vaults = append(record.Vaults, vault)
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", parent.ID).
			Str("template", record.ID).
			Msg("Failed to encrypt secrets")

		return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

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

// CreateProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) CreateProjectTemplateSurvey(ctx context.Context, request CreateProjectTemplateSurveyRequestObject) (CreateProjectTemplateSurveyResponseObject, error) {
	return CreateProjectTemplateSurvey500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplateSurvey(ctx context.Context, request UpdateProjectTemplateSurveyRequestObject) (UpdateProjectTemplateSurveyResponseObject, error) {
	return UpdateProjectTemplateSurvey500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplateSurvey(ctx context.Context, request DeleteProjectTemplateSurveyRequestObject) (DeleteProjectTemplateSurveyResponseObject, error) {
	return DeleteProjectTemplateSurvey500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectTemplateVault implements the v1.ServerInterface.
func (a *API) CreateProjectTemplateVault(ctx context.Context, request CreateProjectTemplateVaultRequestObject) (CreateProjectTemplateVaultResponseObject, error) {
	return CreateProjectTemplateVault500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectTemplateVault implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplateVault(ctx context.Context, request UpdateProjectTemplateVaultRequestObject) (UpdateProjectTemplateVaultResponseObject, error) {
	return UpdateProjectTemplateVault500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectTemplateVault implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplateVault(ctx context.Context, request DeleteProjectTemplateVaultRequestObject) (DeleteProjectTemplateVaultResponseObject, error) {
	return DeleteProjectTemplateVault500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

func (a *API) convertTemplate(record *model.Template) Template {
	result := Template{
		Id:            ToPtr(record.ID),
		Slug:          ToPtr(record.Slug),
		Name:          ToPtr(record.Name),
		Description:   ToPtr(record.Description),
		Playbook:      ToPtr(record.Playbook),
		Arguments:     ToPtr(record.Arguments),
		Limit:         ToPtr(record.Limit),
		Executor:      ToPtr(record.Executor),
		Branch:        ToPtr(record.Branch),
		AllowOverride: ToPtr(record.Override),
		CreatedAt:     ToPtr(record.CreatedAt),
		UpdatedAt:     ToPtr(record.UpdatedAt),
	}

	if record.Repository != nil {
		result.RepositoryId = ToPtr(record.RepositoryID)

		result.Repository = ToPtr(
			a.convertRepository(
				record.Repository,
			),
		)
	}

	if record.Inventory != nil {
		result.InventoryId = ToPtr(record.InventoryID)

		result.Inventory = ToPtr(
			a.convertInventory(
				record.Inventory,
			),
		)
	}

	if record.Environment != nil {
		result.EnvironmentId = ToPtr(record.EnvironmentID)

		result.Environment = ToPtr(
			a.convertEnvironment(
				record.Environment,
			),
		)
	}

	if len(record.Surveys) > 0 {
		surveys := make([]TemplateSurvey, 0)

		for _, survey := range record.Surveys {
			surveys = append(
				surveys,
				a.convertTemplateSurvey(survey),
			)
		}

		result.Surveys = ToPtr(surveys)
	}

	if len(record.Vaults) > 0 {
		vaults := make([]TemplateVault, 0)

		for _, vault := range record.Vaults {
			vaults = append(
				vaults,
				a.convertTemplateVault(vault),
			)
		}

		result.Vaults = ToPtr(vaults)
	}

	return result
}

func (a *API) convertTemplateSurvey(record *model.TemplateSurvey) TemplateSurvey {
	result := TemplateSurvey{
		Id:          ToPtr(record.ID),
		Name:        ToPtr(record.Name),
		Title:       ToPtr(record.Title),
		Description: ToPtr(record.Description),
		Kind:        ToPtr(TemplateSurveyKind(record.Kind)),
		Required:    ToPtr(record.Required),
	}

	values := make([]TemplateValue, 0)

	for _, value := range record.Values {
		values = append(
			values,
			a.convertTemplateValue(value),
		)
	}

	result.Values = ToPtr(values)

	return result
}

func (a *API) convertTemplateValue(record *model.TemplateValue) TemplateValue {
	result := TemplateValue{
		Id:    ToPtr(record.ID),
		Name:  ToPtr(record.Name),
		Value: ToPtr(record.Value),
	}

	return result
}

func (a *API) convertTemplateVault(record *model.TemplateVault) TemplateVault {
	result := TemplateVault{
		Id:     ToPtr(record.ID),
		Name:   ToPtr(record.Name),
		Kind:   ToPtr(TemplateVaultKind(record.Kind)),
		Script: ToPtr(record.Script),
	}

	if record.Credential != nil {
		result.CredentialId = ToPtr(record.CredentialID)

		result.Credential = ToPtr(
			a.convertCredential(
				record.Credential,
			),
		)
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
