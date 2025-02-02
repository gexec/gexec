package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Execution)(nil)
	_ bun.AfterScanRowHook      = (*Event)(nil)
)

// Execution defines the model for executions table.
type Execution struct {
	bun.BaseModel `bun:"table:executions"`

	ID          string    `bun:",pk,type:varchar(20)"`
	ProjectID   string    `bun:"type:varchar(20)"`
	Project     *Project  `bun:"rel:belongs-to,join:project_id=id"`
	TemplateID  string    `bun:"type:varchar(20)"`
	Template    *Template `bun:"rel:belongs-to,join:template_id=id"`
	Name        string    `bun:"-"`
	Status      string    `bun:"type:varchar(255)"`
	Playbook    string    `bun:"type:varchar(255)"`
	Environment string    `bun:"type:varchar(255)"`
	Secret      string    `bun:"type:varchar(255)"`
	Limit       string    `bun:"type:varchar(255)"`
	Branch      string    `bun:"type:varchar(255)"`
	Debug       bool      `bun:"type:bool"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Execution) BeforeAppendModel(_ context.Context, query bun.Query) error {
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

	m.Name = fmt.Sprintf(
		"#%s",
		m.ID,
	)

	return nil
}

// AfterScanRow implements the bun hook interface.
func (m *Execution) AfterScanRow(_ context.Context) error {
	m.Name = fmt.Sprintf(
		"#%s",
		m.ID,
	)

	return nil
}

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *Execution) SerializeSecret(passphrase string) error {
	if m.Template != nil {
		if err := m.Template.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Execution) DeserializeSecret(passphrase string) error {
	if m.Template != nil {
		if err := m.Template.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
