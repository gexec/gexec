package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
)

// Templates provides all database operations related to templates.
type Templates struct {
	client *Store
}

// List implements the listing of all templates.
func (s *Templates) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Template, int64, error) {
	records := make([]*model.Template, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Where("project_id = ?", projectID)

	if val, ok := s.validSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
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

// Show implements the details for a specific template.
func (s *Templates) Show(ctx context.Context, projectID, name string) (*model.Template, error) {
	record := &model.Template{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Where("project_id = ?", projectID).
		Where("id = ? OR slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTemplateNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new template.
func (s *Templates) Create(ctx context.Context, projectID string, record *model.Template) error {
	// if record.Slug == "" {
	// 	record.Slug = Slugify(
	// 		ctx,
	// 		s.client.handle.NewSelect().
	// 			Model((*model.Template)(nil)),
	// 		"slug",
	// 		record.Name,
	// 		"",
	// 		false,
	// 	)
	// }

	if err := s.validate(ctx, record, false); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Update implements the update of an existing template.
func (s *Templates) Update(ctx context.Context, projectID string, record *model.Template) error {
	// if record.Slug == "" {
	// 	record.Slug = Slugify(
	// 		ctx,
	// 		s.client.handle.NewSelect().
	// 			Model((*model.Template)(nil)),
	// 		"slug",
	// 		record.Name,
	// 		record.ID,
	// 		false,
	// 	)
	// }

	if err := s.validate(ctx, record, true); err != nil {
		return err
	}

	if _, err := s.client.handle.NewUpdate().
		Model(record).
		Where("id = ? and project_id = ?", record.ID, projectID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Delete implements the deletion of a template.
func (s *Templates) Delete(ctx context.Context, projectID, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Template)(nil)).
		Where("project_id = ?", projectID).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Templates) validate(ctx context.Context, record *model.Template, _ bool) error {
	errs := validate.Errors{}

	// if err := validation.Validate(
	// 	record.Slug,
	// 	validation.Required,
	// 	validation.Length(3, 255),
	// 	validation.By(s.uniqueValueIsPresent(ctx, "slug", record.ID)),
	// ); err != nil {
	// 	errs.Errors = append(errs.Errors, validate.Error{
	// 		Field: "slug",
	// 		Error: err,
	// 	})
	// }

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Templates) validSort(val string) (string, bool) {
	if val == "" {
		return "foobar", true
	}

	val = strings.ToLower(val)

	for _, name := range []string{
		"foobar",
	} {
		if val == name {
			return val, true
		}
	}

	return "foobar", true
}
