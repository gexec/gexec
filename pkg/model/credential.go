package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Credential)(nil)
)

// Credential defines the model for credentials table.
type Credential struct {
	bun.BaseModel `bun:"table:credentials"`

	ID        string          `bun:",pk,type:varchar(20)"`
	ProjectID string          `bun:"type:varchar(20)"`
	Project   *Project        `bun:"rel:belongs-to,join:project_id=id"`
	Slug      string          `bun:"type:varchar(255)"`
	Name      string          `bun:"type:varchar(255)"`
	Kind      string          `bun:"type:varchar(255)"`
	Override  bool            `bun:"type:bool"`
	Shell     CredentialShell `bun:"embed:shell_"`
	Login     CredentialLogin `bun:"embed:login_"`
	CreatedAt time.Time       `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time       `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Credential) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *Credential) SerializeSecret(passphrase string) error {
	switch m.Kind {
	case "shell":
		return m.Shell.SerializeSecret(passphrase)
	case "login":
		return m.Login.SerializeSecret(passphrase)
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Credential) DeserializeSecret(passphrase string) error {
	switch m.Kind {
	case "shell":
		return m.Shell.DeserializeSecret(passphrase)
	case "login":
		return m.Login.DeserializeSecret(passphrase)
	}

	return nil
}
