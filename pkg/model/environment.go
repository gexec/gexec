package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Environment)(nil)
)

// Environment defines the model for environments table.
type Environment struct {
	bun.BaseModel `bun:"table:environments"`

	ID        string               `bun:",pk,type:varchar(20)"`
	ProjectID string               `bun:"type:varchar(20)"`
	Project   *Project             `bun:"rel:belongs-to,join:project_id=id"`
	Slug      string               `bun:"type:varchar(255)"`
	Name      string               `bun:"type:varchar(255)"`
	Secrets   []*EnvironmentSecret `bun:"rel:has-many,join:id=environment_id"`
	Values    []*EnvironmentValue  `bun:"rel:has-many,join:id=environment_id"`
	CreatedAt time.Time            `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time            `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Environment) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *Environment) SerializeSecret(passphrase string) error {
	for _, row := range m.Secrets {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Values {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Environment) DeserializeSecret(passphrase string) error {
	for _, row := range m.Secrets {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Values {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
