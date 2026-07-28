package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type envelope struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

type requestParser struct {
	dec  *json.Decoder
	seen map[string]bool
}

var errInvalidRequest = errors.New("invalid request")

func parseRequest(r io.Reader, maxBytes int64, codec *codec) ([]byte, error) {
	p := &requestParser{
		dec:  json.NewDecoder(r),
		seen: make(map[string]bool),
	}

	tok, err := p.dec.Token()
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return nil, maxBytesErr
		}
		return nil, errInvalidRequest
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, errInvalidRequest
	}

	var nonceB64, cipherB64 string
	var foundNonce, foundCipher bool

	for p.dec.More() {
		keyTok, err := p.dec.Token()
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				return nil, maxBytesErr
			}
			return nil, errInvalidRequest
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, errInvalidRequest
		}
		if p.seen[key] {
			return nil, errInvalidRequest
		}
		p.seen[key] = true
		if key != "nonce" && key != "ciphertext" {
			return nil, errInvalidRequest
		}

		valTok, err := p.dec.Token()
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				return nil, maxBytesErr
			}
			return nil, errInvalidRequest
		}
		val, ok := valTok.(string)
		if !ok {
			return nil, errInvalidRequest
		}
		if strings.TrimSpace(val) == "" {
			return nil, errInvalidRequest
		}

		if key == "nonce" {
			nonceB64 = val
			foundNonce = true
		} else {
			cipherB64 = val
			foundCipher = true
		}
	}

	tok, err = p.dec.Token()
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return nil, maxBytesErr
		}
		return nil, errInvalidRequest
	}
	if _, ok := tok.(json.Delim); !ok {
		return nil, errInvalidRequest
	}

	_, err = p.dec.Token()
	if err == nil {
		return nil, errInvalidRequest
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return nil, maxBytesErr
	}
	if !errors.Is(err, io.EOF) {
		return nil, errInvalidRequest
	}

	if !foundNonce || !foundCipher {
		return nil, errInvalidRequest
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, errInvalidRequest
	}
	ciphertext, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, errInvalidRequest
	}

	if len(nonce) != codec.nonceSize {
		return nil, errInvalidRequest
	}
	if len(ciphertext) < codec.overhead {
		return nil, errInvalidRequest
	}

	plaintext, err := codec.Decrypt(nonce, ciphertext)
	if err != nil {
		return nil, errInvalidRequest
	}
	return plaintext, nil
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
	data, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	return data, nil
}
