//go:build ignore

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type envelope struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: MTLS_STATIC_KEY=<key> go run scripts/decrypt-response.go <response.json>")
		os.Exit(1)
	}
	key, err := base64.StdEncoding.DecodeString(os.Getenv("MTLS_STATIC_KEY"))
	if err != nil || len(key) != 32 {
		fmt.Fprintln(os.Stderr, "error: MTLS_STATIC_KEY must be base64 of 32 bytes")
		os.Exit(1)
	}
	response, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	var env envelope
	if err := json.Unmarshal(response, &env); err != nil {
		fmt.Fprintln(os.Stderr, "error: invalid response:", err)
		os.Exit(1)
	}
	nonce, err := base64.StdEncoding.DecodeString(env.Nonce)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: invalid nonce:", err)
		os.Exit(1)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(env.Ciphertext)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: invalid ciphertext:", err)
		os.Exit(1)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: decryption failed:", err)
		os.Exit(1)
	}
	fmt.Println(string(plaintext))
}
