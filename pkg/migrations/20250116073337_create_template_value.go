package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TemplateValue struct {
			bun.BaseModel `bun:"table:template_values"`

			ID        string    `bun:",pk,type:varchar(20)"`
			SurveyID  string    `bun:"type:varchar(20)"`
			Name      string    `bun:"type:varchar(255)"`
			Value     string    `bun:"type:text"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*TemplateValue)(nil)).
			WithForeignKeys().
			ForeignKey(`(survey_id) REFERENCES template_surveys (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type TemplateValue struct {
			bun.BaseModel `bun:"table:template_values"`
		}

		_, err := db.NewDropTable().
			Model((*TemplateValue)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
