package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type TemplateVault struct {
			bun.BaseModel `bun:"table:template_vaults"`

			ID           string    `bun:",pk,type:varchar(20)"`
			TemplateID   string    `bun:"type:varchar(20)"`
			CredentialID string    `bun:",nullzero,type:varchar(20)"`
			Name         string    `bun:"type:varchar(255)"`
			Kind         string    `bun:"type:varchar(255)"`
			Script       string    `bun:"type:text"`
			CreatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*TemplateVault)(nil)).
			WithForeignKeys().
			ForeignKey(`(template_id) REFERENCES templates (id) ON DELETE CASCADE`).
			ForeignKey(`(credential_id) REFERENCES credentials (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type TemplateVault struct {
			bun.BaseModel `bun:"table:template_vaults"`
		}

		_, err := db.NewDropTable().
			Model((*TemplateVault)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
