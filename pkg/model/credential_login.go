package model

// CredentialLogin represents credentials for logins.
type CredentialLogin struct {
	Username string `bun:"type:varchar(255)"`
	Password string `bun:"type:varchar(255)"`
}

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *CredentialLogin) SerializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	nonce, err := generateNonce(gcm.NonceSize())

	if err != nil {
		return err
	}

	if m.Password != "" {
		m.Password = encryptSecret(gcm, nonce, m.Password)
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *CredentialLogin) DeserializeSecret(passphrase string) error {
	gcm, err := prepareEncrypt(passphrase)

	if err != nil {
		return err
	}

	if m.Password != "" {
		decrypted, err := decryptSecret(gcm, m.Password)

		if err != nil {
			return err
		}

		m.Password = decrypted
	}

	return nil
}
