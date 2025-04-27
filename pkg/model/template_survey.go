package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TemplateSurvey)(nil)
)

// TemplateSurvey defines the model for template_surveys table.
type TemplateSurvey struct {
	bun.BaseModel `bun:"table:template_surveys"`

	ID          string           `bun:",pk,type:varchar(20)"`
	TemplateID  string           `bun:"type:varchar(20)"`
	Name        string           `bun:"type:varchar(255)"`
	Title       string           `bun:"type:varchar(255)"`
	Description string           `bun:"type:text"`
	Kind        string           `bun:"type:varchar(255)"`
	Required    bool             `bun:"type:bool"`
	Values      []*TemplateValue `bun:"rel:has-many,join:id=survey_id"`
	CreatedAt   time.Time        `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time        `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *TemplateSurvey) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *TemplateSurvey) SerializeSecret(passphrase string) error {
	for _, row := range m.Values {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *TemplateSurvey) DeserializeSecret(passphrase string) error {
	for _, row := range m.Values {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
