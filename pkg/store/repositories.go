package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/dchest/uniuri"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Repositories provides all database operations related to repositories.
type Repositories struct {
	client *Store
}

// List implements the listing of all repositories.
func (s *Repositories) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Repository, int64, error) {
	records := make([]*model.Repository, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Credential").
		Where("repository.project_id = ?", projectID)

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

// Show implements the details for a specific repository.
func (s *Repositories) Show(ctx context.Context, projectID, name string) (*model.Repository, error) {
	record := &model.Repository{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Credential").
		Where("repository.project_id = ?", projectID).
		Where("repository.id = ? OR repository.slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrRepositoryNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new repository.
func (s *Repositories) Create(ctx context.Context, projectID string, record *model.Repository) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
			projectID,
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

	return nil
}

// Update implements the update of an existing repository.
func (s *Repositories) Update(ctx context.Context, projectID string, record *model.Repository) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
			projectID,
		)
	}

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

// Delete implements the deletion of a repository.
func (s *Repositories) Delete(ctx context.Context, projectID, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Repository)(nil)).
		Where("project_id = ?", projectID).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
func (s *Repositories) validate(ctx context.Context, record *model.Repository, _ bool) error {
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
		record.CredentialID,
		validation.Required,
		validation.By(s.client.Credentials.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "credential_id",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.URL,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "url",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Branch,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "branch",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Repositories) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Repository)(nil)).
			Where("project_id = ? AND ? = ?", projectID, bun.Ident(key), val)

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

func (s *Repositories) slugify(ctx context.Context, column, value, id, projectID string) string {
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
			Model((*model.Repository)(nil)).
			Where("project_id = ? AND ? = ?", projectID, bun.Ident(column), slug)

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

func (s *Repositories) validSort(val string) (string, bool) {
	if val == "" {
		return "repository.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"name":   "repository.name",
		"slug":   "repository.slug",
		"url":    "repository.url",
		"branch": "repository.branch",
	} {
		if val == key {
			return name, true
		}
	}

	return "repository.name", true
}
