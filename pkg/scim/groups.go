package scim

import (
	"fmt"
	"net/http"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/genexec/genexec/pkg/config"
	"github.com/genexec/genexec/pkg/model"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

var (
	groupAttributeMapping = map[string]string{
		"displayName": "name",
	}
)

type groupHandlers struct {
	config config.Scim
	store  *bun.DB
	logger zerolog.Logger
}

// GetAll implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	result := scim.Page{
		TotalResults: 0,
		Resources:    []scim.Resource{},
	}

	// q := gs.store.WithContext(
	// 	r.Context(),
	// ).Model(
	// 	&model.Team{},
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 	}),
	// ).Order(
	// 	"name ASC",
	// )

	// if params.FilterValidator != nil {
	// 	validator := params.FilterValidator

	// 	if err := validator.Validate(); err != nil {
	// 		return result, err
	// 	}

	// 	q = gs.filter(
	// 		validator.GetFilter(),
	// 		q,
	// 	)
	// }

	// counter := int64(0)

	// if err := q.Count(
	// 	&counter,
	// ).Error; err != nil {
	// 	return result, err
	// }

	// result.TotalResults = int(counter)

	// if params.Count > 0 {
	// 	q = q.Limit(
	// 		params.Count,
	// 	)

	// 	if params.StartIndex < 1 {
	// 		params.StartIndex = 1
	// 	}

	// 	if params.StartIndex > 1 {
	// 		q = q.Offset(
	// 			params.StartIndex * params.Count,
	// 		)
	// 	}

	// 	records := make([]*model.Team, 0)

	// 	if err := q.Find(
	// 		&records,
	// 	).Error; err != nil {
	// 		return result, err
	// 	}

	// 	for _, record := range records {
	// 		auth := &model.TeamAuth{}

	// 		for _, row := range record.Auths {
	// 			if row.Provider == "scim" {
	// 				auth = row
	// 			}
	// 		}

	// 		result.Resources = append(
	// 			result.Resources,
	// 			scim.Resource{
	// 				ID:         record.ID,
	// 				ExternalID: optional.NewString(auth.Ref),
	// 				Meta: scim.Meta{
	// 					Created:      &record.CreatedAt,
	// 					LastModified: &record.UpdatedAt,
	// 				},
	// 				Attributes: scim.ResourceAttributes{
	// 					"displayName": auth.Name,
	// 				},
	// 			},
	// 		)
	// 	}
	// }

	return result, nil
}

// Get implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) Get(r *http.Request, id string) (scim.Resource, error) {
	record := &model.Team{}

	// if err := gs.store.WithContext(
	// 	r.Context(),
	// ).Model(
	// 	&model.Team{},
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 	}),
	// ).Where(&model.Team{
	// 	ID: id,
	// }).First(
	// 	record,
	// ).Error; err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	// 	}

	// 	return scim.Resource{}, err
	// }

	auth := &model.TeamAuth{}

	for _, row := range record.Auths {
		if row.Provider == "scim" {
			auth = row
		}
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(auth.Ref),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"displayName": auth.Name,
		},
	}

	return result, nil
}

// Create implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// tx := gs.store.WithContext(
	// 	r.Context(),
	// ).Begin()
	// defer tx.Rollback()

	externalId := ""
	if val, ok := attributes["externalId"]; ok {
		externalId = val.(string)
	}

	displayName := ""
	if val, ok := attributes["displayName"]; ok {
		displayName = val.(string)
	}

	auth := &model.TeamAuth{
		Provider: "scim",
		Ref:      externalId,
		Name:     displayName,
	}

	record := &model.Team{}

	// if err := tx.Model(
	// 	&model.Team{},
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 		Name:     displayName,
	// 	}),
	// ).First(record).Error; err != nil && err != gorm.ErrRecordNotFound {
	// 	return scim.Resource{}, err
	// }

	record.Name = displayName

	if record.ID == "" {
		record.Auths = []*model.TeamAuth{auth}

		// if err := tx.Create(record).Error; err != nil {
		// 	return scim.Resource{}, err
		// }
	} else {
		for _, row := range record.Auths {
			if row.Provider == "scim" && row.Ref == externalId {
				auth.Name = displayName

				// if err := tx.Save(record).Error; err != nil {
				// 	return scim.Resource{}, err
				// }
			}
		}
	}

	// if err := tx.Commit().Error; err != nil {
	// 	return scim.Resource{}, err
	// }

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(auth.Ref),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"displayName": auth.Name,
		},
	}

	return result, nil
}

// Replace implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// tx := gs.store.WithContext(
	// 	r.Context(),
	// ).Begin()
	// defer tx.Rollback()

	externalId := ""
	if val, ok := attributes["externalId"]; ok {
		externalId = val.(string)
	}

	displayName := ""
	if val, ok := attributes["displayName"]; ok {
		displayName = val.(string)
	}

	auth := &model.TeamAuth{
		Provider: "scim",
		Ref:      externalId,
		Name:     displayName,
	}

	record := &model.Team{}

	// if err := gs.store.WithContext(
	// 	r.Context(),
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 	}),
	// ).Where(&model.Team{
	// 	ID: id,
	// }).First(
	// 	record,
	// ).Error; err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	// 	}

	// 	return scim.Resource{}, err
	// }

	record.Name = displayName

	for _, row := range record.Auths {
		if auth.Provider == "scim" && auth.Ref == externalId {
			auth = row
			auth.Name = displayName
		}
	}

	if auth.ID == "" {
		auth.TeamID = record.ID

		// if err := tx.Create(auth).Error; err != nil {
		// 	return scim.Resource{}, err
		// }
	} else {
		// if err := tx.Save(auth).Error; err != nil {
		// 	return scim.Resource{}, err
		// }
	}

	// if err := tx.Save(record).Error; err != nil {
	// 	return scim.Resource{}, err
	// }

	// if err := tx.Commit().Error; err != nil {
	// 	return scim.Resource{}, err
	// }

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(auth.Ref),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"displayName": auth.Name,
		},
	}

	return result, nil
}

// Patch implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	record := &model.Team{}

	// if err := gs.store.WithContext(
	// 	r.Context(),
	// ).Model(
	// 	&model.Team{},
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 	}),
	// ).Where(&model.Team{
	// 	ID: id,
	// }).First(
	// 	record,
	// ).Error; err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	// 	}

	// 	return scim.Resource{}, err
	// }

	// tx := gs.store.WithContext(
	// 	r.Context(),
	// ).Begin()
	// defer tx.Rollback()

	for _, operation := range operations {
		switch op := operation.Op; op {
		case "remove":
			switch {
			case operation.Path.String() == "members":
				if is, ok := operation.Value.([]interface{}); ok {
					for _, i := range is {
						if vs, ok := i.(map[string]interface{}); ok {
							if _, ok := vs["value"]; ok {
								// if err := tx.Where(
								// 	model.UserTeam{
								// 		TeamID: record.ID,
								// 		UserID: v.(string),
								// 	},
								// ).Delete(&model.UserTeam{}).Error; err != nil {
								// 	return scim.Resource{}, err
								// }
							} else {
								gs.logger.Error().
									Str("method", "patch").
									Str("id", id).
									Str("operation", op).
									Str("path", "members").
									Msgf("Failed to convert member: %v", vs)
							}
						} else {
							gs.logger.Error().
								Str("method", "patch").
								Str("id", id).
								Str("operation", op).
								Str("path", "members").
								Msgf("Failed to convert values: %v", i)
						}
					}
				} else {
					gs.logger.Error().
						Str("method", "patch").
						Str("id", id).
						Str("operation", op).
						Str("path", "members").
						Msgf("Failed to convert interface: %v", operation.Value)
				}
			default:
				gs.logger.Error().
					Str("method", "patch").
					Str("id", id).
					Str("operation", op).
					Str("path", operation.Path.String()).
					Msg("Unknown path")

				return scim.Resource{}, fmt.Errorf(
					"unknown path: %s",
					operation.Path.String(),
				)
			}
		case "add":
			switch {
			case operation.Path.String() == "members":
				if is, ok := operation.Value.([]interface{}); ok {
					for _, i := range is {
						if vs, ok := i.(map[string]interface{}); ok {
							if _, ok := vs["value"]; ok {
								// if err := tx.Where(
								// 	model.UserTeam{
								// 		TeamID: record.ID,
								// 		UserID: v.(string),
								// 	},
								// ).Attrs(
								// 	model.UserTeam{
								// 		Perm: "owner",
								// 	},
								// ).FirstOrCreate(&model.UserTeam{}).Error; err != nil {
								// 	return scim.Resource{}, err
								// }
							} else {
								gs.logger.Error().
									Str("method", "patch").
									Str("id", id).
									Str("operation", op).
									Str("path", "members").
									Msgf("Failed to convert member: %v", vs)
							}
						} else {
							gs.logger.Error().
								Str("method", "patch").
								Str("id", id).
								Str("operation", op).
								Str("path", "members").
								Msgf("Failed to convert values: %v", i)
						}
					}
				} else {
					gs.logger.Error().
						Str("method", "patch").
						Str("id", id).
						Str("operation", op).
						Str("path", "members").
						Msgf("Failed to convert interface: %v", operation.Value)
				}
			default:
				gs.logger.Error().
					Str("method", "patch").
					Str("id", id).
					Str("operation", op).
					Str("path", operation.Path.String()).
					Msg("Unknown path")

				return scim.Resource{}, fmt.Errorf(
					"unknown path: %s",
					operation.Path.String(),
				)
			}
		default:
			gs.logger.Error().
				Str("method", "patch").
				Str("id", id).
				Str("operation", op).
				Msg("Unknown operation")

			return scim.Resource{}, fmt.Errorf(
				"unknown operation: %s",
				op,
			)
		}
	}

	// if err := tx.Commit().Error; err != nil {
	// 	return scim.Resource{}, err
	// }

	auth := &model.TeamAuth{}

	for _, row := range record.Auths {
		if row.Provider == "scim" {
			auth = row
		}
	}

	result := scim.Resource{
		ID:         record.ID,
		ExternalID: optional.NewString(auth.Ref),
		Meta: scim.Meta{
			Created:      &record.CreatedAt,
			LastModified: &record.UpdatedAt,
		},
		Attributes: scim.ResourceAttributes{
			"displayName": auth.Name,
		},
	}

	return result, nil
}

// Delete implements the SCIM v2 server interface for groups.
func (gs *groupHandlers) Delete(r *http.Request, id string) error {
	// tx := gs.store.WithContext(
	// 	r.Context(),
	// ).Begin()
	// defer tx.Rollback()

	// if err := tx.Model(
	// 	&model.Team{},
	// ).InnerJoins(
	// 	"Auths",
	// 	gs.store.Where(&model.TeamAuth{
	// 		Provider: "scim",
	// 	}),
	// ).Where(&model.Team{
	// 	ID: id,
	// }).Delete(
	// 	&model.Team{},
	// ).Error; err != nil {
	// 	return err
	// }

	// return tx.Commit().Error

	return nil
}

// func (gs *groupHandlers) filter(expr filter.Expression, db *gorm.DB) *gorm.DB {
// 	switch e := expr.(type) {
// 	case *filter.AttributeExpression:
// 		return gs.handleAttributeExpression(e, db)
// 	default:
// 		gs.logger.Error().
// 			Str("type", fmt.Sprintf("%T", e)).
// 			Msg("Unsupported expression type for group filter")
// 	}

// 	return db
// }

// func (gs *groupHandlers) handleAttributeExpression(e *filter.AttributeExpression, db *gorm.DB) *gorm.DB {
// 	scimAttr := e.AttributePath.String()
// 	column, ok := groupAttributeMapping[scimAttr]

// 	if !ok {
// 		gs.logger.Error().
// 			Str("attribute", scimAttr).
// 			Msg("Attribute is not mapped for groups")

// 		return db
// 	}

// 	value := e.CompareValue

// 	switch operator := strings.ToLower(string(e.Operator)); operator {
// 	case "eq":
// 		return db.Where(fmt.Sprintf("%s = ?", column), value)
// 	case "ne":
// 		return db.Where(fmt.Sprintf("%s <> ?", column), value)
// 	case "co":
// 		return db.Where(fmt.Sprintf("%s LIKE ?", column), "%"+fmt.Sprintf("%v", value)+"%")
// 	case "sw":
// 		return db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%v", value)+"%")
// 	case "ew":
// 		return db.Where(fmt.Sprintf("%s LIKE ?", column), "%"+fmt.Sprintf("%v", value))
// 	case "gt":
// 		return db.Where(fmt.Sprintf("%s > ?", column), value)
// 	case "ge":
// 		return db.Where(fmt.Sprintf("%s >= ?", column), value)
// 	case "lt":
// 		return db.Where(fmt.Sprintf("%s < ?", column), value)
// 	case "le":
// 		return db.Where(fmt.Sprintf("%s <= ?", column), value)
// 	default:
// 		gs.logger.Error().
// 			Str("operator", operator).
// 			Msgf("Unsupported attribute operator for group filter")
// 	}

// 	return db
// }
