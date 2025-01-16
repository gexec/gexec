package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Inventory)(nil)
)

// Inventory defines the model for inventories table.
type Inventory struct {
	bun.BaseModel `bun:"table:inventories"`

	ID           string      `bun:",pk,type:varchar(20)"`
	ProjectID    string      `bun:"type:varchar(20)"`
	Project      *Project    `bun:"rel:belongs-to,join:project_id=id"`
	RepositoryID string      `bun:",nullzero,type:varchar(20)"`
	Repository   *Repository `bun:"rel:belongs-to,join:repository_id=id"`
	CredentialID string      `bun:",nullzero,type:varchar(20)"`
	Credential   *Credential `bun:"rel:belongs-to,join:credential_id=id"`
	BecomeID     string      `bun:",nullzero,type:varchar(20)"`
	Become       *Credential `bun:"rel:belongs-to,join:become_id=id"`
	Slug         string      `bun:"type:varchar(255)"`
	Name         string      `bun:"type:varchar(255)"`
	Kind         string      `bun:"type:varchar(255)"`
	Content      string      `bun:"type:text"`
	CreatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt    time.Time   `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Inventory) BeforeAppendModel(_ context.Context, query bun.Query) error {
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

func (m *Inventory) SerializeSecret(passphrase string) error {
	if m.Repository != nil {
		if err := m.Repository.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Credential != nil {
		if err := m.Credential.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Become != nil {
		if err := m.Become.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

func (m *Inventory) DeserializeSecret(passphrase string) error {
	if m.Repository != nil {
		if err := m.Repository.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Credential != nil {
		if err := m.Credential.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Become != nil {
		if err := m.Become.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
