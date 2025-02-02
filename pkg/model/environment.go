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
	_ bun.BeforeAppendModelHook = (*EnvironmentValue)(nil)
	_ bun.BeforeAppendModelHook = (*Environment)(nil)
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

// EnvironmentValue defines the model for environment_values table.
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

// BeforeAppendModel implements the bun hook interface.
func (m *EnvironmentValue) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *EnvironmentValue) SerializeSecret(passphrase string) error {
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
func (m *EnvironmentValue) DeserializeSecret(passphrase string) error {
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

// Environment defines the model for environments table.
type Environment struct {
	bun.BaseModel `bun:"table:environments"`

	ID        string               `bun:",pk,type:varchar(20)"`
	ProjectID string               `bun:"type:varchar(20)"`
	Project   *Project             `bun:"rel:belongs-to,join:project_id=id"`
	Slug      string               `bun:"type:varchar(255)"`
	Name      string               `bun:"type:varchar(255)"`
	Secrets   []*EnvironmentSecret `bun:"rel:has-many,join:id=environment_id"`
	Values    []*EnvironmentValue  `bun:"rel:has-many,join:id=environment_id"`
	CreatedAt time.Time            `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time            `bun:",nullzero,notnull,default:current_timestamp"`
}

// BeforeAppendModel implements the bun hook interface.
func (m *Environment) BeforeAppendModel(_ context.Context, query bun.Query) error {
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
func (m *Environment) SerializeSecret(passphrase string) error {
	for _, row := range m.Secrets {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Values {
		if err := row.SerializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *Environment) DeserializeSecret(passphrase string) error {
	for _, row := range m.Secrets {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	for _, row := range m.Values {
		if err := row.DeserializeSecret(passphrase); err != nil {
			return err
		}
	}

	return nil
}
