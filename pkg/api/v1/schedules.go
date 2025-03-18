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

// ListProjectSchedules implements the v1.ServerInterface.
func (a *API) ListProjectSchedules(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectSchedulesParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listSchedulesSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.List(
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
			Str("action", "ListProjectSchedules").
			Msg("Failed to load schedules")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load schedules"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Schedule, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			log.Error().
				Err(err).
				Str("project", project.ID).
				Str("action", "ListProjectSchedules").
				Msg("Failed to decrypt secrets")

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertSchedule(record)
	}

	render.JSON(w, r, ProjectSchedulesResponse{
		Total:     count,
		Limit:     limit,
		Offset:    offset,
		Project:   ToPtr(a.convertProject(project)),
		Schedules: payload,
	})
}

// ShowProjectSchedule implements the v1.ServerInterface.
func (a *API) ShowProjectSchedule(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ScheduleID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("schedule", record.ID).
			Str("action", "ShowProjectSchedule").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectScheduleResponse(
		a.convertSchedule(record),
	))
}

// CreateProjectSchedule implements the v1.ServerInterface.
func (a *API) CreateProjectSchedule(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectScheduleBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectSchedule").
			Msg("Failed to decode request body")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Schedule{
		ProjectID: project.ID,
	}

	if body.TemplateID != nil {
		record.TemplateID = FromPtr(body.TemplateID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Cron != nil {
		record.Cron = FromPtr(body.Cron)
	}

	if body.Active != nil {
		record.Active = FromPtr(body.Active)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectSchedule").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Create(
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
				Message: ToPtr("Failed to validate schedule"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "CreateProjectSchedule").
			Msg("Failed to create schedule")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectScheduleResponse(
		a.convertSchedule(record),
	))
}

// UpdateProjectSchedule implements the v1.ServerInterface.
func (a *API) UpdateProjectSchedule(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ScheduleID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)
	body := &UpdateProjectScheduleBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("schedule", record.ID).
			Str("action", "UpdateProjectSchedule").
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
			Str("schedule", record.ID).
			Str("action", "UpdateProjectSchedule").
			Msg("Failed to decrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt credentials"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.TemplateID != nil {
		record.TemplateID = FromPtr(body.TemplateID)
	}

	if body.Slug != nil {
		record.Slug = FromPtr(body.Slug)
	}

	if body.Name != nil {
		record.Name = FromPtr(body.Name)
	}

	if body.Cron != nil {
		record.Cron = FromPtr(body.Cron)
	}

	if body.Active != nil {
		record.Active = FromPtr(body.Active)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("schedule", record.ID).
			Str("action", "UpdateProjectSchedule").
			Msg("Failed to encrypt secrets")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Update(
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
				Message: ToPtr("Failed to validate schedule"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("schedule", record.ID).
			Str("action", "UpdateProjectSchedule").
			Msg("Failed to update schedule")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update schedule"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectScheduleResponse(
		a.convertSchedule(record),
	))
}

// DeleteProjectSchedule implements the v1.ServerInterface.
func (a *API) DeleteProjectSchedule(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ScheduleID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectScheduleFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Schedules.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("schedule", record.ID).
			Str("action", "DeletProjectSchedule").
			Msg("Failed to delete schedule")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete schedule"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted schedule"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertSchedule(record *model.Schedule) Schedule {
	result := Schedule{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Cron:      ToPtr(record.Cron),
		Active:    ToPtr(record.Active),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.Template != nil {
		result.TemplateID = ToPtr(record.TemplateID)

		result.Template = ToPtr(
			a.convertTemplate(
				record.Template,
			),
		)
	}

	return result
}

// AllowShowProjectSchedule defines a middleware to check permissions.
func (a *API) AllowShowProjectSchedule(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectSchedule defines a middleware to check permissions.
func (a *API) AllowManageProjectSchedule(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listSchedulesSorting(request ListProjectSchedulesParams) (string, string, int64, int64, string) {
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
