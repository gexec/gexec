package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Project struct {
			bun.BaseModel `bun:"table:projects"`

			ID        string    `bun:",pk,type:varchar(20)"`
			Slug      string    `bun:",unique,type:varchar(255)"`
			Name      string    `bun:"type:varchar(255)"`
			CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Project)(nil)).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Project struct {
			bun.BaseModel `bun:"table:projects"`
		}

		_, err := db.NewDropTable().
			Model((*Project)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
