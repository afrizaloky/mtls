package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

type codec struct {
	key       [32]byte
	nonceSize int
	overhead  int
}

func newCodec(key [32]byte) *codec {
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	return &codec{
		key:       key,
		nonceSize: gcm.NonceSize(),
		overhead:  gcm.Overhead(),
	}
}

func (c *codec) Decrypt(nonce, ciphertext []byte) ([]byte, error) {
	aesCipher, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (c *codec) Encrypt(plaintext []byte) (nonce, ciphertext []byte, err error) {
	aesCipher, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}
