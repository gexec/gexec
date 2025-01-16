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

// Templates provides all database operations related to templates.
type Templates struct {
	client *Store
}

// List implements the listing of all templates.
func (s *Templates) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Template, int64, error) {
	records := make([]*model.Template, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Repository").
		Relation("Inventory").
		Relation("Environment").
		Relation("Surveys").
		Relation("Vaults").
		Where("template.project_id = ?", projectID)

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
		Relation("Repository").
		Relation("Inventory").
		Relation("Environment").
		Relation("Surveys").
		Relation("Vaults").
		Where("template.project_id = ?", projectID).
		Where("template.id = ? OR template.slug = ?", name, name).
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

		for _, survey := range record.Surveys {
			survey.TemplateID = record.ID

			if _, err := tx.NewInsert().
				Model(survey).
				Exec(ctx); err != nil {
				return err
			}

			for _, value := range survey.Values {
				value.SurveyID = survey.ID

				if _, err := tx.NewInsert().
					Model(value).
					Exec(ctx); err != nil {
					return err
				}
			}
		}

		for _, vault := range record.Vaults {
			vault.TemplateID = record.ID

			if _, err := tx.NewInsert().
				Model(vault).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}

// Update implements the update of an existing template.
func (s *Templates) Update(ctx context.Context, projectID string, record *model.Template) error {
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

		for _, survey := range record.Surveys { // TODO: broken for dropped rows
			survey.TemplateID = record.ID

			if survey.ID == "" {
				if _, err := tx.NewInsert().
					Model(survey).
					Exec(ctx); err != nil {
					return err
				}

				for _, value := range survey.Values {
					value.SurveyID = survey.ID

					if _, err := tx.NewInsert().
						Model(value).
						Exec(ctx); err != nil {
						return err
					}
				}
			} else {
				current := &model.TemplateSurvey{}

				if err := tx.NewSelect().
					Model(current).
					Where("id = ? AND template_id = ?", survey.ID, survey.TemplateID).
					Scan(ctx); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return ErrTemplateSurveyNotFound
					}

					return err
				}

				if survey.Name != "" {
					current.Name = survey.Name
				}

				if survey.Title != "" {
					current.Title = survey.Title
				}

				if survey.Description != "" {
					current.Description = survey.Description
				}

				if survey.Kind != "" {
					current.Kind = survey.Kind
				}

				if survey.Required != current.Required {
					current.Required = survey.Required
				}

				if _, err := tx.NewUpdate().
					Model(current).
					Where("id = ? and template_id = ?", survey.ID, survey.TemplateID).
					Exec(ctx); err != nil {
					return err
				}

				for _, value := range survey.Values {
					value.SurveyID = survey.ID

					if value.ID == "" {
						if _, err := tx.NewInsert().
							Model(value).
							Exec(ctx); err != nil {
							return err
						}
					} else {
						current := &model.TemplateValue{}

						if err := tx.NewSelect().
							Model(current).
							Where("id = ? AND survey_id = ?", value.ID, value.SurveyID).
							Scan(ctx); err != nil {
							if errors.Is(err, sql.ErrNoRows) {
								return ErrTemplateValueNotFound
							}

							return err
						}

						if value.Name != "" {
							current.Name = value.Name
						}

						if value.Value != "" {
							current.Value = value.Value
						}

						if _, err := tx.NewUpdate().
							Model(value).
							Where("id = ? and survey_id = ?", value.ID, value.SurveyID).
							Exec(ctx); err != nil {
							return err
						}
					}
				}
			}
		}

		for _, vault := range record.Vaults { // TODO: broken for dropped rows
			vault.TemplateID = record.ID

			if vault.ID == "" {
				if _, err := tx.NewInsert().
					Model(vault).
					Exec(ctx); err != nil {
					return err
				}
			} else {
				current := &model.TemplateVault{}

				if err := tx.NewSelect().
					Model(current).
					Where("id = ? AND template_id = ?", vault.ID, vault.TemplateID).
					Scan(ctx); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return ErrTemplateVaultNotFound
					}

					return err
				}

				if vault.CredentialID != "" {
					current.CredentialID = vault.CredentialID
				}

				if vault.Name != "" {
					current.Name = vault.Name
				}

				if vault.Kind != "" {
					current.Kind = vault.Kind
				}

				if vault.Script != "" {
					current.Script = vault.Script
				}

				if _, err := tx.NewUpdate().
					Model(current).
					Where("id = ? and template_id = ?", vault.ID, vault.TemplateID).
					Exec(ctx); err != nil {
					return err
				}
			}
		}

		return nil
	})
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

func (s *Templates) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Template)(nil)).
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

func (s *Templates) validate(ctx context.Context, record *model.Template, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.RepositoryID,
		validation.Required,
		validation.By(s.client.Repositories.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "repository_id",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.InventoryID,
		validation.By(s.client.Inventories.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "inventory_id",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.EnvironmentID,
		validation.Required,
		validation.By(s.client.Environments.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "environment_id",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Executor,
		validation.Required,
		validation.Length(3, 255),
		validation.In("ansible", "terraform", "opentofu", "asdf"),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "executor",
			Error: err,
		})
	}

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
		record.Playbook,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "playbook",
			Error: err,
		})
	}

	for i, survey := range record.Surveys {
		if err := validation.Validate(
			survey.Kind,
			validation.Required,
			validation.In("string", "number", "enum", "secret"),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("surveys.%d.name", i),
				Error: err,
			})
		}

		if err := validation.Validate(
			survey.Name,
			validation.Required,
			validation.Length(3, 255),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("surveys.%d.name", i),
				Error: err,
			})
		}

		for x, value := range survey.Values {
			if err := validation.Validate(
				value.Name,
				validation.Required,
				validation.Length(1, 255),
			); err != nil {
				errs.Errors = append(errs.Errors, validate.Error{
					Field: fmt.Sprintf("surveys.%d.values.%d.name", i, x),
					Error: err,
				})
			}

			if err := validation.Validate(
				value.Value,
				validation.Required,
				validation.Length(1, 255),
			); err != nil {
				errs.Errors = append(errs.Errors, validate.Error{
					Field: fmt.Sprintf("surveys.%d.values.%d.value", i, x),
					Error: err,
				})
			}
		}
	}

	for i, vault := range record.Vaults {
		if err := validation.Validate(
			vault.CredentialID,
			validation.By(s.client.Credentials.ValidateExists(ctx, record.ProjectID)),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("vaults.%d.credential_id", i),
				Error: err,
			})
		}

		if err := validation.Validate(
			vault.Kind,
			validation.Required,
			validation.In("password", "script"),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("vaults.%d.name", i),
				Error: err,
			})
		}

		if err := validation.Validate(
			vault.Name,
			validation.Required,
			validation.Length(3, 255),
		); err != nil {
			errs.Errors = append(errs.Errors, validate.Error{
				Field: fmt.Sprintf("vaults.%d.name", i),
				Error: err,
			})
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Templates) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Template)(nil)).
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

func (s *Templates) slugify(ctx context.Context, column, value, id, projectID string) string {
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
			Model((*model.Template)(nil)).
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

func (s *Templates) validSort(val string) (string, bool) {
	if val == "" {
		return "template.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"repository":  "repository.name",
		"inventory":   "inventory.name",
		"environment": "environment.name",
		"name":        "template.name",
		"slug":        "template.slug",
		"created":     "template.created_at",
		"updated":     "template.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "template.name", true
}
