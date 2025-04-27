package model

import (
	"context"
	"strings"
	"time"

	"github.com/dchest/uniuri"
	"github.com/uptrace/bun"
)

var (
	_ bun.BeforeAppendModelHook = (*EnvironmentSecret)(nil)
)

// EnvironmentSecret defines the model for environment_secrets table.
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

// BeforeAppendModel implements the bun hook interface.
func (m *EnvironmentSecret) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *EnvironmentSecret) SerializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	nonce, err := generateNonce(gcm.NonceSize())

	if err != nil {
		return err
	}

	if m.Content != "" {
		m.Content = encryptSecret(gcm, nonce, m.Content)
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *EnvironmentSecret) DeserializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	if m.Content != "" {
		decrypted, err := decryptSecret(gcm, m.Content)

		if err != nil {
			return err
		}

		m.Content = decrypted
	}

	return nil
}
