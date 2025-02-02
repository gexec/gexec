package store

import (
	"context"

	"github.com/gexec/gexec/pkg/model"
	"github.com/uptrace/bun"
)

// Events provides all database operations related to events.
type Events struct {
	client *Store
}

// List implements the listing of all events.
func (s *Events) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Event, int64, error) {
	records := make([]*model.Event, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Order("event.created_at DESC")

	if projectID != "" {
		q = q.Where("event.project_id = ?", projectID)
	} else if !s.client.principal.Admin {
		q = q.Where(
			"event.project_id IN (?)",
			bun.In(s.client.Projects.AllowedIDs()),
		)
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}
