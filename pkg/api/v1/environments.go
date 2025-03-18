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

// ListProjectEnvironments implements the v1.ServerInterface.
func (a *API) ListProjectEnvironments(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectEnvironmentsParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listEnvironmentsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.List(
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
			Str("action", "ListProjectEnvironments").
			Msg("Failed to load environments")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load environments"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Environment, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			log.Error().
				Err(err).
				Str("project", project.ID).
				Str("action", "ListProjectEnvironments").
				Msg("Failed to decrypt secrets")

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertEnvironment(record)
	}

	render.JSON(w, r, ProjectEnvironmentsResponse{
		Total:        count,
		Limit:        limit,
		Offset:       offset,
		Project:      ToPtr(a.convertProject(project)),
		Environments: payload,
	})
}

// ShowProjectEnvironment implements the v1.ServerInterface.
func (a *API) ShowProjectEnvironment(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "ShowProjectEnvironment").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentResponse(
		a.convertEnvironment(record),
	))
}

// CreateProjectEnvironment implements the v1.ServerInterface.
func (a *API) CreateProjectEnvironment(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectEnvironmentBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectEnvironment").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Environment{
		ProjectID: project.ID,
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Secrets != nil {
		record.Secrets = make([]*model.EnvironmentSecret, 0)

		for _, row := range FromPtr(body.Secrets) {
			secret := &model.EnvironmentSecret{}

			if row.Kind != nil {
				secret.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				secret.Name = FromPtr(row.Name)
			}

			if row.Content != nil {
				secret.Content = FromPtr(row.Content)
			}

			record.Secrets = append(record.Secrets, secret)
		}
	}

	if body.Values != nil {
		record.Values = make([]*model.EnvironmentValue, 0)

		for _, row := range FromPtr(body.Values) {
			value := &model.EnvironmentValue{}

			if row.Kind != nil {
				value.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				value.Name = FromPtr(row.Name)
			}

			if row.Content != nil {
				value.Content = FromPtr(row.Content)
			}

			record.Values = append(record.Values, value)
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectEnvironment").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Create(
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
				Message: ToPtr("Failed to validate environment"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectEnvironment").
			Msg("Failed to create environment")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentResponse(
		a.convertEnvironment(record),
	))
}

// UpdateProjectEnvironment implements the v1.ServerInterface.
func (a *API) UpdateProjectEnvironment(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	body := &UpdateProjectEnvironmentBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "UpdateProjectEnvironment").
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
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "UpdateProjectEnvironment").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Secrets != nil {
		record.Secrets = make([]*model.EnvironmentSecret, 0)

		for _, row := range FromPtr(body.Secrets) {
			secret := &model.EnvironmentSecret{}

			if row.ID != nil {
				secret.ID = FromPtr(row.ID)
			}

			if row.Kind != nil {
				secret.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				secret.Name = FromPtr(row.Name)
			}

			if row.Content != nil {
				secret.Content = FromPtr(row.Content)
			}

			record.Secrets = append(record.Secrets, secret)
		}
	}

	if body.Values != nil {
		record.Values = make([]*model.EnvironmentValue, 0)

		for _, row := range FromPtr(body.Values) {
			value := &model.EnvironmentValue{}

			if row.ID != nil {
				value.ID = FromPtr(row.ID)
			}

			if row.Kind != nil {
				value.Kind = string(FromPtr(row.Kind))
			}

			if row.Name != nil {
				value.Name = FromPtr(row.Name)
			}

			if row.Content != nil {
				value.Content = FromPtr(row.Content)
			}

			record.Values = append(record.Values, value)
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "UpdateProjectEnvironment").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Update(
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
				Message: ToPtr("Failed to validate environment"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "UpdateProjectEnvironment").
			Msg("Failed to update environment")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update environment"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentResponse(
		a.convertEnvironment(record),
	))
}

// DeleteProjectEnvironment implements the v1.ServerInterface.
func (a *API) DeleteProjectEnvironment(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "DeletProjectEnvironment").
			Msg("Failed to delete environment")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete environment"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted environment"),
		Status:  ToPtr(http.StatusOK),
	})
}

// CreateProjectEnvironmentSecret implements the v1.ServerInterface.
func (a *API) CreateProjectEnvironmentSecret(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	body := &CreateProjectEnvironmentSecretBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentSecret").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	child := &model.EnvironmentSecret{
		EnvironmentID: record.ID,
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Content != nil {
		child.Content = FromPtr(body.Content)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentSecret").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.CreateSecret(
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
				Message: ToPtr("Failed to validate environment secret"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentSecret").
			Msg("Failed to create environment secret")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create environment secret"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "CreateProjectEnvironmentSecret").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentSecretResponse(
		a.convertEnvironmentSecret(child),
	))
}

// UpdateProjectEnvironmentSecret implements the v1.ServerInterface.
func (a *API) UpdateProjectEnvironmentSecret(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID, _ SecretID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	child := a.ProjectEnvironmentSecretFromContext(ctx)
	body := &UpdateProjectEnvironmentSecretBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "UpdateProjectEnvironmentSecret").
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
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "UpdateProjectEnvironmentSecret").
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

	if body.Content != nil {
		child.Content = FromPtr(body.Content)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "UpdateProjectEnvironmentSecret").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.UpdateSecret(
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
				Message: ToPtr("Failed to validate environment secret"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "UpdateProjectEnvironmentSecret").
			Msg("Failed to update environment secret")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update environment secret"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "UpdateProjectEnvironmentSecret").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentSecretResponse(
		a.convertEnvironmentSecret(child),
	))
}

// DeleteProjectEnvironmentSecret implements the v1.ServerInterface.
func (a *API) DeleteProjectEnvironmentSecret(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID, _ SecretID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	child := a.ProjectEnvironmentSecretFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.DeleteSecret(
		ctx,
		record,
		child.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("secret", child.ID).
			Str("action", "DeleteProjectEnvironmentSecret").
			Msg("Failed to delete environment secret")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete environment secret"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted environment secret"),
		Status:  ToPtr(http.StatusOK),
	})
}

// CreateProjectEnvironmentValue implements the v1.ServerInterface.
func (a *API) CreateProjectEnvironmentValue(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	body := &CreateProjectEnvironmentValueBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentValue").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	child := &model.EnvironmentValue{
		EnvironmentID: record.ID,
	}

	if body.Kind != nil {
		child.Kind = string(FromPtr(body.Kind))
	}

	if body.Name != nil {
		child.Name = FromPtr(body.Name)
	}

	if body.Content != nil {
		child.Content = FromPtr(body.Content)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentValue").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.CreateValue(
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
				Message: ToPtr("Failed to validate environment value"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("action", "CreateProjectEnvironmentValue").
			Msg("Failed to create environment value")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create environment value"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "CreateProjectEnvironmentValue").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentValueResponse(
		a.convertEnvironmentValue(child),
	))
}

// UpdateProjectEnvironmentValue implements the v1.ServerInterface.
func (a *API) UpdateProjectEnvironmentValue(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID, _ ValueID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	child := a.ProjectEnvironmentValueFromContext(ctx)
	body := &UpdateProjectEnvironmentValueBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "UpdateProjectEnvironmentValue").
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
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "UpdateProjectEnvironmentValue").
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

	if body.Content != nil {
		child.Content = FromPtr(body.Content)
	}

	if err := child.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "UpdateProjectEnvironmentValue").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.UpdateValue(
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
				Message: ToPtr("Failed to validate environment value"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "UpdateProjectEnvironmentValue").
			Msg("Failed to update environment value")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update environment value"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := child.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "UpdateProjectEnvironmentValue").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectEnvironmentValueResponse(
		a.convertEnvironmentValue(child),
	))
}

// DeleteProjectEnvironmentValue implements the v1.ServerInterface.
func (a *API) DeleteProjectEnvironmentValue(w http.ResponseWriter, r *http.Request, _ ProjectID, _ EnvironmentID, _ ValueID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectEnvironmentFromContext(ctx)
	child := a.ProjectEnvironmentValueFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Environments.DeleteValue(
		ctx,
		record,
		child.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("environment", record.ID).
			Str("value", child.ID).
			Str("action", "DeletProjectEnvironmentValue").
			Msg("Failed to delete environment")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete environment value"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted environment value"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertEnvironment(record *model.Environment) Environment {
	result := Environment{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if len(record.Secrets) > 0 {
		secrets := make([]EnvironmentSecret, 0)

		for _, secret := range record.Secrets {
			secrets = append(
				secrets,
				a.convertEnvironmentSecret(secret),
			)
		}

		result.Secrets = ToPtr(secrets)
	}

	if len(record.Values) > 0 {
		values := make([]EnvironmentValue, 0)

		for _, value := range record.Values {
			values = append(
				values,
				a.convertEnvironmentValue(value),
			)
		}

		result.Values = ToPtr(values)
	}

	return result
}

func (a *API) convertEnvironmentSecret(record *model.EnvironmentSecret) EnvironmentSecret {
	result := EnvironmentSecret{
		ID:      ToPtr(record.ID),
		Kind:    ToPtr(EnvironmentSecretKind(record.Kind)),
		Name:    ToPtr(record.Name),
		Content: ToPtr(record.Content),
	}

	return result
}

func (a *API) convertEnvironmentValue(record *model.EnvironmentValue) EnvironmentValue {
	result := EnvironmentValue{
		ID:      ToPtr(record.ID),
		Kind:    ToPtr(EnvironmentValueKind(record.Kind)),
		Name:    ToPtr(record.Name),
		Content: ToPtr(record.Content),
	}

	return result
}

// AllowShowProjectEnvironment defines a middleware to check permissions.
func (a *API) AllowShowProjectEnvironment(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectEnvironment defines a middleware to check permissions.
func (a *API) AllowManageProjectEnvironment(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listEnvironmentsSorting(request ListProjectEnvironmentsParams) (string, string, int64, int64, string) {
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
