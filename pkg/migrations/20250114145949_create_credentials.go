package migrations

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		type CredentialShell struct {
			Username   string `bun:"type:varchar(255)"`
			Password   string `bun:"type:varchar(255)"`
			PrivateKey string `bun:"type:text"`
		}

		type CredentialLogin struct {
			Username string `bun:"type:varchar(255)"`
			Password string `bun:"type:varchar(255)"`
		}

		type Credential struct {
			bun.BaseModel `bun:"table:credentials"`

			ID        string          `bun:",pk,type:varchar(20)"`
			ProjectID string          `bun:"type:varchar(20)"`
			Slug      string          `bun:"type:varchar(255)"`
			Name      string          `bun:"type:varchar(255)"`
			Kind      string          `bun:"type:varchar(255)"`
			Override  bool            `bun:"type:bool"`
			Shell     CredentialShell `bun:"embed:shell_"`
			Login     CredentialLogin `bun:"embed:login_"`
			CreatedAt time.Time       `bun:",nullzero,notnull,default:current_timestamp"`
			UpdatedAt time.Time       `bun:",nullzero,notnull,default:current_timestamp"`
		}

		_, err := db.NewCreateTable().
			Model((*Credential)(nil)).
			WithForeignKeys().
			ForeignKey(`(project_id) REFERENCES projects (id) ON DELETE CASCADE`).
			Exec(ctx)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		type Credential struct {
			bun.BaseModel `bun:"table:credentials"`
		}

		_, err := db.NewDropTable().
			Model((*Credential)(nil)).
			IfExists().
			Exec(ctx)

		return err
	})
}
