package v1

import (
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

// ListProjectEvents implements the v1.ServerInterface.
func (a *API) ListProjectEvents(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectEventsParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
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
		project.ID,
		model.ListParams{
			Limit:  limit,
			Offset: offset,
			Search: search,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("project", project.ID).
			Str("action", "ListProjectEvents").
			Msg("Failed to load events")

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load events"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Event, len(records))
	for id, record := range records {
		payload[id] = a.convertEvent(record)
	}

	render.JSON(w, r, ProjectEventsResponse{
		Total:   count,
		Limit:   limit,
		Offset:  offset,
		Project: ToPtr(a.convertProject(project)),
		Events:  payload,
	})
}

func (a *API) convertEvent(record *model.Event) Event {
	result := Event{
		UserID:        ToPtr(record.UserID),
		UserHandle:    ToPtr(record.UserHandle),
		UserDisplay:   ToPtr(record.UserDisplay),
		ObjectID:      ToPtr(record.ObjectID),
		ObjectDisplay: ToPtr(record.ObjectDisplay),
		ObjectType:    ToPtr(EventObjectType(record.ObjectType)),
		Action:        ToPtr(EventAction(record.Action)),
		CreatedAt:     ToPtr(record.CreatedAt),
	}

	if record.Attrs != nil {
		result.Attrs = ToPtr(record.Attrs)
	}

	return result
}
