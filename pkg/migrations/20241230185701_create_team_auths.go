package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TeamAuth struct {
			bun.BaseModel `bun:"table:team_auths"`

			ID        string    `bun:",pk,type:varchar(20)"`
			TeamID    string    `bun:"type:varchar(255)"`
			Provider  string    `bun:"type:varchar(255)"`
			Ref       string    `bun:"type:varchar(255)"`
			Name      string    `bun:"type:varchar(255)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*TeamAuth)(nil)).
			WithForeignKeys().
			ForeignKey(`(team_id) REFERENCES teams (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type TeamAuth struct {
			bun.BaseModel `bun:"table:team_auths"`
		}

		_, err := db.NewDropTable().
			Model((*TeamAuth)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
