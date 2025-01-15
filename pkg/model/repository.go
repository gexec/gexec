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

// Repository defines the model for repositories table.
type Repository struct {
	bun.BaseModel `bun:"table:repositories"`

	ID           string      `bun:",pk,type:varchar(20)"`
	ProjectID    string      `bun:"type:varchar(20)"`
	Project      *Project    `bun:"rel:belongs-to,join:project_id=id"`
	CredentialID string      `bun:"type:varchar(20)"`
	Credential   *Credential `bun:"rel:belongs-to,join:credential_id=id"`
	Slug         string      `bun:"type:varchar(255)"`
	Name         string      `bun:"type:varchar(255)"`
	URL          string      `bun:"type:text"`
	Branch       string      `bun:"type:varchar(255)"`
	CreatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Repository) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
