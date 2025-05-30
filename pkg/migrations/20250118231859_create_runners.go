package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Runner struct {
			bun.BaseModel `bun:"table:runners"`

			ID        string    `bun:",pk,type:varchar(20)"`
			ProjectID string    `bun:",nullzero,type:varchar(20)"`
			Slug      string    `bun:"type:varchar(255)"`
			Name      string    `bun:"type:varchar(255)"`
			Token     string    `bun:"type:varchar(255)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Runner)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Runner struct {
			bun.BaseModel `bun:"table:runners"`
		}

		_, err := db.NewDropTable().
			Model((*Runner)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
