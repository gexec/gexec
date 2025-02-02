package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*GroupProject)(nil)
)

const (
	// GroupProjectOwnerPerm defines the permission for an owner on group projects.
	GroupProjectOwnerPerm = OwnerPerm

	// GroupProjectAdminPerm defines the permission for an admin on group projects.
	GroupProjectAdminPerm = AdminPerm

	// GroupProjectUserPerm defines the permission for an user on group projects.
	GroupProjectUserPerm = UserPerm
)

// GroupProject defines the model for group_projects table.
type GroupProject struct {
	bun.BaseModel `bun:"table:group_projects"`

	GroupID   string    `bun:",pk,type:varchar(20)"`
	Group     *Group    `bun:"rel:belongs-to,join:group_id=id"`
	ProjectID string    `bun:",pk,type:varchar(20)"`
	Project   *Project  `bun:"rel:belongs-to,join:project_id=id"`
	Perm      string    `bun:"type:varchar(32)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *GroupProject) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}

	return nil
}
