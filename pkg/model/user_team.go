package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*UserTeam)(nil)
)

const (
	// UserTeamOwnerPerm defines the permission for an owner on user teams.
	UserTeamOwnerPerm = OwnerPerm

	// UserTeamAdminPerm defines the permission for an admin on user teams.
	UserTeamAdminPerm = AdminPerm

	// UserTeamUserPerm defines the permission for an user on user teams.
	UserTeamUserPerm = UserPerm
)

// UserTeam defines the model for user_teams table.
type UserTeam struct {
	bun.BaseModel `bun:"table:user_teams"`

	UserID    string    `bun:",pk,type:varchar(20)"`
	User      *User     `bun:"rel:belongs-to,join:user_id=id"`
	TeamID    string    `bun:",pk,type:varchar(20)"`
	Team      *Team     `bun:"rel:belongs-to,join:team_id=id"`
	Perm      string    `bun:"type:varchar(32)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *UserTeam) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}

	return nil
}
