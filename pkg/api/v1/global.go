package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/secret"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
)

// ListGlobalEvents implements the v1.ServerInterface.
func (a *API) ListGlobalEvents(w http.ResponseWriter, r *http.Request, params ListGlobalEventsParams) {
	ctx := r.Context()
	_, limit, offset, search := toPageParams(
		nil,
		params.Limit,
		params.Offset,
		params.Search,
	)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Events.List(
		ctx,
		"",
		model.ListParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		},
	)

	if err != nil {
		slog.Error(
			"Failed to load events",
			slog.Any("error", err),
			slog.String("action", "ListGlobalEvents"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load events"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Event, len(records))
	for id, record := range records {
		payload[id] = a.convertGlobalEvent(record)
	}

	render.JSON(w, r, GlobalEventsResponse{
		Total:  count,
		Limit:  limit,
		Offset: offset,
		Events: payload,
	})
}

// ListGlobalRunners implements the v1.ServerInterface.
func (a *API) ListGlobalRunners(w http.ResponseWriter, r *http.Request, params ListGlobalRunnersParams) {
	ctx := r.Context()
	sort, limit, offset, search := toPageParams(
		params.Sort,
		params.Limit,
		params.Offset,
		params.Search,
	)

	order := ""

	if params.Order != nil {
		sort = string(FromPtr(params.Order))
	}

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.List(
		ctx,
		"",
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
			"Failed to load runners",
			slog.Any("error", err),
			slog.String("action", "ListGlobalRunners"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load runners"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Runner, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			slog.Error(
				"Failed to decrypt secrets",
				slog.Any("error", err),
				slog.String("action", "ListGlobalRunners"),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertGlobalRunner(record)
	}

	render.JSON(w, r, GlobalRunnersResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Runners: payload,
	})
}

// ShowGlobalRunner implements the v1.ServerInterface.
func (a *API) ShowGlobalRunner(w http.ResponseWriter, r *http.Request, _ RunnerID) {
	ctx := r.Context()
	record := a.GlobalRunnerFromContext(ctx)

	render.JSON(w, r, GlobalRunnerResponse(
		a.convertGlobalRunner(record),
	))
}

// CreateGlobalRunner implements the v1.ServerInterface.
func (a *API) CreateGlobalRunner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := &CreateGlobalRunnerBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("action", "CreateGlobalRunner"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Runner{}

	if body.ProjectID != nil {
		record.ProjectID = FromPtr(body.ProjectID)
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
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("action", "CreateGlobalRunner"),
		)

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
		&model.Project{},
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

		slog.Error(
			"Failed to create runner",
			slog.Any("error", err),
			slog.String("action", "CreateGlobalRunner"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create runner"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, GlobalRunnerResponse(
		a.convertGlobalRunner(record),
	))
}

// UpdateGlobalRunner implements the v1.ServerInterface.
func (a *API) UpdateGlobalRunner(w http.ResponseWriter, r *http.Request, _ RunnerID) {
	ctx := r.Context()
	record := a.GlobalRunnerFromContext(ctx)
	body := &UpdateGlobalRunnerBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("action", "UpdateGlobalRunner"),
			slog.String("runner", record.ID),
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
			slog.String("action", "UpdateGlobalRunner"),
			slog.String("runner", record.ID),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt runners"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if body.ProjectID != nil {
		record.ProjectID = FromPtr(body.ProjectID)
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
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("action", "UpdateGlobalRunner"),
			slog.String("runner", record.ID),
		)

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
		&model.Project{},
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

		slog.Error(
			"Failed to update runner",
			slog.Any("error", err),
			slog.String("action", "UpdateGlobalRunner"),
			slog.String("runner", record.ID),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update runner"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, GlobalRunnerResponse(
		a.convertGlobalRunner(record),
	))
}

// DeleteGlobalRunner implements the v1.ServerInterface.
func (a *API) DeleteGlobalRunner(w http.ResponseWriter, r *http.Request, _ RunnerID) {
	ctx := r.Context()
	record := a.UserFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Runners.Delete(
		ctx,
		&model.Project{},
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete runner",
			slog.Any("error", err),
			slog.String("action", "DeletGlobalRunner"),
			slog.String("runner", record.ID),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete runner"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted runner"),
		Status:  ToPtr(http.StatusOK),
	})
}

func (a *API) convertGlobalEvent(record *model.Event) Event {
	result := Event{
		UserID:         ToPtr(record.UserID),
		UserHandle:     ToPtr(record.UserHandle),
		UserDisplay:    ToPtr(record.UserDisplay),
		ProjectID:      ToPtr(record.ProjectID),
		ProjectDisplay: ToPtr(record.ProjectDisplay),
		ObjectID:       ToPtr(record.ObjectID),
		ObjectDisplay:  ToPtr(record.ObjectDisplay),
		ObjectType:     ToPtr(EventObjectType(record.ObjectType)),
		Action:         ToPtr(EventAction(record.Action)),
		CreatedAt:      ToPtr(record.CreatedAt),
	}

	if record.Attrs != nil {
		result.Attrs = ToPtr(record.Attrs)
	}

	return result
}

func (a *API) convertGlobalRunner(record *model.Runner) Runner {
	result := Runner{
		ID:        ToPtr(record.ID),
		Slug:      ToPtr(record.Slug),
		Name:      ToPtr(record.Name),
		Token:     ToPtr(record.Token),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.ProjectID != "" {
		result.ProjectID = ToPtr(record.ProjectID)

		result.Project = ToPtr(
			a.convertProject(
				record.Project,
			),
		)
	}

	return result
}
