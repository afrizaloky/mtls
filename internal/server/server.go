package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"mtls-server/internal/config"
)

type Server struct {
	http *http.Server
}

func New(cfg config.Config) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(cfg.ServerCertFile(), cfg.ServerKeyFile())
	if err != nil {
		return nil, fmt.Errorf("load server certificate: %w", err)
	}

	caPEM, err := os.ReadFile(cfg.ClientCAFile())
	if err != nil {
		return nil, fmt.Errorf("read client CA: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse client CA: no certificates found")
	}

	codec := newCodec(cfg.StaticKey())
	handler := newEchoHandler(codec)
	routes := setupRoutes(handler, codec)
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}
	srv := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           routes,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return &Server{http: srv}, nil
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("starting server", "addr", s.http.Addr)
	listenErr := make(chan error, 1)
	go func() {
		listenErr <- s.http.ListenAndServeTLS("", "")
	}()

	select {
	case err := <-listenErr:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("serve HTTPS: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down", "addr", s.http.Addr)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown failed", "error", err)
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	}
}
