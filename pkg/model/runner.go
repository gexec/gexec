package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*Runner)(nil)
)

// Runner defines the model for runners table.
type Runner struct {
	bun.BaseModel `bun:"table:runners"`

	ID        string    `bun:",pk,type:varchar(20)"`
	ProjectID string    `bun:",nullzero,type:varchar(20)"`
	Project   *Project  `bun:"rel:belongs-to,join:project_id=id"`
	Slug      string    `bun:"type:varchar(255)"`
	Name      string    `bun:"type:varchar(255)"`
	Token     string    `bun:"type:varchar(255)"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Runner) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}

		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		if m.ID == "" {
			m.ID = strings.ToLower(uniuri.NewLen(uniuri.UUIDLen))
		}

		m.UpdatedAt = time.Now()
	}

	return nil
}

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *Runner) SerializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	nonce, err := generateNonce(gcm.NonceSize())

	if err != nil {
		return err
	}

	if m.Token != "" {
		m.Token = encryptSecret(gcm, nonce, m.Token)
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Runner) DeserializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	if m.Token != "" {
		decrypted, err := decryptSecret(gcm, m.Token)

		if err != nil {
			return err
		}

		m.Token = decrypted
	}

	return nil
}
