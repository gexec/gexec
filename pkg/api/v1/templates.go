package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// ListProjectTemplates implements the v1.ServerInterface.
func (a *API) ListProjectTemplates(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectTemplatesParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listTemplatesSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.List(
		ctx,
		project.ID,
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
			Str("project", project.ID).
			Str("action", "ListProjectTemplates").
			Msg("Failed to load templates")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load templates"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Template, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			log.Error().
				Err(err).
				Str("project", project.ID).
				Str("action", "ListProjectTemplates").
				Msg("Failed to decrypt secrets")

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertTemplate(record)
	}

	render.JSON(w, r, ProjectTemplatesResponse{
		Total:     count,
		Limit:     limit,
		Offset:    offset,
		Project:   ToPtr(a.convertProject(project)),
		Templates: payload,
	})
}

// ShowProjectTemplate implements the v1.ServerInterface.
func (a *API) ShowProjectTemplate(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "ShowProjectTemplate").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateResponse(
		a.convertTemplate(record),
	))
}

// CreateProjectTemplate implements the v1.ServerInterface.
func (a *API) CreateProjectTemplate(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectTemplateBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectTemplate").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Template{
		ProjectID: project.ID,
	}

	if body.Executor != nil {
		record.Executor = FromPtr(body.Executor)
	}

	if body.RepositoryID != nil {
		record.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.InventoryID != nil {
		record.InventoryID = FromPtr(body.InventoryID)
	}

	if body.EnvironmentID != nil {
		record.EnvironmentID = FromPtr(body.EnvironmentID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Description != nil {
		record.Description = FromPtr(body.Description)
	}

	if body.Playbook != nil {
		record.Playbook = FromPtr(body.Playbook)
	}

	if body.Arguments != nil {
		record.Arguments = FromPtr(body.Arguments)
	}

	if body.Limit != nil {
		record.Limit = FromPtr(body.Limit)
	}

	if body.Branch != nil {
		record.Branch = FromPtr(body.Branch)
	}

	if body.AllowOverride != nil {
		record.Override = FromPtr(body.AllowOverride)
	}

	if body.Surveys != nil {
		record.Surveys = make([]*model.TemplateSurvey, 0)

		for _, row := range FromPtr(body.Surveys) {
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

	if body.Vaults != nil {
		record.Vaults = make([]*model.TemplateVault, 0)

		for _, row := range FromPtr(body.Vaults) {
			vault := &model.TemplateVault{}

			if row.CredentialID != nil {
				vault.CredentialID = FromPtr(row.CredentialID)
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
			Str("project", project.ID).
			Str("action", "CreateProjectTemplate").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Create(
		ctx,
		project,
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
				Message: ToPtr("Failed to validate template"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectTemplate").
			Msg("Failed to create template")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create template"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateResponse(
		a.convertTemplate(record),
	))
}

// UpdateProjectTemplate implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplate(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	body := &UpdateProjectTemplateBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "UpdateProjectTemplate").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectTemplate").
			Str("project", project.ID).
			Str("template", record.ID).
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.RepositoryID != nil {
		record.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.InventoryID != nil {
		record.InventoryID = FromPtr(body.InventoryID)
	}

	if body.EnvironmentID != nil {
		record.EnvironmentID = FromPtr(body.EnvironmentID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Description != nil {
		record.Description = FromPtr(body.Description)
	}

	if body.Playbook != nil {
		record.Playbook = FromPtr(body.Playbook)
	}

	if body.Arguments != nil {
		record.Arguments = FromPtr(body.Arguments)
	}

	if body.Limit != nil {
		record.Limit = FromPtr(body.Limit)
	}

	if body.Branch != nil {
		record.Branch = FromPtr(body.Branch)
	}

	if body.AllowOverride != nil {
		record.Override = FromPtr(body.AllowOverride)
	}

	if body.Surveys != nil {
		record.Surveys = make([]*model.TemplateSurvey, 0)

		for _, row := range FromPtr(body.Surveys) {
			survey := &model.TemplateSurvey{}

			if row.ID != nil {
				survey.ID = FromPtr(row.ID)
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

					if val.ID != nil {
						value.ID = FromPtr(val.ID)
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

	if body.Vaults != nil {
		record.Vaults = make([]*model.TemplateVault, 0)

		for _, row := range FromPtr(body.Vaults) {
			vault := &model.TemplateVault{}

			if row.ID != nil {
				vault.ID = FromPtr(row.ID)
			}

			if row.CredentialID != nil {
				vault.CredentialID = FromPtr(row.CredentialID)
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
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "UpdateProjectTemplate").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Update(
		ctx,
		project,
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
				Message: ToPtr("Failed to validate template"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "UpdateProjectTemplate").
			Msg("Failed to update template")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update template"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateResponse(
		a.convertTemplate(record),
	))
}

// DeleteProjectTemplate implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplate(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "DeletProjectTemplate").
			Msg("Failed to delete template")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete template"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted template"),
		Status:  ToPtr(http.StatusOK),
	})
}

// CreateProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) CreateProjectTemplateSurvey(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	body := &CreateProjectTemplateSurveyBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateSurvey").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	child := &model.TemplateSurvey{
		TemplateID: record.ID,
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Title != nil {
		child.Title = FromPtr(body.Title)
	}

	if body.Description != nil {
		child.Description = FromPtr(body.Description)
	}

	if body.Required != nil {
		child.Required = FromPtr(body.Required)
	}

	if body.Values != nil {
		child.Values = make([]*model.TemplateValue, 0)

		for _, val := range FromPtr(body.Values) {
			value := &model.TemplateValue{}

			if val.Name != nil {
				value.Name = FromPtr(val.Name)
			}

			if val.Value != nil {
				value.Value = FromPtr(val.Value)
			}

			child.Values = append(child.Values, value)
		}
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateSurvey").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.CreateSurvey(
		ctx,
		record,
		child,
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
				Message: ToPtr("Failed to validate template survey"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateSurvey").
			Msg("Failed to create template survey")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create template survey"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateSurvey").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateSurveyResponse(
		a.convertTemplateSurvey(child),
	))
}

// UpdateProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplateSurvey(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID, _ SurveyID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	child := a.ProjectTemplateSurveyFromContext(ctx)
	body := &UpdateProjectTemplateSurveyBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "UpdateProjectTemplateSurvey").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "UpdateProjectTemplateSurvey").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Title != nil {
		child.Title = FromPtr(body.Title)
	}

	if body.Description != nil {
		child.Description = FromPtr(body.Description)
	}

	if body.Required != nil {
		child.Required = FromPtr(body.Required)
	}

	if body.Values != nil {
		child.Values = make([]*model.TemplateValue, 0)

		for _, val := range FromPtr(body.Values) {
			value := &model.TemplateValue{}

			if val.Name != nil {
				value.Name = FromPtr(val.Name)
			}

			if val.Value != nil {
				value.Value = FromPtr(val.Value)
			}

			child.Values = append(child.Values, value)
		}
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "UpdateProjectTemplateSurvey").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.UpdateSurvey(
		ctx,
		record,
		child,
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
				Message: ToPtr("Failed to validate template survey"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "UpdateProjectTemplateSurvey").
			Msg("Failed to update template survey")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update template survey"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "UpdateProjectTemplateSurvey").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateSurveyResponse(
		a.convertTemplateSurvey(child),
	))
}

// DeleteProjectTemplateSurvey implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplateSurvey(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID, _ SurveyID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	child := a.ProjectTemplateSurveyFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.DeleteSurvey(
		ctx,
		record,
		child.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("survey", child.ID).
			Str("action", "DeletProjectTemplateSurvey").
			Msg("Failed to delete template")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete template survey"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted template survey"),
		Status:  ToPtr(http.StatusOK),
	})
}

// CreateProjectTemplateVault implements the v1.ServerInterface.
func (a *API) CreateProjectTemplateVault(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	body := &CreateProjectTemplateVaultBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateVault").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	child := &model.TemplateVault{
		TemplateID: record.ID,
	}

	if body.CredentialID != nil {
		child.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Script != nil {
		child.Script = FromPtr(body.Script)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateVault").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.CreateVault(
		ctx,
		record,
		child,
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
				Message: ToPtr("Failed to validate template vault"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateVault").
			Msg("Failed to create template vault")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create template vault"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("action", "CreateProjectTemplateVault").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateVaultResponse(
		a.convertTemplateVault(child),
	))
}

// UpdateProjectTemplateVault implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplateVault(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID, _ VaultID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	child := a.ProjectTemplateVaultFromContext(ctx)
	body := &UpdateProjectTemplateVaultBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "UpdateProjectTemplateVault").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "UpdateProjectTemplateVault").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.CredentialID != nil {
		child.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Script != nil {
		child.Script = FromPtr(body.Script)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "UpdateProjectTemplateVault").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.UpdateVault(
		ctx,
		record,
		child,
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
				Message: ToPtr("Failed to validate template vault"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "UpdateProjectTemplateVault").
			Msg("Failed to update template vault")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update template vault"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "UpdateProjectTemplateVault").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectTemplateVaultResponse(
		a.convertTemplateVault(child),
	))
}

// DeleteProjectTemplateVault implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplateVault(w http.ResponseWriter, r *http.Request, _ ProjectID, _ TemplateID, _ VaultID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectTemplateFromContext(ctx)
	child := a.ProjectTemplateVaultFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.DeleteVault(
		ctx,
		record,
		child.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("template", record.ID).
			Str("vault", child.ID).
			Str("action", "DeletProjectTemplateVault").
			Msg("Failed to delete template")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete template vault"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted template vault"),
	})
}

func (a *API) convertTemplate(record *model.Template) Template {
	result := Template{
		ID:            ToPtr(record.ID),
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
		result.RepositoryID = ToPtr(record.RepositoryID)

		result.Repository = ToPtr(
			a.convertRepository(
				record.Repository,
			),
		)
	}

	if record.Inventory != nil {
		result.InventoryID = ToPtr(record.InventoryID)

		result.Inventory = ToPtr(
			a.convertInventory(
				record.Inventory,
			),
		)
	}

	if record.Environment != nil {
		result.EnvironmentID = ToPtr(record.EnvironmentID)

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
		ID:          ToPtr(record.ID),
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
		ID:    ToPtr(record.ID),
		Name:  ToPtr(record.Name),
		Value: ToPtr(record.Value),
	}

	return result
}

func (a *API) convertTemplateVault(record *model.TemplateVault) TemplateVault {
	result := TemplateVault{
		ID:     ToPtr(record.ID),
		Name:   ToPtr(record.Name),
		Kind:   ToPtr(TemplateVaultKind(record.Kind)),
		Script: ToPtr(record.Script),
	}

	if record.Credential != nil {
		result.CredentialID = ToPtr(record.CredentialID)

		result.Credential = ToPtr(
			a.convertCredential(
				record.Credential,
			),
		)
	}

	return result
}

// AllowShowProjectTemplate defines a middleware to check permissions.
func (a *API) AllowShowProjectTemplate(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectTemplate defines a middleware to check permissions.
func (a *API) AllowManageProjectTemplate(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listTemplatesSorting(request ListProjectTemplatesParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	//

	if request.Order != nil {
		sort = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}
