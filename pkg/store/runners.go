package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/dchest/uniuri"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Runners provides all database operations related to runners.
type Runners struct {
	client *Store
}

// List implements the listing of all runners.
func (s *Runners) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Runner, int64, error) {
	records := make([]*model.Runner, 0)

	q := s.client.handle.NewSelect().
		Model(&records)

	if projectID != "" {
		q = q.Where("runner.project_id = ?", projectID)
	}

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

// Show implements the details for a specific runner.
func (s *Runners) Show(ctx context.Context, project *model.Project, name string) (*model.Runner, error) {
	record := &model.Runner{}

	q := s.client.handle.NewSelect().
		Model(record).
		Where("runner.id = ? OR runner.slug = ?", name, name)

	if project.ID != "" {
		q = q.Where("runner.project_id = ?", project.ID)
	}

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrRunnerNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new runner.
func (s *Runners) Create(ctx context.Context, project *model.Project, record *model.Runner) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
			project.ID,
		)
	}

	if err := s.validate(ctx, record, false); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeRunner,
				Action:         model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Update implements the update of an existing runner.
func (s *Runners) Update(ctx context.Context, project *model.Project, record *model.Runner) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			record.ID,
			project.ID,
		)
	}

	if err := s.validate(ctx, record, true); err != nil {
		return err
	}

	q := s.client.handle.NewUpdate().
		Model(record).
		Where("id = ?", record.ID)

	if project.ID != "" {
		q = q.Where("project_id = ?", project.ID)
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeRunner,
				Action:         model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Delete implements the deletion of a runner.
func (s *Runners) Delete(ctx context.Context, project *model.Project, name string) error {
	record, err := s.Show(ctx, project, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.Runner)(nil)).
		Where("id = ? OR slug = ?", name, name)

	if project.ID != "" {
		q = q.Where("project_id = ?", project.ID)
	}

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeRunner,
				Action:         model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ValidateExists simply provides a validator for this record type.
func (s *Runners) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Runner)(nil)).
			Where("id = ?", val)

		if projectID != "" {
			q = q.Where("project_id = ?", projectID)
		}

		exists, err := q.Exists(ctx)

		if err != nil {
			return err
		}

		if !exists {
			return errors.New("does not exist")
		}

		return nil
	}
}

func (s *Runners) validate(ctx context.Context, record *model.Runner, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Slug,
		validation.Required,
		validation.Length(3, 255),
		validation.By(s.uniqueValueIsPresent(ctx, "slug", record.ID, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "slug",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Name,
		validation.Required,
		validation.Length(3, 255),
		validation.By(s.uniqueValueIsPresent(ctx, "name", record.ID, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "name",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Token,
		validation.Required,
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "token",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Runners) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Runner)(nil)).
			Where("? = ?", bun.Ident(key), val)

		if projectID != "" {
			q = q.Where("project_id = ?", projectID)
		}

		if id != "" {
			q = q.Where(
				"id != ?",
				id,
			)
		}

		exists, err := q.Exists(ctx)

		if err != nil {
			return err
		}

		if exists {
			return errors.New("is already taken")
		}

		return nil
	}
}

func (s *Runners) slugify(ctx context.Context, column, value, id, projectID string) string {
	var (
		slug string
	)

	for i := 0; true; i++ {
		if i == 0 {
			slug = slugify.Slugify(value)
		} else {
			slug = slugify.Slugify(
				fmt.Sprintf("%s-%s", value, uniuri.NewLen(6)),
			)
		}

		query := s.client.handle.NewSelect().
			Model((*model.Runner)(nil)).
			Where("? = ?", bun.Ident(column), slug)

		if projectID != "" {
			query = query.Where("project_id = ?", projectID)
		}

		if id != "" {
			query = query.Where(
				"id != ?",
				id,
			)
		}

		if count, err := query.Count(
			ctx,
		); err == nil && count == 0 {
			break
		}
	}

	return slug
}

func (s *Runners) validSort(val string) (string, bool) {
	if val == "" {
		return "runner.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"name":    "runner.name",
		"slug":    "runner.slug",
		"token":   "runner.token",
		"created": "runner.created_at",
		"updated": "runner.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "runner.name", true
}
