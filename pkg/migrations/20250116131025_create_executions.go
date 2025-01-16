package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Execution struct {
			bun.BaseModel `bun:"table:executions"`

			ID          string    `bun:",pk,type:varchar(20)"`
			ProjectID   string    `bun:"type:varchar(20)"`
			TemplateID  string    `bun:"type:varchar(20)"`
			Status      string    `bun:"type:varchar(255)"`
			Playbook    string    `bun:"type:varchar(255)"`
			Environment string    `bun:"type:varchar(255)"`
			Secret      string    `bun:"type:varchar(255)"`
			Limit       string    `bun:"type:varchar(255)"`
			Branch      string    `bun:"type:varchar(255)"`
			Debug       bool      `bun:"type:bool"`
			CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Execution)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			ForeignKey(`(template_id) REFERENCES templates (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Execution struct {
			bun.BaseModel `bun:"table:executions"`
		}

		_, err := db.NewDropTable().
			Model((*Execution)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
