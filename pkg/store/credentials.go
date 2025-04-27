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

// Credentials provides all database operations related to credentials.
type Credentials struct {
	client *Store
}

// List implements the listing of all credentials.
func (s *Credentials) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Credential, int64, error) {
	records := make([]*model.Credential, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Where("credential.project_id = ?", projectID)

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

// Show implements the details for a specific credential.
func (s *Credentials) Show(ctx context.Context, project *model.Project, name string) (*model.Credential, error) {
	record := &model.Credential{}

	q := s.client.handle.NewSelect().
		Model(record).
		Where("credential.project_id = ?", project.ID).
		Where("credential.id = ? OR credential.slug = ?", name, name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrCredentialNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new credential.
func (s *Credentials) Create(ctx context.Context, project *model.Project, record *model.Credential) (*model.Credential, error) {
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
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeCredential,
				Action:         model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, project, record.ID)
}

// Update implements the update of an existing credential.
func (s *Credentials) Update(ctx context.Context, project *model.Project, record *model.Credential) (*model.Credential, error) {
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
		return nil, err
	}

	q := s.client.handle.NewUpdate().
		Model(record).
		Where("project_id = ?", project.ID).
		Where("id = ?", record.ID)

	if _, err := q.Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeCredential,
				Action:         model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, project, record.ID)
}

// Delete implements the deletion of a credential.
func (s *Credentials) Delete(ctx context.Context, project *model.Project, name string) error {
	record, err := s.Show(ctx, project, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.Credential)(nil)).
		Where("project_id = ?", project.ID).
		Where("id = ? OR slug = ?", name, name)

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
				ObjectType:     model.EventTypeCredential,
				Action:         model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ValidateExists simply provides a validator for this record type.
func (s *Credentials) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Credential)(nil)).
			Where("project_id = ?", projectID).
			Where("id = ? OR slug = ?", val, val)

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

func (s *Credentials) validate(ctx context.Context, record *model.Credential, _ bool) error {
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
		record.Kind,
		validation.Required,
		validation.In("empty", "shell", "login"),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "kind",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Credentials) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Credential)(nil)).
			Where("project_id = ?", projectID).
			Where("? = ?", bun.Ident(key), val)

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

func (s *Credentials) slugify(ctx context.Context, column, value, id, projectID string) string {
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
			Model((*model.Credential)(nil)).
			Where("project_id = ?", projectID).
			Where("? = ?", bun.Ident(column), slug)

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

func (s *Credentials) validSort(val string) (string, bool) {
	if val == "" {
		return "credential.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"name":    "credential.name",
		"slug":    "credential.slug",
		"kind":    "credential.url",
		"created": "credential.created_at",
		"updated": "credential.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "credential.name", true
}
