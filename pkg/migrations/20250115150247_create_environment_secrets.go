package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type EnvironmentSecret struct {
			bun.BaseModel `bun:"table:environment_secrets"`

			ID            string    `bun:",pk,type:varchar(20)"`
			EnvironmentID string    `bun:"type:varchar(20)"`
			Kind          string    `bun:"type:varchar(255)"`
			Name          string    `bun:"type:varchar(255)"`
			Content       string    `bun:"type:text"`
			CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*EnvironmentSecret)(nil)).
			WithForeignKeys().
			ForeignKey(`(environment_id) REFERENCES environments (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type EnvironmentSecret struct {
			bun.BaseModel `bun:"table:environment_secrets"`
		}

		_, err := db.NewDropTable().
			Model((*EnvironmentSecret)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
