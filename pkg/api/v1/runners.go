package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/secret"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// ListProjectRunners implements the v1.ServerInterface.
func (a *API) ListProjectRunners(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectRunnersParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listRunnersSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.List(
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
			Str("action", "ListProjectRunners").
			Msg("Failed to load runners")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load runners"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Runner, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			log.Error().
				Err(err).
				Str("project", project.ID).
				Str("action", "ListProjectRunners").
				Msg("Failed to decrypt secrets")

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertRunner(record)
	}

	render.JSON(w, r, ProjectRunnersResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(project)),
		Runners: payload,
	})
}

// ShowProjectRunner implements the v1.ServerInterface.
func (a *API) ShowProjectRunner(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RunnerID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectRunnerFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("runner", record.ID).
			Str("action", "ShowProjectRunner").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRunnerResponse(
		a.convertRunner(record),
	))
}

// CreateProjectRunner implements the v1.ServerInterface.
func (a *API) CreateProjectRunner(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectRunnerBody{}

	record := &model.Runner{
		ProjectID: project.ID,
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Token != nil {
		record.Token = FromPtr(body.Token)
	}

	if record.Token == "" {
		record.Token = secret.Generate(32)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectRunner").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.Create(
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
				Message: ToPtr("Failed to validate runner"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectRunner").
			Msg("Failed to create runner")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create runner"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRunnerResponse(
		a.convertRunner(record),
	))
}

// UpdateProjectRunner implements the v1.ServerInterface.
func (a *API) UpdateProjectRunner(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RunnerID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectRunnerFromContext(ctx)
	body := &UpdateProjectRunnerBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("runner", record.ID).
			Str("action", "UpdateProjectRunner").
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
			Str("runner", record.ID).
			Str("action", "UpdateProjectRunner").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt runners"),
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

	if body.Token != nil {
		record.Token = FromPtr(body.Token)
	}

	if record.Token == "" {
		record.Token = secret.Generate(32)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("runner", record.ID).
			Str("action", "UpdateProjectRunner").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.Update(
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
				Message: ToPtr("Failed to validate runner"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("runner", record.ID).
			Str("action", "UpdateProjectRunner").
			Msg("Failed to update runner")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update runner"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectRunnerResponse(
		a.convertRunner(record),
	))
}

// DeleteProjectRunner implements the v1.ServerInterface.
func (a *API) DeleteProjectRunner(w http.ResponseWriter, r *http.Request, _ ProjectID, _ RunnerID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("runner", record.ID).
			Str("action", "DeletProjectRunner").
			Msg("Failed to delete runner")

		a.RenderNotify(w, r, Notification{
			Status:  ToPtr(http.StatusBadRequest),
			Message: ToPtr("Failed to delete runner"),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted runner"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertRunner(record *model.Runner) Runner {
	result := Runner{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Token:     ToPtr(record.Token),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	return result
}

// AllowShowProjectRunner defines a middleware to check permissions.
func (a *API) AllowShowProjectRunner(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectRunner defines a middleware to check permissions.
func (a *API) AllowManageProjectRunner(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listRunnersSorting(request ListProjectRunnersParams) (string, string, int64, int64, string) {
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
