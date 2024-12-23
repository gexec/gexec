package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type UserTeam struct {
			bun.BaseModel `bun:"table:user_teams"`

			UserID    string    `bun:",pk,type:varchar(20)"`
			TeamID    string    `bun:",pk,type:varchar(20)"`
			Perm      string    `bun:"type:varchar(32)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*UserTeam)(nil)).
			WithForeignKeys().
			ForeignKey(`(user_id) REFERENCES users (id) ON DELETE CASCADE`).
			ForeignKey(`(team_id) REFERENCES teams (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type UserTeam struct {
			bun.BaseModel `bun:"table:user_teams"`
		}

		_, err := db.NewDropTable().
			Model((*UserTeam)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
