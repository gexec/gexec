package model

// CredentialShell represents credentials for shells.
type CredentialShell struct {
	Username   string `bun:"type:varchar(255)"`
	Password   string `bun:"type:varchar(255)"`
	PrivateKey string `bun:"type:text"`
}

// SerializeSecret ensures to encrypt all related secrets stored on the database.
func (m *CredentialShell) SerializeSecret(passphrase string) error {
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

	if m.PrivateKey != "" {
		m.PrivateKey = encryptSecret(gcm, nonce, m.PrivateKey)
	}

	return nil
}

// DeserializeSecret ensures to decrypt all related secrets stored on the database.
func (m *CredentialShell) DeserializeSecret(passphrase string) error {
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

	if m.PrivateKey != "" {
		decrypted, err := decryptSecret(gcm, m.PrivateKey)

		if err != nil {
			return err
		}

		m.PrivateKey = decrypted
	}

	return nil
}
