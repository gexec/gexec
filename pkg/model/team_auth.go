package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TeamAuth)(nil)
)

// TeamAuth defines the model for team_auths table.
type TeamAuth struct {
	bun.BaseModel `bun:"table:team_auths"`

	ID        string    `bun:",pk,type:varchar(20)"`
	TeamID    string    `bun:"type:varchar(255)"`
	Team      *Team     `bun:"rel:belongs-to,join:team_id=id"`
	Provider  string    `bun:"type:varchar(255)"`
	Ref       string    `bun:"type:varchar(255)"`
	Name      string    `bun:"type:varchar(255)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *TeamAuth) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
