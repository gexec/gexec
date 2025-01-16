package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type EnvironmentValue struct {
			bun.BaseModel `bun:"table:environment_values"`

			ID            string    `bun:",pk,type:varchar(20)"`
			EnvironmentID string    `bun:"type:varchar(20)"`
			Kind          string    `bun:"type:varchar(255)"`
			Name          string    `bun:"type:varchar(255)"`
			Content       string    `bun:"type:text"`
			CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		if _, err := db.NewCreateTable().
			Model((*EnvironmentValue)(nil)).
			WithForeignKeys().
			ForeignKey(`(environment_id) REFERENCES environments (id) ON DELETE CASCADE`).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		type EnvironmentValue struct {
			bun.BaseModel `bun:"table:environment_secrets"`
		}

		if _, err := db.NewDropTable().
			Model((*EnvironmentValue)(nil)).
			IfExists().
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}
