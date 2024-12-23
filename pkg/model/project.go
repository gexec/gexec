package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Project)(nil)
)

// Project defines the model for projects table.
type Project struct {
	bun.BaseModel `bun:"table:projects"`

	ID        string         `bun:",pk,type:varchar(20)"`
	Slug      string         `bun:",unique,type:varchar(255)"`
	Name      string         `bun:"type:varchar(255)"`
	CreatedAt time.Time      `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time      `bun:",nullzero,notnull,default:current_timestamp"`
	Teams     []*TeamProject `bun:"rel:has-many,join:id=project_id"`
	Users     []*UserProject `bun:"rel:has-many,join:id=project_id"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Project) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}

		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}

		m.UpdatedAt = time.Now()
	}

	return nil
}
