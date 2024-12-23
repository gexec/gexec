package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Team)(nil)
)

// Team defines the model for teams table.
type Team struct {
	bun.BaseModel `bun:"table:teams"`

	ID        string         `bun:",pk,type:varchar(20)"`
	Slug      string         `bun:",unique,type:varchar(255)"`
	Name      string         `bun:"type:varchar(255)"`
	CreatedAt time.Time      `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time      `bun:",nullzero,notnull,default:current_timestamp"`
	Auths     []*TeamAuth    `bun:"rel:has-many,join:id=team_id"`
	Users     []*UserTeam    `bun:"rel:has-many,join:id=team_id"`
	Projects  []*TeamProject `bun:"rel:has-many,join:id=team_id"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Team) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
