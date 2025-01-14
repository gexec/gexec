package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
)

// Environments provides all database operations related to environments.
type Environments struct {
	client *Store
}

// List implements the listing of all environments.
func (s *Environments) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Environment, int64, error) {
	records := make([]*model.Environment, 0)

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

// Show implements the details for a specific environment.
func (s *Environments) Show(ctx context.Context, projectID, name string) (*model.Environment, error) {
	record := &model.Environment{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Where("project_id = ?", projectID).
		Where("id = ? OR slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrEnvironmentNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new environment.
func (s *Environments) Create(ctx context.Context, projectID string, record *model.Environment) error {
	// if record.Slug == "" {
	// 	record.Slug = Slugify(
	// 		ctx,
	// 		s.client.handle.NewSelect().
	// 			Model((*model.Environment)(nil)),
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

// Update implements the update of an existing environment.
func (s *Environments) Update(ctx context.Context, projectID string, record *model.Environment) error {
	// if record.Slug == "" {
	// 	record.Slug = Slugify(
	// 		ctx,
	// 		s.client.handle.NewSelect().
	// 			Model((*model.Environment)(nil)),
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

// Delete implements the deletion of a environment.
func (s *Environments) Delete(ctx context.Context, projectID, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Environment)(nil)).
		Where("project_id = ?", projectID).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Environments) validate(ctx context.Context, record *model.Environment, _ bool) error {
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

func (s *Environments) validSort(val string) (string, bool) {
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
