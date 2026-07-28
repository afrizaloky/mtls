package server

import (
	"encoding/base64"
	"encoding/json"
	"io"
)

type envelope struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func readPlaintext(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

func buildResponse(plaintext []byte, codec *codec) ([]byte, error) {
	nonce, ciphertext, err := codec.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}
	env := envelope{
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}
	return json.Marshal(env)
}
