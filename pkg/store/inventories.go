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

// Inventories provides all database operations related to inventories.
type Inventories struct {
	client *Store
}

// List implements the listing of all inventories.
func (s *Inventories) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Inventory, int64, error) {
	records := make([]*model.Inventory, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Repository").
		Relation("Credential").
		Relation("Become").
		Where("inventory.project_id = ?", projectID)

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

// Show implements the details for a specific inventory.
func (s *Inventories) Show(ctx context.Context, project *model.Project, name string) (*model.Inventory, error) {
	record := &model.Inventory{}

	q := s.client.handle.NewSelect().
		Model(record).
		Relation("Repository").
		Relation("Credential").
		Relation("Become").
		Where("inventory.project_id = ?", project.ID).
		Where("inventory.id = ? OR inventory.slug = ?", name, name)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrInventoryNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new inventory.
func (s *Inventories) Create(ctx context.Context, project *model.Project, record *model.Inventory) error {
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
				ObjectType:     model.EventTypeInventory,
				Action:         model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Update implements the update of an existing inventory.
func (s *Inventories) Update(ctx context.Context, project *model.Project, record *model.Inventory) error {
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
		return err
	}

	q := s.client.handle.NewUpdate().
		Model(record).
		Where("project_id = ?", project.ID).
		Where("id = ?", record.ID)

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
				ObjectType:     model.EventTypeInventory,
				Action:         model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Delete implements the deletion of a inventory.
func (s *Inventories) Delete(ctx context.Context, project *model.Project, name string) error {
	record, err := s.Show(ctx, project, name)

	if err != nil {
		return err
	}

	q := s.client.handle.NewDelete().
		Model((*model.Inventory)(nil)).
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
				ObjectType:     model.EventTypeInventory,
				Action:         model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ValidateExists simply provides a validator for this record type.
func (s *Inventories) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Inventory)(nil)).
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

func (s *Inventories) validate(ctx context.Context, record *model.Inventory, _ bool) error {
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
		validation.In("static", "file", "workspace"),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "kind",
			Error: err,
		})
	}

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
		record.CredentialID,
		validation.By(s.client.Credentials.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "credential_id",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.BecomeID,
		validation.By(s.client.Credentials.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "become_id",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Inventories) uniqueValueIsPresent(ctx context.Context, key, id, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Inventory)(nil)).
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

func (s *Inventories) slugify(ctx context.Context, column, value, id, projectID string) string {
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
			Model((*model.Inventory)(nil)).
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

func (s *Inventories) validSort(val string) (string, bool) {
	if val == "" {
		return "inventory.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"repository": "repository.name",
		"credential": "credential.name",
		"become":     "become.name",
		"name":       "inventory.name",
		"slug":       "inventory.slug",
		"kind":       "inventory.kind",
		"created":    "inventory.created_at",
		"updated":    "inventory.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "inventory.name", true
}
