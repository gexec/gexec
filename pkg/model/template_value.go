package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TemplateValue)(nil)
)

// TemplateValue defines the model for template:values table.
type TemplateValue struct {
	bun.BaseModel `bun:"table:template_values"`

	ID        string    `bun:",pk,type:varchar(20)"`
	SurveyID  string    `bun:"type:varchar(20)"`
	Name      string    `bun:"type:varchar(255)"`
	Value     string    `bun:"type:text"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *TemplateValue) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *TemplateValue) SerializeSecret(_ string) error {
	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *TemplateValue) DeserializeSecret(_ string) error {
	return nil
}
