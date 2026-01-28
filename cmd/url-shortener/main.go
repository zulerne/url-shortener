package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zulerne/url-shortener/internal/config"
	"github.com/zulerne/url-shortener/internal/lib/logger"
	"github.com/zulerne/url-shortener/internal/server"
	"github.com/zulerne/url-shortener/internal/server/handler"
	"github.com/zulerne/url-shortener/internal/storage/sqlite"
)

func main() {
	cfg := config.MustLoad()

	logger.SetupLogger(cfg.Env)
	slog.Info("Starting url-shortener", "env", cfg.Env)
	slog.Debug("Debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv := &server.Server{
		HttpServer: &http.Server{
			Addr:         cfg.HttpConfig.Address,
			Handler:      handler.NewHandler(storage, cfg.AliasLength),
			ReadTimeout:  cfg.HttpConfig.Timeout,
			WriteTimeout: cfg.HttpConfig.Timeout,
			IdleTimeout:  cfg.HttpConfig.IdleTimeout,
		},
		ShutdownTimeout: cfg.HttpConfig.Timeout,
	}
	// todo: Maybe remove blocking operation

	if err = srv.Listen(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
