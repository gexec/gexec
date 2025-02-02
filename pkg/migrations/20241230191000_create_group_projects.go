package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type GroupProject struct {
			bun.BaseModel `bun:"table:group_projects"`

			GroupID   string    `bun:",pk,type:varchar(20)"`
			ProjectID string    `bun:",pk,type:varchar(20)"`
			Perm      string    `bun:"type:varchar(32)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*GroupProject)(nil)).
			WithForeignKeys().
			ForeignKey(`(group_id) REFERENCES groups (id) ON DELETE CASCADE`).
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type GroupProject struct {
			bun.BaseModel `bun:"table:group_projects"`
		}

		_, err := db.NewDropTable().
			Model((*GroupProject)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
