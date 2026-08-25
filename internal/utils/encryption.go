package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"jaga-backend/internal/config"
)

func loadKey(cfg *config.Config) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(cfg.EvidenceEncryptionKey)
	if err != nil {
		return nil, errors.New("kunci enkripsi tidak valid")
	}
	if len(key) != 32 {
		return nil, errors.New("kunci enkripsi harus 32 byte (AES-256)")
	}
	return key, nil
}

// EncryptFile mengenkripsi data file mentah, mengembalikan bytes siap simpan (nonce + ciphertext).
func EncryptFile(cfg *config.Config, plainData []byte) ([]byte, error) {
	key, err := loadKey(cfg)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal menempelkan nonce di depan ciphertext secara otomatis lewat parameter kedua.
	ciphertext := gcm.Seal(nonce, nonce, plainData, nil)
	return ciphertext, nil
}

// DecryptFile mengembalikan bytes hasil enkripsi (nonce + ciphertext) menjadi data asli.
func DecryptFile(cfg *config.Config, encryptedData []byte) ([]byte, error) {
	key, err := loadKey(cfg)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("data terenkripsi tidak valid")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
