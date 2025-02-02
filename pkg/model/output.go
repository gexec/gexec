package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Output)(nil)
)

// Output defines the model for outputs table.
type Output struct {
	bun.BaseModel `bun:"table:outputs"`

	ID          string     `bun:",pk,type:varchar(20)"`
	ExecutionID string     `bun:"type:varchar(20)"`
	Execution   *Execution `bun:"rel:belongs-to,join:execution_id=id"`
	Content     string     `bun:"type:text"`
	CreatedAt   time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt   time.Time  `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Output) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
