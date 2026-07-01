package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

const marker = "enc:v1:"

type Cipher struct{ aead cipher.AEAD }

func New(hexKey string) (*Cipher, error) {
	if hexKey == "" {
		return nil, nil
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(plaintext string) string {
	if c == nil || plaintext == "" || strings.HasPrefix(plaintext, marker) {
		return plaintext
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}
	ct := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return marker + base64.RawStdEncoding.EncodeToString(ct)
}

func (c *Cipher) Decrypt(stored string) string {
	if c == nil || !strings.HasPrefix(stored, marker) {
		return stored
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(stored, marker))
	if err != nil || len(raw) < c.aead.NonceSize() {
		return stored
	}
	nonce, ct := raw[:c.aead.NonceSize()], raw[c.aead.NonceSize():]
	pt, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return stored
	}
	return string(pt)
}
