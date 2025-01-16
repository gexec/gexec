package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Environment struct {
			bun.BaseModel `bun:"table:environments"`

			ID        string    `bun:",pk,type:varchar(20)"`
			ProjectID string    `bun:"type:varchar(20)"`
			Slug      string    `bun:"type:varchar(255)"`
			Name      string    `bun:"type:varchar(255)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		if _, err := db.NewCreateTable().
			Model((*Environment)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewCreateIndex().
			Model((*Environment)(nil)).
			Index("environments_project_id_and_slug_idx").
			Column("project_id").
			Column("slug").
			Exec(ctx); err != nil {
			return err
		}

		// TODO: unique index for project_id and slug
		// TODO: unique index for project_id and name

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		type Environment struct {
			bun.BaseModel `bun:"table:environments"`
		}

		if _, err := db.NewDropTable().
			Model((*Environment)(nil)).
			IfExists().
			Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewDropIndex().
			Model((*Environment)(nil)).
			IfExists().
			Index("environments_project_id_and_slug_idx").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
