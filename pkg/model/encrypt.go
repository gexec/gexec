package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

var (
	// ErrWrongEncryptionPassphrase defines the error if it fails to validate the encryption passphrase.
	ErrWrongEncryptionPassphrase = fmt.Errorf("wrong encryption passphrase")
)

func prepareEncrypt(passphrase string) (cipher.AEAD, error) {
	c, err := aes.NewCipher([]byte(passphrase))

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)

	if err != nil {
		return nil, err
	}

	return gcm, nil
}

func generateNonce(size int) ([]byte, error) {
	nonce := make(
		[]byte,
		size,
	)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}

	return nonce, nil
}

func encryptSecret(gcm cipher.AEAD, nonce []byte, secret string) string {
	return base64.StdEncoding.EncodeToString(
		gcm.Seal(
			nonce,
			nonce,
			[]byte(secret),
			nil,
		),
	)
}

func decryptSecret(gcm cipher.AEAD, secret string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(secret)

	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()

	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	decrypted, err := gcm.Open(
		nil,
		encrypted[:nonceSize],
		encrypted[nonceSize:],
		nil,
	)

	if err != nil {
		if err.Error() == "cipher: message authentication failed" {
			return "", ErrWrongEncryptionPassphrase
		}

		return "", err
	}

	return string(decrypted), nil
}
