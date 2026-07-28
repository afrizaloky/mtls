package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
)

const (
	serverCertEnv = "MTLS_SERVER_CERT_FILE"
	serverKeyEnv  = "MTLS_SERVER_KEY_FILE"
	clientCAEnv   = "MTLS_CLIENT_CA_FILE"
	staticKeyEnv  = "MTLS_STATIC_KEY"
	listenAddrEnv = "MTLS_LISTEN_ADDR"
	defaultAddr   = ":8443"
)

type Config struct {
	serverCertFile string
	serverKeyFile  string
	clientCAFile   string
	staticKey      [32]byte
	listenAddr     string
}

func (c Config) ServerCertFile() string { return c.serverCertFile }

func (c Config) ServerKeyFile() string { return c.serverKeyFile }

func (c Config) ClientCAFile() string { return c.clientCAFile }

func (c Config) StaticKey() [32]byte { return c.staticKey }

func (c Config) ListenAddr() string { return c.listenAddr }

func Load() (Config, error) {
	serverCertFile, err := requiredPath(serverCertEnv)
	if err != nil {
		return Config{}, err
	}
	serverKeyFile, err := requiredPath(serverKeyEnv)
	if err != nil {
		return Config{}, err
	}
	clientCAFile, err := requiredPath(clientCAEnv)
	if err != nil {
		return Config{}, err
	}

	staticKeyEncoded := os.Getenv(staticKeyEnv)
	if staticKeyEncoded == "" {
		return Config{}, fmt.Errorf("%s is required", staticKeyEnv)
	}
	staticKeyBytes, err := base64.StdEncoding.DecodeString(staticKeyEncoded)
	if err != nil {
		return Config{}, fmt.Errorf("%s must be valid base64: %w", staticKeyEnv, err)
	}
	if len(staticKeyBytes) != 32 {
		return Config{}, fmt.Errorf("%s must decode to exactly 32 bytes", staticKeyEnv)
	}
	var staticKey [32]byte
	copy(staticKey[:], staticKeyBytes)

	listenAddr := os.Getenv(listenAddrEnv)
	if listenAddr == "" {
		listenAddr = defaultAddr
	}
	if _, err := net.ResolveTCPAddr("tcp", listenAddr); err != nil {
		return Config{}, fmt.Errorf("%s is invalid: %w", listenAddrEnv, err)
	}

	return Config{
		serverCertFile: serverCertFile,
		serverKeyFile:  serverKeyFile,
		clientCAFile:   clientCAFile,
		staticKey:      staticKey,
		listenAddr:     listenAddr,
	}, nil
}

func requiredPath(envName string) (string, error) {
	path := os.Getenv(envName)
	if path == "" {
		return "", fmt.Errorf("%s is required", envName)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s must name a readable file: %w", envName, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s must name a readable file", envName)
	}
	return path, nil
}
