// Package main is the entrypoint for the Appointly API server.
// It performs application bootstrap in the correct order:
//  1. Load configuration from environment
//  2. Initialize logger
//  3. Connect to infrastructure (database, Redis)
//  4. Run database migrations (if enabled)
//  5. Initialize repositories, use cases, handlers
//  6. Start HTTP server
//  7. Wait for shutdown signal, then gracefully stop
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/appointly/appointly/backend/internal/config"
	"github.com/appointly/appointly/backend/internal/pkg/logger"
	"github.com/appointly/appointly/backend/internal/server"
)

// Version is set at build time via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	// ---------------------------------------------------------------------------
	// 1. Load configuration
	// ---------------------------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return 1
	}

	// ---------------------------------------------------------------------------
	// 2. Initialize logger
	// ---------------------------------------------------------------------------
	log := logger.New(logger.Config{
		Level:     cfg.Log.Level,
		Format:    cfg.Log.Format,
		AddSource: cfg.Log.AddSource,
	})

	log.Info("appointly api starting",
		"version", Version,
		"env", cfg.App.Env,
	)

	// ---------------------------------------------------------------------------
	// 3. Connect to infrastructure
	// NOTE: Database and Redis connection initialization will be added here
	// as we build the repository layer.
	// ---------------------------------------------------------------------------

	// ---------------------------------------------------------------------------
	// 4. Initialize server
	// ---------------------------------------------------------------------------
	srv := server.New(server.Dependencies{
		Config: cfg,
		Logger: log,
	})

	// ---------------------------------------------------------------------------
	// 5. Start server in background
	// ---------------------------------------------------------------------------
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// ---------------------------------------------------------------------------
	// 6. Wait for shutdown signal
	// ---------------------------------------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("received shutdown signal", "signal", sig.String())
	case err := <-serverErr:
		log.Error("server error", "error", err)
		return 1
	}

	// ---------------------------------------------------------------------------
	// 7. Graceful shutdown
	// ---------------------------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		return 1
	}

	log.Info("server stopped gracefully")

	// Give background jobs a moment to finish
	time.Sleep(100 * time.Millisecond)

	return 0
}
