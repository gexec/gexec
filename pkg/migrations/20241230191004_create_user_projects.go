package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type UserProject struct {
			bun.BaseModel `bun:"table:user_projects"`

			UserID    string    `bun:",pk,type:varchar(20)"`
			ProjectID string    `bun:",pk,type:varchar(20)"`
			Perm      string    `bun:"type:varchar(32)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*UserProject)(nil)).
			WithForeignKeys().
			ForeignKey(`(user_id) REFERENCES users (id) ON DELETE CASCADE`).
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type UserProject struct {
			bun.BaseModel `bun:"table:user_projects"`
		}

		_, err := db.NewDropTable().
			Model((*UserProject)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
