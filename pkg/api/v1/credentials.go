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
		slog.Error(
			"Failed to load credentials",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "ListProjectCredentials"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Credential, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			slog.Error(
				"Failed to decrypt secrets",
				slog.Any("error", err),
				slog.String("project", project.ID),
				slog.String("action", "ListProjectCredentials"),
			)

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
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", project.ID),
			slog.String("action", "ShowProjectCredential"),
		)

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
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	incoming := &model.Credential{
		ProjectID: project.ID,
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Kind != nil {
		incoming.Kind = FromPtr(body.Kind)
	}

	if body.Override != nil {
		incoming.Override = FromPtr(body.Override)
	}

	switch incoming.Kind {
	case "shell":
		if body.Shell != nil {
			if body.Shell.Username != nil {
				incoming.Shell.Username = FromPtr(body.Shell.Username)
			}

			if body.Shell.Password != nil {
				incoming.Shell.Password = FromPtr(body.Shell.Password)
			}

			if body.Shell.PrivateKey != nil {
				incoming.Shell.PrivateKey = FromPtr(body.Shell.PrivateKey)
			}
		}
	case "login":
		if body.Login != nil {
			if body.Login.Username != nil {
				incoming.Login.Username = FromPtr(body.Login.Username)
			}

			if body.Login.Password != nil {
				incoming.Login.Password = FromPtr(body.Login.Password)
			}
		}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Create(
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
				Message: ToPtr("Failed to validate credential"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to create credential",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", project.ID),
			slog.String("action", "CreateProjectCredential"),
		)

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

// UpdateProjectCredential implements the v1.ServerInterface.
func (a *API) UpdateProjectCredential(w http.ResponseWriter, r *http.Request, _ ProjectID, _ CredentialID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	incoming := a.ProjectCredentialFromContext(ctx)
	body := &UpdateProjectCredentialBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", incoming.ID),
			slog.String("action", "UpdateProjectCredential"),
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
			slog.String("credential", project.ID),
			slog.String("action", "UpdateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.Slug != nil {
		incoming.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		incoming.Name = FromPtr(body.Name)
	}

	if body.Kind != nil {
		incoming.Kind = FromPtr(body.Kind)
	}

	if body.Override != nil {
		incoming.Override = FromPtr(body.Override)
	}

	switch incoming.Kind {
	case "shell":
		incoming.Login = model.CredentialLogin{}

		if body.Shell != nil {
			if body.Shell.Username != nil {
				incoming.Shell.Username = FromPtr(body.Shell.Username)
			}

			if body.Shell.Password != nil {
				incoming.Shell.Password = FromPtr(body.Shell.Password)
			}

			if body.Shell.PrivateKey != nil {
				incoming.Shell.PrivateKey = FromPtr(body.Shell.PrivateKey)
			}
		}
	case "login":
		incoming.Shell = model.CredentialShell{}

		if body.Login != nil {
			if body.Login.Username != nil {
				incoming.Login.Username = FromPtr(body.Login.Username)
			}

			if body.Login.Password != nil {
				incoming.Login.Password = FromPtr(body.Login.Password)
			}
		}
	default:
		incoming.Shell = model.CredentialShell{}
		incoming.Login = model.CredentialLogin{}
	}

	if err := incoming.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", incoming.ID),
			slog.String("action", "UpdateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Update(
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
				Message: ToPtr("Failed to validate credential"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update credential",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", record.ID),
			slog.String("action", "UpdateProjectCredential"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", project.ID),
			slog.String("action", "UpdateProjectCredential"),
		)

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
		slog.Error(
			"Failed to delete credential",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("credential", record.ID),
			slog.String("action", "DeletProjectCredential"),
		)

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
