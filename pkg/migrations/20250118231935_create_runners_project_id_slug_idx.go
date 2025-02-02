package migrations

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Runner struct {
			bun.BaseModel `bun:"table:runners"`

			ID        string `bun:",pk,type:varchar(20)"`
			ProjectID string `bun:"type:varchar(20)"`
			Slug      string `bun:"type:varchar(255)"`
		}

		_, err := db.NewCreateIndex().
			Model((*Runner)(nil)).
			Index("runners_project_id_and_slug_idx").
			Column("project_id").
			Column("slug").
			Unique().
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Runner struct {
			bun.BaseModel `bun:"table:runners"`
		}

		_, err := db.NewDropIndex().
			Model((*Runner)(nil)).
			IfExists().
			Index("runners_project_id_and_slug_idx").
			Exec(ctx)

		return err
	})
}
