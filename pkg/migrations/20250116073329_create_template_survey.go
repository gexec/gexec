package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TemplateSurvey struct {
			bun.BaseModel `bun:"table:template_surveys"`

			ID          string    `bun:",pk,type:varchar(20)"`
			TemplateID  string    `bun:"type:varchar(20)"`
			Name        string    `bun:"type:varchar(255)"`
			Title       string    `bun:"type:varchar(255)"`
			Description string    `bun:"type:text"`
			Kind        string    `bun:"type:varchar(255)"`
			Required    bool      `bun:"type:bool"`
			CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*TemplateSurvey)(nil)).
			WithForeignKeys().
			ForeignKey(`(template_id) REFERENCES templates (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type TemplateSurvey struct {
			bun.BaseModel `bun:"table:template_surveys"`
		}

		_, err := db.NewDropTable().
			Model((*TemplateSurvey)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
