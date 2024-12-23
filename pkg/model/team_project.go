package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*TeamProject)(nil)
)

const (
	// TeamProjectOwnerPerm defines the permission for an owner on team projects.
	TeamProjectOwnerPerm = OwnerPerm

	// TeamProjectAdminPerm defines the permission for an admin on team projects.
	TeamProjectAdminPerm = AdminPerm

	// TeamProjectUserPerm defines the permission for an user on team projects.
	TeamProjectUserPerm = UserPerm
)

// TeamProject defines the model for team_projects table.
type TeamProject struct {
	bun.BaseModel `bun:"table:team_projects"`

	TeamID    string    `bun:",pk,type:varchar(20)"`
	Team      *Team     `bun:"rel:belongs-to,join:team_id=id"`
	ProjectID string    `bun:",pk,type:varchar(20)"`
	Project   *Project  `bun:"rel:belongs-to,join:project_id=id"`
	Perm      string    `bun:"type:varchar(32)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *TeamProject) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}

	return nil
}
