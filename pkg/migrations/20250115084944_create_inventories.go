package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Inventory struct {
			bun.BaseModel `bun:"table:inventories"`

			ID           string    `bun:",pk,type:varchar(20)"`
			ProjectID    string    `bun:"type:varchar(20)"`
			RepositoryID string    `bun:",nullzero,type:varchar(20)"`
			CredentialID string    `bun:",nullzero,type:varchar(20)"`
			BecomeID     string    `bun:",nullzero,type:varchar(20)"`
			Slug         string    `bun:"type:varchar(255)"`
			Name         string    `bun:"type:varchar(255)"`
			Kind         string    `bun:"type:varchar(255)"`
			Content      string    `bun:"type:text"`
			CreatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Inventory)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			ForeignKey(`(repository_id) REFERENCES repositories (id) ON DELETE CASCADE`).
			ForeignKey(`(credential_id) REFERENCES credentials (id) ON DELETE CASCADE`).
			ForeignKey(`(become_id) REFERENCES credentials (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Inventory struct {
			bun.BaseModel `bun:"table:inventories"`
		}

		_, err := db.NewDropTable().
			Model((*Inventory)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
