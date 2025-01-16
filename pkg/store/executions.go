package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Executions provides all database operations related to executions.
type Executions struct {
	client *Store
}

// List implements the listing of all executions.
func (s *Executions) List(ctx context.Context, projectID string, params model.ListParams) ([]*model.Execution, int64, error) {
	records := make([]*model.Execution, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Template").
		Where("execution.project_id = ?", projectID)

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

// Show implements the details for a specific execution.
func (s *Executions) Show(ctx context.Context, projectID, name string) (*model.Execution, error) {
	record := &model.Execution{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Template").
		Where("execution.project_id = ?", projectID).
		Where("execution.id = ?", name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrExecutionNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new execution.
func (s *Executions) Create(ctx context.Context, projectID string, record *model.Execution) error {
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

// Update implements the update of an existing execution.
func (s *Executions) Update(ctx context.Context, projectID string, record *model.Execution) error {
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

// Delete implements the deletion of a execution.
func (s *Executions) Delete(ctx context.Context, projectID, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Execution)(nil)).
		Where("project_id = ?", projectID).
		Where("id = ?", name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Executions) ValidateExists(ctx context.Context, projectID string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		if val == "" {
			return nil
		}

		q := s.client.handle.NewSelect().
			Model((*model.Execution)(nil)).
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

func (s *Executions) validate(ctx context.Context, record *model.Execution, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.TemplateID,
		validation.Required,
		validation.By(s.client.Templates.ValidateExists(ctx, record.ProjectID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "template_id",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Executions) validSort(val string) (string, bool) {
	if val == "" {
		return "execution.created_at", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"template": "template.name",
		"status":   "execution.status",
		"created":  "execution.created_at",
		"updated":  "execution.updated_at",
	} {
		if val == key {
			return name, true
		}
	}

	return "execution.created_at", true
}
