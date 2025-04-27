package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
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
		slog.Error(
			"Failed to load templates",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "ListProjectTemplates"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load templates"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Template, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			slog.Error(
				"Failed to decrypt secrets",
				slog.Any("error", err),
				slog.String("project", project.ID),
				slog.String("action", "ListProjectTemplates"),
			)

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
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", project.ID),
			slog.String("action", "ShowProjectTemplate"),
		)

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
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectTemplate"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	incoming := &model.Template{
		ProjectID: project.ID,
	}

	if body.Executor != nil {
		incoming.Executor = FromPtr(body.Executor)
	}

	if body.RepositoryID != nil {
		incoming.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.InventoryID != nil {
		incoming.InventoryID = FromPtr(body.InventoryID)
	}

	if body.EnvironmentID != nil {
		incoming.EnvironmentID = FromPtr(body.EnvironmentID)
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Description != nil {
		incoming.Description = FromPtr(body.Description)
	}

	if body.Path != nil {
		incoming.Path = FromPtr(body.Path)
	}

	if body.Arguments != nil {
		incoming.Arguments = FromPtr(body.Arguments)
	}

	if body.Limit != nil {
		incoming.Limit = FromPtr(body.Limit)
	}

	if body.Branch != nil {
		incoming.Branch = FromPtr(body.Branch)
	}

	if body.AllowOverride != nil {
		incoming.Override = FromPtr(body.AllowOverride)
	}

	if body.Surveys != nil {
		incoming.Surveys = make([]*model.TemplateSurvey, 0)

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

			incoming.Surveys = append(incoming.Surveys, survey)
		}
	}

	if body.Vaults != nil {
		incoming.Vaults = make([]*model.TemplateVault, 0)

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

			incoming.Vaults = append(incoming.Vaults, vault)
		}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectTemplate"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Create(
		ctx,
		project,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to create template",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectTemplate"),
		)

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
	incoming := a.ProjectTemplateFromContext(ctx)
	body := &UpdateProjectTemplateBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", incoming.ID),
			slog.String("action", "UpdateProjectTemplate"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := incoming.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", incoming.ID),
			slog.String("action", "UpdateProjectTemplate"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.RepositoryID != nil {
		incoming.RepositoryID = FromPtr(body.RepositoryID)
	}

	if body.InventoryID != nil {
		incoming.InventoryID = FromPtr(body.InventoryID)
	}

	if body.EnvironmentID != nil {
		incoming.EnvironmentID = FromPtr(body.EnvironmentID)
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Description != nil {
		incoming.Description = FromPtr(body.Description)
	}

	if body.Path != nil {
		incoming.Path = FromPtr(body.Path)
	}

	if body.Arguments != nil {
		incoming.Arguments = FromPtr(body.Arguments)
	}

	if body.Limit != nil {
		incoming.Limit = FromPtr(body.Limit)
	}

	if body.Branch != nil {
		incoming.Branch = FromPtr(body.Branch)
	}

	if body.AllowOverride != nil {
		incoming.Override = FromPtr(body.AllowOverride)
	}

	if body.Surveys != nil {
		incoming.Surveys = make([]*model.TemplateSurvey, 0)

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

			incoming.Surveys = append(incoming.Surveys, survey)
		}
	}

	if body.Vaults != nil {
		incoming.Vaults = make([]*model.TemplateVault, 0)

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

			incoming.Vaults = append(incoming.Vaults, vault)
		}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", incoming.ID),
			slog.String("action", "UpdateProjectTemplate"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Update(
		ctx,
		project,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to update template",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "UpdateProjectTemplate"),
		)

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
	record := a.ProjectTemplateFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete template",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "DeletProjectTemplate"),
		)

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
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	incoming := &model.TemplateSurvey{
		TemplateID: record.ID,
	}

	if body.Kind != nil {
		incoming.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Title != nil {
		incoming.Title = FromPtr(body.Title)
	}

	if body.Description != nil {
		incoming.Description = FromPtr(body.Description)
	}

	if body.Required != nil {
		incoming.Required = FromPtr(body.Required)
	}

	if body.Values != nil {
		incoming.Values = make([]*model.TemplateValue, 0)

		for _, val := range FromPtr(body.Values) {
			value := &model.TemplateValue{}

			if val.Name != nil {
				value.Name = FromPtr(val.Name)
			}

			if val.Value != nil {
				value.Value = FromPtr(val.Value)
			}

			incoming.Values = append(incoming.Values, value)
		}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	child, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.CreateSurvey(
		ctx,
		record,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to create template survey",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create template survey"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", project.ID),
			slog.String("action", "CreateProjectTemplateSurvey"),
		)

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
	incoming := a.ProjectTemplateSurveyFromContext(ctx)
	body := &UpdateProjectTemplateSurveyBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", incoming.ID),
			slog.String("action", "UpdateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := incoming.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", incoming.ID),
			slog.String("action", "UpdateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.Kind != nil {
		incoming.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Title != nil {
		incoming.Title = FromPtr(body.Title)
	}

	if body.Description != nil {
		incoming.Description = FromPtr(body.Description)
	}

	if body.Required != nil {
		incoming.Required = FromPtr(body.Required)
	}

	if body.Values != nil {
		incoming.Values = make([]*model.TemplateValue, 0)

		for _, val := range FromPtr(body.Values) {
			value := &model.TemplateValue{}

			if val.Name != nil {
				value.Name = FromPtr(val.Name)
			}

			if val.Value != nil {
				value.Value = FromPtr(val.Value)
			}

			incoming.Values = append(incoming.Values, value)
		}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", incoming.ID),
			slog.String("action", "UpdateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	child, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.UpdateSurvey(
		ctx,
		record,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to update template survey",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", child.ID),
			slog.String("action", "UpdateProjectTemplateSurvey"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update template survey"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", child.ID),
			slog.String("action", "UpdateProjectTemplateSurvey"),
		)

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
		slog.Error(
			"Failed to delete template survey",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("survey", child.ID),
			slog.String("action", "DeletProjectTemplateSurvey"),
		)

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
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	incoming := &model.TemplateVault{
		TemplateID: record.ID,
	}

	if body.CredentialID != nil {
		incoming.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Kind != nil {
		incoming.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Script != nil {
		incoming.Script = FromPtr(body.Script)
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	child, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.CreateVault(
		ctx,
		record,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to create template vault",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("action", "CreateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create template vault"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", project.ID),
			slog.String("action", "CreateProjectTemplateVault"),
		)

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
	incoming := a.ProjectTemplateVaultFromContext(ctx)
	body := &UpdateProjectTemplateVaultBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", incoming.ID),
			slog.String("action", "UpdateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := incoming.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", incoming.ID),
			slog.String("action", "UpdateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.CredentialID != nil {
		incoming.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Kind != nil {
		incoming.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Script != nil {
		incoming.Script = FromPtr(body.Script)
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", incoming.ID),
			slog.String("action", "UpdateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	child, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Templates.UpdateVault(
		ctx,
		record,
		incoming,
	)

	if err != nil {
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

		slog.Error(
			"Failed to update template vault",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", child.ID),
			slog.String("action", "UpdateProjectTemplateVault"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update template vault"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", child.ID),
			slog.String("action", "UpdateProjectTemplateVault"),
		)

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
		slog.Error(
			"Failed to delete template vault",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("template", record.ID),
			slog.String("vault", child.ID),
			slog.String("action", "DeletProjectTemplateVault"),
		)

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
		Path:          ToPtr(record.Path),
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
		order = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}
