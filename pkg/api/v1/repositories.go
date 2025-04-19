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

// ListProjectRepositories implements the v1.ServerInterface.
func (a *API) ListProjectRepositories(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectRepositoriesParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listRepositoriesSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.List(
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
			"Failed to load repositories",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "ListProjectRepositories"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load repositories"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Repository, len(records))
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

		payload[id] = a.convertRepository(record)
	}

	render.JSON(w, r, ProjectRepositoriesResponse{
		Total:        count,
		Limit:        limit,
		Offset:       offset,
		Project:      ToPtr(a.convertProject(project)),
		Repositories: payload,
	})
}

// ShowProjectRepository implements the v1.ServerInterface.
func (a *API) ShowProjectRepository(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RepositoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectRepositoryFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", project.ID),
			slog.String("action", "ShowProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRepositoryResponse(
		a.convertRepository(record),
	))
}

// CreateProjectRepository implements the v1.ServerInterface.
func (a *API) CreateProjectRepository(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectRepositoryBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Repository{
		ProjectID: project.ID,
	}

	if body.CredentialID != nil {
		record.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Url != nil {
		record.URL = FromPtr(body.Url)
	}

	if body.Branch != nil {
		record.Branch = FromPtr(body.Branch)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Create(
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
				Message: ToPtr("Failed to validate repository"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to create repository",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRepositoryResponse(
		a.convertRepository(record),
	))
}

// UpdateProjectRepository implements the v1.ServerInterface.
func (a *API) UpdateProjectRepository(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RepositoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectRepositoryFromContext(ctx)
	body := &UpdateProjectRepositoryBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", record.ID),
			slog.String("action", "UpdateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", project.ID),
			slog.String("action", "UpdateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.CredentialID != nil {
		record.CredentialID = FromPtr(body.CredentialID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Url != nil {
		record.URL = FromPtr(body.Url)
	}

	if body.Branch != nil {
		record.Branch = FromPtr(body.Branch)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", record.ID),
			slog.String("action", "UpdateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Update(
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
				Message: ToPtr("Failed to validate repository"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update repository",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", record.ID),
			slog.String("action", "UpdateProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRepositoryResponse(
		a.convertRepository(record),
	))
}

// DeleteProjectRepository implements the v1.ServerInterface.
func (a *API) DeleteProjectRepository(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RepositoryID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete repository",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("repository", record.ID),
			slog.String("action", "DeleteProjectRepository"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete repository"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted repository"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertRepository(record *model.Repository) Repository {
	result := Repository{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Url:       ToPtr(record.URL),
		Branch:    ToPtr(record.Branch),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
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

// AllowShowProjectRepository defines a middleware to check permissions.
func (a *API) AllowShowProjectRepository(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectRepository defines a middleware to check permissions.
func (a *API) AllowManageProjectRepository(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listRepositoriesSorting(request ListProjectRepositoriesParams) (string, string, int64, int64, string) {
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
