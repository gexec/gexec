package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type Event struct {
			bun.BaseModel `bun:"table:events"`

			ID             string    `bun:",pk,type:varchar(20)"`
			UserID         string    `bun:"type:varchar(20)"`
			UserHandle     string    `bun:"type:varchar(255)"`
			UserDisplay    string    `bun:"type:varchar(255)"`
			ProjectID      string    `bun:"type:varchar(20)"`
			ProjectDisplay string    `bun:"type:varchar(255)"`
			ObjectID       string    `bun:"type:varchar(20)"`
			ObjectDisplay  string    `bun:"type:varchar(255)"`
			ObjectType     string    `bun:"type:varchar(255)"`
			Action         string    `bun:"type:varchar(255)"`
			RawAttrs       string    `bun:"attrs,nullzero,type:text"`
			CreatedAt      time.Time `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Event)(nil)).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Event struct {
			bun.BaseModel `bun:"table:events"`
		}

		_, err := db.NewDropTable().
			Model((*Event)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
