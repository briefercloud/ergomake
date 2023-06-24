package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

func Encrypt(encryptionKey string, text string) (string, error) {
	key, err := hex.DecodeString(encryptionKey)
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)

	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext, plaintext)

	return hex.EncodeToString(iv) + ":" + hex.EncodeToString(ciphertext), nil
}

func Decrypt(encryptionKey string, hash string) (string, error) {
	key, err := hex.DecodeString(encryptionKey)
	if err != nil {
		return "", err
	}

	textParts := bytes.SplitN([]byte(hash), []byte(":"), 2)
	if len(textParts) != 2 {
		return "", errors.New("invalid hash format")
	}

	iv, err := hex.DecodeString(string(textParts[0]))
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(string(textParts[1]))
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
