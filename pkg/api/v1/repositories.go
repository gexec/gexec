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

// ListProjectRepositories implements the v1.ServerInterface.
func (a *API) ListProjectRepositories(ctx context.Context, request ListProjectRepositoriesRequestObject) (ListProjectRepositoriesResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectRepositories404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectRepositories").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectRepositories500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectRepositories404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listRepositoriesSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.List(
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
			Str("action", "ListProjectRepositories").
			Str("project", parent.ID).
			Msg("Failed to load repositories")

		return ListProjectRepositories500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load repositories"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Repository, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
			log.Error().
				Err(err).
				Str("action", "ListProjectCredentials").
				Str("project", parent.ID).
				Msg("Failed to decrypt secrets")

			return ListProjectRepositories500JSONResponse{InternalServerErrorJSONResponse{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			}}, nil
		}

		payload[id] = a.convertRepository(record)
	}

	return ListProjectRepositories200JSONResponse{ProjectRepositoriesResponseJSONResponse{
		Total:        count,
		Limit:        limit,
		Offset:       offset,
		Project:      ToPtr(a.convertProject(parent)),
		Repositories: payload,
	}}, nil
}

// ShowProjectRepository implements the v1.ServerInterface.
func (a *API) ShowProjectRepository(ctx context.Context, request ShowProjectRepositoryRequestObject) (ShowProjectRepositoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectRepository").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Show(
		ctx,
		parent.ID,
		request.RepositoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrRepositoryNotFound) {
			return ShowProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectRepository").
			Str("project", parent.ID).
			Str("repository", request.RepositoryId).
			Msg("Failed to load repository")

		return ShowProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowRepository(ctx, repositoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or repository"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil { // TODO: remove it, security risk
		log.Error().
			Err(err).
			Str("action", "ShowProjectRepository").
			Str("project", parent.ID).
			Str("credential", record.ID).
			Msg("Failed to decrypt secrets")

		return ShowProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return ShowProjectRepository200JSONResponse{ProjectRepositoryResponseJSONResponse(
		a.convertRepository(record),
	)}, nil
}

// CreateProjectRepository implements the v1.ServerInterface.
func (a *API) CreateProjectRepository(ctx context.Context, request CreateProjectRepositoryRequestObject) (CreateProjectRepositoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectRepository").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateRepository(ctx, repositoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Repository{
		ProjectID: parent.ID,
	}

	if request.Body.CredentialId != nil {
		record.CredentialID = FromPtr(request.Body.CredentialId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Url != nil {
		record.URL = FromPtr(request.Body.Url)
	}

	if request.Body.Branch != nil {
		record.Branch = FromPtr(request.Body.Branch)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateProjectRepository").
			Str("project", parent.ID).
			Msg("Failed to encrypt secrets")

		return CreateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Create(
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

			return CreateProjectRepository422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate repository"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectRepository").
			Str("project", parent.ID).
			Msg("Failed to create repository")

		return CreateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectRepository200JSONResponse{ProjectRepositoryResponseJSONResponse(
		a.convertRepository(record),
	)}, nil
}

// UpdateProjectRepository implements the v1.ServerInterface.
func (a *API) UpdateProjectRepository(ctx context.Context, request UpdateProjectRepositoryRequestObject) (UpdateProjectRepositoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectRepository").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Show(
		ctx,
		parent.ID,
		request.RepositoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrRepositoryNotFound) {
			return UpdateProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectRepository").
			Str("project", parent.ID).
			Str("repository", request.RepositoryId).
			Msg("Failed to load repository")

		return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageRepository(ctx, repositoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or repository"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectRepository").
			Str("project", parent.ID).
			Str("repository", record.ID).
			Msg("Failed to decrypt secrets")

		return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if request.Body.CredentialId != nil {
		record.CredentialID = FromPtr(request.Body.CredentialId)
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Url != nil {
		record.URL = FromPtr(request.Body.Url)
	}

	if request.Body.Branch != nil {
		record.Branch = FromPtr(request.Body.Branch)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectRepository").
			Str("project", parent.ID).
			Str("repository", record.ID).
			Msg("Failed to encrypt secrets")

		return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Update(
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

			return UpdateProjectRepository422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate repository"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectRepository").
			Str("project", parent.ID).
			Str("repository", record.ID).
			Msg("Failed to update repository")

		return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectRepository200JSONResponse{ProjectRepositoryResponseJSONResponse(
		a.convertRepository(record),
	)}, nil
}

// DeleteProjectRepository implements the v1.ServerInterface.
func (a *API) DeleteProjectRepository(ctx context.Context, request DeleteProjectRepositoryRequestObject) (DeleteProjectRepositoryResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectRepository").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Show(
		ctx,
		parent.ID,
		request.RepositoryId,
	)

	if err != nil {
		if errors.Is(err, store.ErrRepositoryNotFound) {
			return DeleteProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or repository"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectRepository").
			Str("project", parent.ID).
			Str("repository", request.RepositoryId).
			Msg("Failed to load repository")

		return DeleteProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load repository"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageRepository(ctx, repositoryPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectRepository404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or repository"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Repositories.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectRepository").
			Str("project", parent.ID).
			Str("repository", record.ID).
			Msg("Failed to delete repository")

		return DeleteProjectRepository400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete repository"),
		}}, nil
	}

	return DeleteProjectRepository200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted repository"),
	}}, nil
}

func (a *API) convertRepository(record *model.Repository) Repository {
	result := Repository{
		Id:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Url:       ToPtr(record.URL),
		Branch:    ToPtr(record.Branch),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
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

type repositoryPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Repository
	OwnerOnly bool
}

func (a *API) permitCreateRepository(ctx context.Context, definition repositoryPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowRepository(ctx context.Context, definition repositoryPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageRepository(ctx context.Context, definition repositoryPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listRepositoriesSorting(request ListProjectRepositoriesRequestObject) (string, string, int64, int64, string) {
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
