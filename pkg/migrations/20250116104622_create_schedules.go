package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Schedule struct {
			bun.BaseModel `bun:"table:schedules"`

			ID         string    `bun:",pk,type:varchar(20)"`
			ProjectID  string    `bun:"type:varchar(20)"`
			TemplateID string    `bun:"type:varchar(20)"`
			Slug       string    `bun:"type:varchar(255)"`
			Name       string    `bun:"type:varchar(255)"`
			Cron       string    `bun:"type:varchar(255)"`
			Active     bool      `bun:"type:bool"`
			CreatedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Schedule)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			ForeignKey(`(template_id) REFERENCES templates (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Schedule struct {
			bun.BaseModel `bun:"table:schedules"`
		}

		_, err := db.NewDropTable().
			Model((*Schedule)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
