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

// ListProjectCredentials implements the v1.ServerInterface.
func (a *API) ListProjectCredentials(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectCredentialsParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listCredentialsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.List(
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
			Str("action", "ListProjectCredentials").
			Msg("Failed to load credentials")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Credential, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			log.Error().
				Err(err).
				Str("project", project.ID).
				Str("action", "ListProjectCredentials").
				Msg("Failed to decrypt secrets")

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertCredential(record)
	}

	render.JSON(w, r, ProjectCredentialsResponse{
		Total:       count,
		Limit:       limit,
		Offset:      offset,
		Project:     ToPtr(a.convertProject(project)),
		Credentials: payload,
	})
}

// ShowProjectCredential implements the v1.ServerInterface.
func (a *API) ShowProjectCredential(w http.ResponseWriter, r *http.Request, _ ProjectID, _ CredentialID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectCredentialFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("credential", record.ID).
			Str("action", "ShowProjectCredential").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectCredentialResponse(
		a.convertCredential(record),
	))
}

// CreateProjectCredential implements the v1.ServerInterface.
func (a *API) CreateProjectCredential(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectCredentialBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectCredential").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Credential{
		ProjectID: project.ID,
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Kind != nil {
		record.Kind = FromPtr(body.Kind)
	}

	if body.Override != nil {
		record.Override = FromPtr(body.Override)
	}

	switch record.Kind {
	case "shell":
		if body.Shell != nil {
			if body.Shell.Username != nil {
				record.Shell.Username = FromPtr(body.Shell.Username)
			}

			if body.Shell.Password != nil {
				record.Shell.Password = FromPtr(body.Shell.Password)
			}

			if body.Shell.PrivateKey != nil {
				record.Shell.PrivateKey = FromPtr(body.Shell.PrivateKey)
			}
		}
	case "login":
		if body.Login != nil {
			if body.Login.Username != nil {
				record.Login.Username = FromPtr(body.Login.Username)
			}

			if body.Login.Password != nil {
				record.Login.Password = FromPtr(body.Login.Password)
			}
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectCredential").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Create(
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
				Message: ToPtr("Failed to validate credential"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectCredential").
			Msg("Failed to create credential")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectCredentialResponse(
		a.convertCredential(record),
	))
}

// UpdateProjectCredential implements the v1.ServerInterface.
func (a *API) UpdateProjectCredential(w http.ResponseWriter, r *http.Request, _ ProjectID, _ CredentialID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectCredentialFromContext(ctx)
	body := &UpdateProjectCredentialBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("credential", record.ID).
			Str("action", "UpdateProjectCredential").
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
			Str("credential", record.ID).
			Str("action", "UpdateProjectCredential").
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

	if body.Kind != nil {
		record.Kind = FromPtr(body.Kind)
	}

	if body.Override != nil {
		record.Override = FromPtr(body.Override)
	}

	switch record.Kind {
	case "shell":
		record.Login = model.CredentialLogin{}

		if body.Shell != nil {
			if body.Shell.Username != nil {
				record.Shell.Username = FromPtr(body.Shell.Username)
			}

			if body.Shell.Password != nil {
				record.Shell.Password = FromPtr(body.Shell.Password)
			}

			if body.Shell.PrivateKey != nil {
				record.Shell.PrivateKey = FromPtr(body.Shell.PrivateKey)
			}
		}
	case "login":
		record.Shell = model.CredentialShell{}

		if body.Login != nil {
			if body.Login.Username != nil {
				record.Login.Username = FromPtr(body.Login.Username)
			}

			if body.Login.Password != nil {
				record.Login.Password = FromPtr(body.Login.Password)
			}
		}
	default:
		record.Shell = model.CredentialShell{}
		record.Login = model.CredentialLogin{}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("credential", record.ID).
			Str("action", "UpdateProjectCredential").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Update(
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
				Message: ToPtr("Failed to validate credential"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("credential", record.ID).
			Str("action", "UpdateProjectCredential").
			Msg("Failed to update credential")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectCredentialResponse(
		a.convertCredential(record),
	))
}

// DeleteProjectCredential implements the v1.ServerInterface.
func (a *API) DeleteProjectCredential(w http.ResponseWriter, r *http.Request, _ ProjectID, _ CredentialID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectCredentialFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("credential", record.ID).
			Str("action", "DeletProjectCredential").
			Msg("Failed to delete credential")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete credential"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted credential"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertCredential(record *model.Credential) Credential {
	result := Credential{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Kind:      ToPtr(CredentialKind(record.Kind)),
		Override:  ToPtr(record.Override),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	switch record.Kind {
	case "shell":
		result.Shell = ToPtr(
			a.converCredentialShell(
				record.Shell,
			),
		)
	case "login":
		result.Login = ToPtr(
			a.convertCredentialLogin(
				record.Login,
			),
		)
	}

	return result
}

func (a *API) converCredentialShell(record model.CredentialShell) CredentialShell {
	return CredentialShell{
		Username:   ToPtr(record.Username),
		Password:   ToPtr(record.Password),
		PrivateKey: ToPtr(record.PrivateKey),
	}
}

func (a *API) convertCredentialLogin(record model.CredentialLogin) CredentialLogin {
	return CredentialLogin{
		Username: ToPtr(record.Username),
		Password: ToPtr(record.Password),
	}
}

// AllowShowProjectCredential defines a middleware to check permissions.
func (a *API) AllowShowProjectCredential(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectCredential defines a middleware to check permissions.
func (a *API) AllowManageProjectCredential(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listCredentialsSorting(request ListProjectCredentialsParams) (string, string, int64, int64, string) {
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
