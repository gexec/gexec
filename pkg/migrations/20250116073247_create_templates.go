package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Template struct {
			bun.BaseModel `bun:"table:templates"`

			ID            string    `bun:",pk,type:varchar(20)"`
			ProjectID     string    `bun:"type:varchar(20)"`
			RepositoryID  string    `bun:"type:varchar(20)"`
			InventoryID   string    `bun:"type:varchar(20)"`
			EnvironmentID string    `bun:"type:varchar(20)"`
			Slug          string    `bun:"type:varchar(255)"`
			Name          string    `bun:"type:varchar(255)"`
			Description   string    `bun:"type:text"`
			Playbook      string    `bun:"type:varchar(255)"`
			Arguments     string    `bun:"type:varchar(255)"`
			Limit         string    `bun:"type:varchar(255)"`
			Executor      string    `bun:"type:varchar(255)"`
			Branch        string    `bun:"type:varchar(255)"`
			Override      bool      `bun:"type:bool"`
			CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Template)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			ForeignKey(`(repository_id) REFERENCES repositories (id) ON DELETE CASCADE`).
			ForeignKey(`(inventory_id) REFERENCES inventories (id) ON DELETE CASCADE`).
			ForeignKey(`(environment_id) REFERENCES environments (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Template struct {
			bun.BaseModel `bun:"table:templates"`
		}

		_, err := db.NewDropTable().
			Model((*Template)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
