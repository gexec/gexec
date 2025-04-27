package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Template)(nil)
)

// Template defines the model for templates table.
type Template struct {
	bun.BaseModel `bun:"table:templates"`

	ID            string            `bun:",pk,type:varchar(20)"`
	ProjectID     string            `bun:"type:varchar(20)"`
	Project       *Project          `bun:"rel:belongs-to,join:project_id=id"`
	RepositoryID  string            `bun:",nullzero,type:varchar(20)"`
	Repository    *Repository       `bun:"rel:belongs-to,join:repository_id=id"`
	InventoryID   string            `bun:",nullzero,type:varchar(20)"`
	Inventory     *Inventory        `bun:"rel:belongs-to,join:inventory_id=id"`
	EnvironmentID string            `bun:",nullzero,type:varchar(20)"`
	Environment   *Environment      `bun:"rel:belongs-to,join:environment_id=id"`
	Slug          string            `bun:"type:varchar(255)"`
	Name          string            `bun:"type:varchar(255)"`
	Description   string            `bun:"type:text"`
	Path          string            `bun:"type:varchar(255)"`
	Arguments     string            `bun:"type:varchar(255)"`
	Limit         string            `bun:"type:varchar(255)"`
	Executor      string            `bun:"type:varchar(255)"`
	Branch        string            `bun:"type:varchar(255)"`
	Override      bool              `bun:"type:bool"`
	CreatedAt     time.Time         `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt     time.Time         `bun:",nullzero,notnull,default:current_timestamp"`
	Surveys       []*TemplateSurvey `bun:"rel:has-many,join:id=template_id"`
	Vaults        []*TemplateVault  `bun:"rel:has-many,join:id=template_id"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Template) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *Template) SerializeSecret(passphrase string) error {
	if m.Repository != nil {
		if err := m.Repository.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Inventory != nil {
		if err := m.Inventory.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Environment != nil {
		if err := m.Environment.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Surveys {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Vaults {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Template) DeserializeSecret(passphrase string) error {
	if m.Repository != nil {
		if err := m.Repository.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Inventory != nil {
		if err := m.Inventory.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	if m.Environment != nil {
		if err := m.Environment.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Surveys {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Vaults {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
