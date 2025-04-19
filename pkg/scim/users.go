package scim

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	serrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/secret"
	"github.com/scim2/filter-parser/v2"
	"github.com/uptrace/bun"
)

var (
	userAttributeMapping = map[string]string{
		"userName":    "username",
		"email":       "email",
		"displayName": "fullname",
		"active":      "active",
	}
)

type userHandlers struct {
	config config.Scim
	store  *bun.DB
	logger *slog.Logger
}

// GetAll implements the SCIM v2 server interface for users.
func (us *userHandlers) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	result := scim.Page{
		TotalResults: 0,
		Resources:    []scim.Resource{},
	}

	records := make([]*model.User, 0)

	q := us.store.NewSelect().
		Model(&records).
		Order("username ASC")

	if params.FilterValidator != nil {
		validator := params.FilterValidator

		if err := validator.Validate(); err != nil {
			return result, err
		}

		q = us.filter(
			validator.GetFilter(),
			q,
		)
	}

	counter, err := q.Count(r.Context())

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, nil
		}

		us.logger.Error(
			"Failed to count all",
			slog.Any("error", err),
		)

		return result, err
	}

	result.TotalResults = counter

	if params.Count > 0 {
		q = q.Limit(
			params.Count,
		)

		if params.StartIndex < 1 {
			params.StartIndex = 1
		}

		if params.StartIndex > 1 {
			q = q.Offset(
				params.StartIndex * params.Count,
			)
		}

		if err := q.Scan(r.Context()); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return result, nil
			}

			us.logger.Error(
				"Failed to fetch all",
				slog.Any("error", err),
			)

			return result, err
		}

		for _, record := range records {
			result.Resources = append(
				result.Resources,
				scim.Resource{
					ID:         record.ID,
					ExternalID: optional.NewString(record.Scim),
					Meta: scim.Meta{
						Created:      &record.CreatedAt,
						LastModified: &record.UpdatedAt,
					},
					Attributes: scim.ResourceAttributes{
						"userName":    record.Username,
						"displayName": record.Fullname,
						"active":      record.Active,
					},
				},
			)
		}
	}

	return result, nil
}

// Get implements the SCIM v2 server interface for users.
func (us *userHandlers) Get(r *http.Request, id string) (scim.Resource, error) {
	record := &model.User{}

	if err := us.store.NewSelect().
		Model(record).
		Where("id = ?", id).
		Scan(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return scim.Resource{}, serrors.ScimErrorResourceNotFound(id)
		}

		us.logger.Error(
			"Failed to fetch user",
			slog.Any("error", err),
			slog.String("id", id),
		)

		return scim.Resource{}, err
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(record.Scim),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"userName":    record.Username,
			"displayName": record.Fullname,
			"active":      record.Active,
		},
	}

	return result, nil
}

// Create implements the SCIM v2 server interface for users.
func (us *userHandlers) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	externalID := ""
	if val, ok := attributes["externalId"]; ok {
		externalID = val.(string)
	}

	userName := ""
	if val, ok := attributes["userName"]; ok {
		userName = val.(string)
	}

	displayName := ""
	if val, ok := attributes["displayName"]; ok {
		displayName = val.(string)
	}

	active := false
	if val, ok := attributes["active"]; ok {
		active = val.(bool)
	}

	email := ""
	if val, ok := attributes["emails"]; ok {
		if is, ok := val.([]interface{}); ok {
			for _, i := range is {
				if vs, ok := i.(map[string]interface{}); ok {
					if p, ok := vs["primary"]; ok && p.(bool) {
						email = vs["value"].(string)
					}
				} else {
					us.logger.Error(
						"Failed to convert email",
						slog.String("method", "create"),
						slog.String("path", "emails"),
						slog.Any("value", i),
					)
				}
			}
		} else {
			us.logger.Error(
				"Failed to convert interface",
				slog.String("method", "create"),
				slog.String("path", "emails"),
				slog.Any("value", val),
			)
		}
	}

	record := &model.User{}

	if err := us.store.NewSelect().
		Model(record).
		Where("username = ? OR scim = ?", userName, externalID).
		Scan(r.Context()); err != nil && err != sql.ErrNoRows {
		us.logger.Error(
			"Failed to check if user exists",
			slog.Any("error", err),
			slog.String("user", userName),
		)

		return scim.Resource{}, err
	}

	record.Scim = externalID
	record.Username = userName
	record.Fullname = displayName
	record.Active = active
	record.Email = email

	if record.ID == "" {
		record.Password = secret.Generate(32)

		us.logger.Debug(
			"Creating new user",
			slog.String("user", record.Username),
		)

		if _, err := us.store.NewInsert().
			Model(record).
			Exec(r.Context()); err != nil {
			us.logger.Error(
				"Failed to create user",
				slog.Any("error", err),
				slog.String("user", record.Username),
			)

			return scim.Resource{}, err
		}
	} else {
		if _, err := us.store.NewUpdate().
			Model(record).
			Where("id = ?", record.ID).
			Exec(r.Context()); err != nil {
			us.logger.Error(
				"Failed to update user",
				slog.Any("error", err),
				slog.String("user", record.Username),
			)

			return scim.Resource{}, err
		}
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(record.Scim),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"userName":    record.Username,
			"displayName": record.Fullname,
			"active":      record.Active,
		},
	}

	return result, nil
}

// Replace implements the SCIM v2 server interface for users.
func (us *userHandlers) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	externalID := ""
	if val, ok := attributes["externalId"]; ok {
		externalID = val.(string)
	}

	userName := ""
	if val, ok := attributes["userName"]; ok {
		userName = val.(string)
	}

	displayName := ""
	if val, ok := attributes["displayName"]; ok {
		displayName = val.(string)
	}

	active := false
	if val, ok := attributes["active"]; ok {
		active = val.(bool)
	}

	email := ""
	if val, ok := attributes["emails"]; ok {
		if is, ok := val.([]interface{}); ok {
			for _, i := range is {
				if vs, ok := i.(map[string]interface{}); ok {
					if p, ok := vs["primary"]; ok && p.(bool) {
						email = vs["value"].(string)
					}
				} else {
					us.logger.Error(
						"Failed to convert email",
						slog.String("method", "create"),
						slog.String("path", "emails"),
						slog.Any("value", i),
					)
				}
			}
		} else {
			us.logger.Error(
				"Failed to convert interface",
				slog.String("method", "create"),
				slog.String("path", "emails"),
				slog.Any("value", val),
			)
		}
	}

	record := &model.User{}

	if err := us.store.NewSelect().
		Model(record).
		Where("id = ? OR scim = ?", id, externalID).
		Scan(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return scim.Resource{}, serrors.ScimErrorResourceNotFound(id)
		}

		us.logger.Error(
			"Failed to fetch user",
			slog.Any("error", err),
			slog.String("id", id),
		)

		return scim.Resource{}, err
	}

	record.Scim = externalID
	record.Username = userName
	record.Fullname = displayName
	record.Active = active
	record.Email = email

	if _, err := us.store.NewUpdate().
		Model(record).
		Where("id = ?", record.ID).
		Exec(r.Context()); err != nil {
		us.logger.Error(
			"Failed to update user",
			slog.Any("error", err),
			slog.String("id", id),
		)

		return scim.Resource{}, err
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(record.Scim),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"userName":    record.Username,
			"displayName": record.Fullname,
			"active":      record.Active,
		},
	}

	return result, nil
}

// Patch implements the SCIM v2 server interface for users.
func (us *userHandlers) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	record := &model.User{}

	if err := us.store.NewSelect().
		Model(record).
		Where("id = ?", id).
		Scan(r.Context()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return scim.Resource{}, serrors.ScimErrorResourceNotFound(id)
		}

		us.logger.Error(
			"Failed to fetch user",
			slog.Any("error", err),
			slog.String("id", id),
		)

		return scim.Resource{}, err
	}

	for _, operation := range operations {
		switch op := operation.Op; op {
		default:
			us.logger.Error(
				"Unknown operation",
				slog.String("method", "patch"),
				slog.String("id", id),
				slog.String("operation", op),
			)

			return scim.Resource{}, fmt.Errorf(
				"unknown operation: %s",
				op,
			)
		}
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(record.Scim),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"userName":    record.Username,
			"displayName": record.Fullname,
			"active":      record.Active,
		},
	}

	return result, nil
}

// Delete implements the SCIM v2 server interface for users.
func (us *userHandlers) Delete(r *http.Request, id string) error {
	if _, err := us.store.NewDelete().
		Model((*model.User)(nil)).
		Where("id = ?", id).
		Exec(r.Context()); err != nil {
		us.logger.Error(
			"Failed to delete user",
			slog.Any("error", err),
			slog.String("id", id),
		)

		return err
	}

	return nil
}

func (us *userHandlers) filter(expr filter.Expression, db *bun.SelectQuery) *bun.SelectQuery {
	switch e := expr.(type) {
	case *filter.AttributeExpression:
		return us.handleAttributeExpression(e, db)
	default:
		us.logger.Error(
			"Unsupported expression type for user filter",
			slog.String("type", fmt.Sprintf("%T", e)),
		)
	}

	return db
}

func (us *userHandlers) handleAttributeExpression(e *filter.AttributeExpression, db *bun.SelectQuery) *bun.SelectQuery {
	scimAttr := e.AttributePath.String()
	column, ok := userAttributeMapping[scimAttr]

	if !ok {
		us.logger.Error(
			"Attribute is not mapped for users",
			slog.String("attribute", scimAttr),
		)

		return db
	}

	value := e.CompareValue

	switch operator := strings.ToLower(string(e.Operator)); operator {
	case "eq":
		return db.Where("? = ?", bun.Ident(column), value)
	case "ne":
		return db.Where("? <> ?", bun.Ident(column), value)
	case "co":
		return db.Where("? LIKE ?", bun.Ident(column), "%"+fmt.Sprintf("%v", value)+"%")
	case "sw":
		return db.Where("? LIKE ?", bun.Ident(column), fmt.Sprintf("%v", value)+"%")
	case "ew":
		return db.Where("? LIKE ?", bun.Ident(column), "%"+fmt.Sprintf("%v", value))
	case "gt":
		return db.Where("? > ?", bun.Ident(column), value)
	case "ge":
		return db.Where("? >= ?", bun.Ident(column), value)
	case "lt":
		return db.Where("? < ?", bun.Ident(column), value)
	case "le":
		return db.Where("? <= ?", bun.Ident(column), value)
	default:
		us.logger.Error(
			"Unsupported attribute operator for user filter",
			slog.String("operator", operator),
		)
	}

	return db
}
