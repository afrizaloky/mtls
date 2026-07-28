package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"mtls-server/internal/config"
	"mtls-server/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}
	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("server setup error", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil {
		slog.Error("server run error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
