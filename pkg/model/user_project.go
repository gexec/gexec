package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*UserProject)(nil)
)

const (
	// UserProjectOwnerPerm defines the permission for an owner on user projects.
	UserProjectOwnerPerm = OwnerPerm

	// UserProjectAdminPerm defines the permission for an admin on user projects.
	UserProjectAdminPerm = AdminPerm

	// UserProjectUserPerm defines the permission for an user on user projects.
	UserProjectUserPerm = UserPerm
)

// UserProject defines the model for user_projects table.
type UserProject struct {
	bun.BaseModel `bun:"table:user_projects"`

	UserID    string    `bun:",pk,type:varchar(20)"`
	User      *User     `bun:"rel:belongs-to,join:user_id=id"`
	ProjectID string    `bun:",pk,type:varchar(20)"`
	Project   *Project  `bun:"rel:belongs-to,join:project_id=id"`
	Perm      string    `bun:"type:varchar(32)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *UserProject) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}

	return nil
}
