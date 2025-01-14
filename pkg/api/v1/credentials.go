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

// ListProjectCredentials implements the v1.ServerInterface.
func (a *API) ListProjectCredentials(ctx context.Context, request ListProjectCredentialsRequestObject) (ListProjectCredentialsResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ListProjectCredentials404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ListProjectCredentials").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ListProjectCredentials500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowProject(ctx, projectPermissions{
		User:   current.GetUser(ctx),
		Record: parent,
	}) {
		return ListProjectCredentials404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	sort, order, limit, offset, search := listCredentialsSorting(request)
	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.List(
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
			Str("action", "ListProjectCredentials").
			Str("project", parent.ID).
			Msg("Failed to load credentials")

		return ListProjectCredentials500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	payload := make([]Credential, len(records))
	for id, record := range records {
		payload[id] = a.convertCredential(record)
	}

	return ListProjectCredentials200JSONResponse{ProjectCredentialsResponseJSONResponse{
		Total:       count,
		Limit:       limit,
		Offset:      offset,
		Project:     ToPtr(a.convertProject(parent)),
		Credentials: payload,
	}}, nil
}

// ShowProjectCredential implements the v1.ServerInterface.
func (a *API) ShowProjectCredential(ctx context.Context, request ShowProjectCredentialRequestObject) (ShowProjectCredentialResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return ShowProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectCredential").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return ShowProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Show(
		ctx,
		parent.ID,
		request.CredentialId,
	)

	if err != nil {
		if errors.Is(err, store.ErrCredentialNotFound) {
			return ShowProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "ShowProjectCredential").
			Str("project", parent.ID).
			Str("credential", request.CredentialId).
			Msg("Failed to load credential")

		return ShowProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitShowCredential(ctx, credentialPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return ShowProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or credential"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	return ShowProjectCredential200JSONResponse{ProjectCredentialResponseJSONResponse(
		a.convertCredential(record),
	)}, nil
}

// CreateProjectCredential implements the v1.ServerInterface.
func (a *API) CreateProjectCredential(ctx context.Context, request CreateProjectCredentialRequestObject) (CreateProjectCredentialResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return CreateProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectCredential").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return CreateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitCreateCredential(ctx, credentialPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
	}) {
		return CreateProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	record := &model.Credential{
		ProjectID: parent.ID,
		Name:      request.Body.Name,
		Kind:      request.Body.Kind,
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Override != nil {
		record.Override = FromPtr(request.Body.Override)
	}

	switch request.Body.Kind {
	case "shell":
		if request.Body.Shell != nil {
			if request.Body.Shell.Username != nil {
				record.Shell.Username = FromPtr(request.Body.Shell.Username)
			}

			if request.Body.Shell.Password != nil {
				record.Shell.Password = FromPtr(request.Body.Shell.Password)
			}

			if request.Body.Shell.PrivateKey != nil {
				record.Shell.PrivateKey = FromPtr(request.Body.Shell.PrivateKey)
			}
		}
	case "login":
		if request.Body.Login != nil {
			if request.Body.Login.Username != nil {
				record.Login.Username = FromPtr(request.Body.Login.Username)
			}

			if request.Body.Login.Password != nil {
				record.Login.Password = FromPtr(request.Body.Login.Password)
			}
		}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "CreateProjectCredential").
			Str("project", parent.ID).
			Msg("Failed to encrypt secrets")

		return CreateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Create(
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

			return CreateProjectCredential422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate credential"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "CreateProjectCredential").
			Str("project", parent.ID).
			Msg("Failed to create credential")

		return CreateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to create credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return CreateProjectCredential200JSONResponse{ProjectCredentialResponseJSONResponse(
		a.convertCredential(record),
	)}, nil
}

// UpdateProjectCredential implements the v1.ServerInterface.
func (a *API) UpdateProjectCredential(ctx context.Context, request UpdateProjectCredentialRequestObject) (UpdateProjectCredentialResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return UpdateProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectCredential").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return UpdateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Show(
		ctx,
		parent.ID,
		request.CredentialId,
	)

	if err != nil {
		if errors.Is(err, store.ErrCredentialNotFound) {
			return UpdateProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectCredential").
			Str("project", parent.ID).
			Str("credential", request.CredentialId).
			Msg("Failed to load credential")

		return UpdateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageCredential(ctx, credentialPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return UpdateProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or credential"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if request.Body.Slug != nil {
		record.Slug = FromPtr(request.Body.Slug)
	}

	if request.Body.Name != nil {
		record.Name = FromPtr(request.Body.Name)
	}

	if request.Body.Kind != nil {
		record.Kind = FromPtr(request.Body.Kind)
	}

	if request.Body.Override != nil {
		record.Override = FromPtr(request.Body.Override)
	}

	switch record.Kind {
	case "shell":
		record.Login = model.CredentialLogin{}

		if request.Body.Shell != nil {
			if request.Body.Shell.Username != nil {
				record.Shell.Username = FromPtr(request.Body.Shell.Username)
			}

			if request.Body.Shell.Password != nil {
				record.Shell.Password = FromPtr(request.Body.Shell.Password)
			}

			if request.Body.Shell.PrivateKey != nil {
				record.Shell.PrivateKey = FromPtr(request.Body.Shell.PrivateKey)
			}
		}

		// TODO: drop this
		record.Shell.Password = "p455w0rd"
	case "login":
		record.Shell = model.CredentialShell{}

		if request.Body.Login != nil {
			if request.Body.Login.Username != nil {
				record.Login.Username = FromPtr(request.Body.Login.Username)
			}

			if request.Body.Login.Password != nil {
				record.Login.Password = FromPtr(request.Body.Login.Password)
			}
		}

		// TODO: drop this
		record.Login.Password = "p455w0rd"
	default:
		record.Shell = model.CredentialShell{}
		record.Login = model.CredentialLogin{}
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("action", "UpdateProjectCredential").
			Str("project", parent.ID).
			Str("credential", record.ID).
			Msg("Failed to encrypt secrets")

		return UpdateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to encrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Update(
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

			return UpdateProjectCredential422JSONResponse{ValidationErrorJSONResponse{
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Message: ToPtr("Failed to validate credential"),
				Errors:  ToPtr(errors),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "UpdateProjectCredential").
			Str("project", parent.ID).
			Str("credential", record.ID).
			Msg("Failed to update credential")

		return UpdateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to update credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	return UpdateProjectCredential200JSONResponse{ProjectCredentialResponseJSONResponse(
		a.convertCredential(record),
	)}, nil
}

// DeleteProjectCredential implements the v1.ServerInterface.
func (a *API) DeleteProjectCredential(ctx context.Context, request DeleteProjectCredentialRequestObject) (DeleteProjectCredentialResponseObject, error) {
	parent, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Projects.Show(
		ctx,
		request.ProjectId,
	)

	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			return DeleteProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectCredential").
			Str("project", request.ProjectId).
			Msg("Failed to load project")

		return DeleteProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load project"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	record, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Show(
		ctx,
		parent.ID,
		request.CredentialId,
	)

	if err != nil {
		if errors.Is(err, store.ErrCredentialNotFound) {
			return DeleteProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
				Message: ToPtr("Failed to find project or credential"),
				Status:  ToPtr(http.StatusNotFound),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("action", "DeleteProjectCredential").
			Str("project", parent.ID).
			Str("credential", request.CredentialId).
			Msg("Failed to load credential")

		return DeleteProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
			Message: ToPtr("Failed to load credential"),
			Status:  ToPtr(http.StatusInternalServerError),
		}}, nil
	}

	if !a.permitManageCredential(ctx, credentialPermissions{
		User:    current.GetUser(ctx),
		Project: parent,
		Record:  record,
	}) {
		return DeleteProjectCredential404JSONResponse{NotFoundErrorJSONResponse{
			Message: ToPtr("Failed to find project or credential"),
			Status:  ToPtr(http.StatusNotFound),
		}}, nil
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Credentials.Delete(
		ctx,
		parent.ID,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("action", "DeletProjectCredential").
			Str("project", parent.ID).
			Str("credential", record.ID).
			Msg("Failed to delete credential")

		return DeleteProjectCredential400JSONResponse{ActionFailedErrorJSONResponse{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete credential"),
		}}, nil
	}

	return DeleteProjectCredential200JSONResponse{SuccessMessageJSONResponse{
		Status:  ToPtr(http.StatusOK),
		Message: ToPtr("Successfully deleted credential"),
	}}, nil
}

func (a *API) convertCredential(record *model.Credential) Credential {
	result := Credential{
		Id:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Kind:      ToPtr(CredentialKind(record.Kind)),
		Override:  ToPtr(record.Override),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	switch record.Kind {
	case "shell":
		result.Shell = ToPtr(CredentialShell{
			Username: ToPtr(record.Shell.Username),
		})
	case "login":
		result.Login = ToPtr(CredentialLogin{
			Username: ToPtr(record.Shell.Username),
		})
	}

	return result
}

type credentialPermissions struct {
	User      *model.User
	Project   *model.Project
	Record    *model.Credential
	OwnerOnly bool
}

func (a *API) permitCreateCredential(ctx context.Context, definition credentialPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitShowCredential(ctx context.Context, definition credentialPermissions) bool {
	return a.permitShowProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func (a *API) permitManageCredential(ctx context.Context, definition credentialPermissions) bool {
	return a.permitManageProject(ctx, projectPermissions{
		User:   definition.User,
		Record: definition.Project,
	})
}

func listCredentialsSorting(request ListProjectCredentialsRequestObject) (string, string, int64, int64, string) {
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
