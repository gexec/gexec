package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Repository struct {
			bun.BaseModel `bun:"table:repositories"`

			ID           string    `bun:",pk,type:varchar(20)"`
			ProjectID    string    `bun:"type:varchar(20)"`
			CredentialID string    `bun:"type:varchar(20)"`
			Slug         string    `bun:"type:varchar(255)"`
			Name         string    `bun:"type:varchar(255)"`
			URL          string    `bun:"type:text"`
			Branch       string    `bun:"type:varchar(255)"`
			CreatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		if _, err := db.NewCreateTable().
			Model((*Repository)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			ForeignKey(`(credential_id) REFERENCES credentials (id) ON DELETE CASCADE`).
			Exec(ctx); err != nil {
			return err
		}

		if _, err := db.NewCreateIndex().
			Model((*Repository)(nil)).
			Index("prooject_id_and_slug_idx").
			Column("project_id").
			Column("slug").
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		type Repository struct {
			bun.BaseModel `bun:"table:repositories"`
		}

		_, err := db.NewDropTable().
			Model((*Repository)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
