package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Output struct {
			bun.BaseModel `bun:"table:outputs"`

			ID          string    `bun:",pk,type:varchar(20)"`
			ExecutionID string    `bun:"type:varchar(20)"`
			Content     string    `bun:"type:text"`
			CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Output)(nil)).
			WithForeignKeys().
			ForeignKey(`(execution_id) REFERENCES executions (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Output struct {
			bun.BaseModel `bun:"table:outputs"`
		}

		_, err := db.NewDropTable().
			Model((*Output)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
