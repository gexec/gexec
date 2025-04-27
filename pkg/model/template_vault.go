package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TemplateVault)(nil)
)

// TemplateVault defines the model for template_vaults table.
type TemplateVault struct {
	bun.BaseModel `bun:"table:template_vaults"`

	ID           string      `bun:",pk,type:varchar(20)"`
	TemplateID   string      `bun:"type:varchar(20)"`
	CredentialID string      `bun:",nullzero,type:varchar(20)"`
	Credential   *Credential `bun:"rel:belongs-to,join:credential_id=id"`
	Name         string      `bun:"type:varchar(255)"`
	Kind         string      `bun:"type:varchar(255)"`
	Script       string      `bun:"type:text"`
	CreatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *TemplateVault) BeforeAppendModel(_ context.Context, query bun.Query) error {
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

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *TemplateVault) SerializeSecret(_ string) error {
	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *TemplateVault) DeserializeSecret(_ string) error {
	return nil
}
