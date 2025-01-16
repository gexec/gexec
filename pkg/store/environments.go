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

// Environments provides all database operations related to environments.
type Environments struct {
	client *Store
}

// List implements the listing of all environments.
func (s *Environments) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Environment, int64, error) {
	records := make([]*model.Environment, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Secrets").
		Relation("Values").
		Where("environment.project_id = ?", projectID)

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
		Relation("Secrets").
		Relation("Values").
		Where("environment.project_id = ?", projectID).
		Where("environment.id = ? OR environment.slug = ?", name, name).
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

	return s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(record).
			Exec(ctx); err != nil {
			return err
		}

		for _, secret := range record.Secrets {
			secret.EnvironmentID = record.ID

			if _, err := tx.NewInsert().
				Model(secret).
				Exec(ctx); err != nil {
				return err
			}
		}

		for _, value := range record.Values {
			value.EnvironmentID = record.ID

			if _, err := tx.NewInsert().
				Model(value).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}

// Update implements the update of an existing environment.
func (s *Environments) Update(ctx context.Context, projectID string, record *model.Environment) error {
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

	return s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().
			Model(record).
			Where("id = ? and project_id = ?", record.ID, projectID).
			Exec(ctx); err != nil {
			return err
		}

		for _, secret := range record.Secrets { // TODO: broken for dropped rows
			secret.EnvironmentID = record.ID

			if secret.ID == "" {
				if _, err := tx.NewInsert().
					Model(secret).
					Exec(ctx); err != nil {
					return err
				}
			} else {
				current := &model.EnvironmentSecret{}

				if err := tx.NewSelect().
					Model(current).
					Where("id = ? AND environment_id = ?", secret.ID, secret.EnvironmentID).
					Scan(ctx); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return ErrEnvironmentSecretNotFound
					}

					return err
				}

				if secret.Kind != "" {
					current.Kind = secret.Kind
				}

				if secret.Name != "" {
					current.Name = secret.Name
				}

				if secret.Content != "" {
					current.Content = secret.Content
				}

				if _, err := tx.NewUpdate().
					Model(current).
					Where("id = ? and environment_id = ?", secret.ID, secret.EnvironmentID).
					Exec(ctx); err != nil {
					return err
				}
			}
		}

		for _, value := range record.Values { // TODO: broken for dropped rows
			value.EnvironmentID = record.ID

			if value.ID == "" {
				if _, err := tx.NewInsert().
					Model(value).
					Exec(ctx); err != nil {
					return err
				}
			} else {
				current := &model.EnvironmentValue{}

				if err := tx.NewSelect().
					Model(current).
					Where("id = ? AND environment_id = ?", value.ID, value.EnvironmentID).
					Scan(ctx); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return ErrEnvironmentValueNotFound
					}

					return err
				}

				if value.Kind != "" {
					current.Kind = value.Kind
				}

				if value.Name != "" {
					current.Name = value.Name
				}

				if value.Content != "" {
					current.Content = value.Content
				}

				if _, err := tx.NewUpdate().
					Model(current).
					Where("id = ? and environment_id = ?", value.ID, value.EnvironmentID).
					Exec(ctx); err != nil {
					return err
				}
			}
		}

		return nil
	})
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

func (s *Environments) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Environment)(nil)).
			Where("project_id = ?", projectID).
			Where("id = ?", val)

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

func (s *Environments) validate(ctx context.Context, record *model.Environment, _ bool) error {
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

	for i, secret := range record.Secrets {
		if err := validation.Validate(
			secret.Kind,
			validation.Required,
			validation.In("var", "env"),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("secrets.%d.name", i),
				Error: err,
			})
		}

		if err := validation.Validate(
			secret.Name,
			validation.Required,
			validation.Length(3, 255),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("secrets.%d.name", i),
				Error: err,
			})
		}
	}

	for i, value := range record.Values {
		if err := validation.Validate(
			value.Kind,
			validation.Required,
			validation.In("var", "env"),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("values.%d.name", i),
				Error: err,
			})
		}

		if err := validation.Validate(
			value.Name,
			validation.Required,
			validation.Length(3, 255),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("values.%d.name", i),
				Error: err,
			})
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Environments) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Environment)(nil)).
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

func (s *Environments) slugify(ctx context.Context, column, value, id, projectID string) string {
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
			Model((*model.Environment)(nil)).
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

func (s *Environments) validSort(val string) (string, bool) {
	if val == "" {
		return "environment.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"name":    "environment.name",
		"slug":    "environment.slug",
		"created": "environment.created_at",
		"updated": "environment.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "environment.name", true
}
