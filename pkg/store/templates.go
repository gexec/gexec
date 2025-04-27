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
func (s *Templates) Show(ctx context.Context, project *model.Project, name string) (*model.Template, error) {
	record := &model.Template{}

	q := s.client.handle.NewSelect().
		Model(record).
		Relation("Project").
		Relation("Repository").
		Relation("Inventory").
		Relation("Environment").
		Relation("Surveys").
		Relation("Vaults").
		Where("template.project_id = ?", project.ID).
		Where("template.id = ? OR template.slug = ?", name, name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTemplateNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new template.
func (s *Templates) Create(ctx context.Context, project *model.Project, record *model.Template) (*model.Template, error) {
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

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
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
	}); err != nil {
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
				ObjectType:     model.EventTypeTemplate,
				Action:         model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, project, record.ID)
}

// Update implements the update of an existing template.
func (s *Templates) Update(ctx context.Context, project *model.Project, record *model.Template) (*model.Template, error) {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
			project.ID,
		)
	}

	if err := s.validate(ctx, record, true); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		q := tx.NewUpdate().
			Model(record).
			Where("project_id = ?", project.ID).
			Where("id = ?", record.ID)

		if _, err := q.Exec(ctx); err != nil {
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

				q := tx.NewSelect().
					Model(current).
					Where("template_id = ?", survey.TemplateID).
					Where("id = ?", survey.ID)

				if err := q.Scan(ctx); err != nil {
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

				up := tx.NewUpdate().
					Model(current).
					Where("template_id = ?", survey.TemplateID).
					Where("id = ?", survey.ID)

				if _, err := up.Exec(ctx); err != nil {
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

						q := tx.NewSelect().
							Model(current).
							Where("survey_id = ?", value.SurveyID).
							Where("id = ?", value.ID)

						if err := q.Scan(ctx); err != nil {
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

						up := tx.NewUpdate().
							Model(value).
							Where("survey_id = ?", value.SurveyID).
							Where("id = ?", value.ID)

						if _, err := up.Exec(ctx); err != nil {
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

				q := tx.NewSelect().
					Model(current).
					Where("template_id = ?", vault.TemplateID).
					Where("id = ?", vault.ID)

				if err := q.Scan(ctx); err != nil {
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

				up := tx.NewUpdate().
					Model(current).
					Where("template_id = ?", vault.TemplateID).
					Where("id = ?", vault.ID)

				if _, err := up.Exec(ctx); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
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
				ObjectType:     model.EventTypeTemplate,
				Action:         model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, project, record.ID)
}

// Delete implements the deletion of a template.
func (s *Templates) Delete(ctx context.Context, project *model.Project, name string) error {
	record, err := s.Show(ctx, project, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.Template)(nil)).
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
				ObjectType:     model.EventTypeTemplate,
				Action:         model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ShowSurvey implements the details for a specific template survey.
func (s *Templates) ShowSurvey(ctx context.Context, template *model.Template, name string) (*model.TemplateSurvey, error) {
	record := &model.TemplateSurvey{}

	q := s.client.handle.NewSelect().
		Model(record).
		Relation("Values").
		Where("template_survey.template_id = ?", template.ID).
		Where("template_survey.id = ?", name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTemplateSurveyNotFound
		}

		return record, err
	}

	return record, nil
}

// CreateSurvey implements the create of a new template survey.
func (s *Templates) CreateSurvey(ctx context.Context, template *model.Template, record *model.TemplateSurvey) (*model.TemplateSurvey, error) {
	if err := s.validateSurvey(ctx, template, record, false); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(record).
			Exec(ctx); err != nil {
			return err
		}

		for _, row := range record.Values {
			row.SurveyID = record.ID

			if _, err := tx.NewInsert().
				Model(row).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateSurvey,
				Action:         model.EventActionCreate,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.ShowSurvey(ctx, template, record.ID)
}

// UpdateSurvey implements the update of an existing template survey.
func (s *Templates) UpdateSurvey(ctx context.Context, template *model.Template, record *model.TemplateSurvey) (*model.TemplateSurvey, error) {
	if err := s.validateSurvey(ctx, template, record, true); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		q := tx.NewUpdate().
			Model(record).
			Where("template_id = ?", template.ID).
			Where("id = ?", record.ID)

		if _, err := q.Exec(ctx); err != nil {
			return err
		}

		for _, row := range record.Values { // TODO: broken for dropped rows
			row.SurveyID = record.ID

			if _, err := tx.NewUpdate().
				Model(row).
				Where("survey_id = ?", record.ID).
				Where("id = ?", row.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateSurvey,
				Action:         model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.ShowSurvey(ctx, template, record.ID)
}

// DeleteSurvey implements the deletion of a template survey.
func (s *Templates) DeleteSurvey(ctx context.Context, template *model.Template, name string) error {
	record, err := s.ShowSurvey(ctx, template, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.TemplateSurvey)(nil)).
		Where("template_id = ?", template.ID).
		Where("id = ?", record.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateSurvey,
				Action:         model.EventActionDelete,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ShowVault implements the details for a specific template vault.
func (s *Templates) ShowVault(ctx context.Context, template *model.Template, name string) (*model.TemplateVault, error) {
	record := &model.TemplateVault{}

	q := s.client.handle.NewSelect().
		Model(record).
		Where("template_id = ?", template.ID).
		Where("id = ?", name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTemplateVaultNotFound
		}

		return record, err
	}

	return record, nil
}

// CreateVault implements the create of a new template vault.
func (s *Templates) CreateVault(ctx context.Context, template *model.Template, record *model.TemplateVault) (*model.TemplateVault, error) {
	if err := s.validateVault(ctx, template, record, false); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(record).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateVault,
				Action:         model.EventActionCreate,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.ShowVault(ctx, template, record.ID)
}

// UpdateVault implements the update of an existing template vault.
func (s *Templates) UpdateVault(ctx context.Context, template *model.Template, record *model.TemplateVault) (*model.TemplateVault, error) {
	if err := s.validateVault(ctx, template, record, true); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		q := tx.NewUpdate().
			Model(record).
			Where("template_id = ?", template.ID).
			Where("id = ?", record.ID)

		if _, err := q.Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateVault,
				Action:         model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.ShowVault(ctx, template, record.ID)
}

// DeleteVault implements the deletion of a template vault.
func (s *Templates) DeleteVault(ctx context.Context, template *model.Template, name string) error {
	record, err := s.ShowVault(ctx, template, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.TemplateVault)(nil)).
		Where("template_id = ?", template.ID).
		Where("id = ?", record.ID)

	if _, err := q.Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      template.Project.ID,
				ProjectDisplay: template.Project.Name,
				ObjectID:       record.ID,
				ObjectDisplay:  record.Name,
				ObjectType:     model.EventTypeTemplateVault,
				Action:         model.EventActionDelete,
				Attrs: map[string]interface{}{
					"template_id":      template.ID,
					"template_display": template.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ValidateExists simply provides a validator for this record type.
func (s *Templates) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Template)(nil)).
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
		record.Path,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "path",
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

func (s *Templates) validateSurvey(_ context.Context, _ *model.Template, record *model.TemplateSurvey, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Name,
		validation.Required,
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "name",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Kind,
		validation.Required,
		validation.Length(3, 255),
		validation.In("string", "number", "enum", "secret"),
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

func (s *Templates) validateVault(ctx context.Context, template *model.Template, record *model.TemplateVault, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Name,
		validation.Required,
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "name",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Kind,
		validation.Required,
		validation.Length(3, 255),
		validation.In("password", "script"),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "kind",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.CredentialID,
		validation.By(s.client.Environments.ValidateExists(ctx, template.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "credential_id",
			Error: err,
		})
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
