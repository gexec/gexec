package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Schedule)(nil)
)

// Schedule defines the model for schedules table.
type Schedule struct {
	bun.BaseModel `bun:"table:schedules"`

	ID         string    `bun:",pk,type:varchar(20)"`
	ProjectID  string    `bun:"type:varchar(20)"`
	Project    *Project  `bun:"rel:belongs-to,join:project_id=id"`
	TemplateID string    `bun:"type:varchar(20)"`
	Template   *Template `bun:"rel:belongs-to,join:template_id=id"`
	Slug       string    `bun:"type:varchar(255)"`
	Name       string    `bun:"type:varchar(255)"`
	Cron       string    `bun:"type:varchar(255)"`
	Active     bool      `bun:"type:bool"`
	CreatedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Schedule) BeforeAppendModel(_ context.Context, query bun.Query) error {
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

func (m *Schedule) SerializeSecret(passphrase string) error {
	if m.Template != nil {
		if err := m.Template.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

func (m *Schedule) DeserializeSecret(passphrase string) error {
	if m.Template != nil {
		if err := m.Template.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
