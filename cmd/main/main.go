package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main/internal/app"
	"main/internal/config"
	"main/internal/infrastructure"
	"main/internal/infrastructure/database"

	"main/internal/shared/ports"

	"gorm.io/gorm"
)

var seedFlag = flag.Bool("seed", false, "create the admin from ADMIN_EMAIL/ADMIN_PASSWORD/ADMIN_NAME, then exit")

func main() {
	flag.Parse()

	// run returns instead of exiting, so its deferred cleanup always happens.
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ctx is cancelled on Ctrl+C (SIGINT) or on `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config:\n%w", err)
	}

	adapters, db, err := infrastructure.Initialize(ctx, cfg)
	if err != nil {
		return err
	}

	logger := adapters.Logger
	defer func() { _ = infrastructure.CloseLogger(logger) }()

	if *seedFlag {
		return runSeed(ctx, adapters, db)
	}

	logger.Info("application starting", slog.String("mode", cfg.App.Mode))

	router := app.SetupRouter(cfg, adapters, db)

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Bind the ports first, so "port in use" is reported before anything says "listening".
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		logger.Error("http server failed", slog.Any("error", err))
		return errors.Join(fmt.Errorf("http server: %w", err), infrastructure.Shutdown(adapters, db))
	}

	logger.Info("http server listening", slog.String("port", cfg.Server.Port))

	// Serve blocks, so each server runs in a goroutine and reports an unexpected stop on serverErr.
	serverErr := make(chan error, 2)
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- fmt.Errorf("http server: %w", err)
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErr:
		runErr = err
		logger.Error("server failed", slog.Any("error", err))
	}
	// Restore the default signal handling: a second Ctrl+C now kills the process at once.
	stop()

	// End every WebSocket and SSE connection first: srv.Shutdown doesn't track hijacked
	// WebSocket connections, and it would wait for SSE streams until the timeout.
	// adapters.Realtime.Close()

	// Stop accepting new requests and give in-flight ones time to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", slog.Any("error", err))
		runErr = errors.Join(runErr, err)
	}

	if err := infrastructure.Shutdown(adapters, db); err != nil {
		logger.Error("closing connections failed", slog.Any("error", err))
		runErr = errors.Join(runErr, err)
	}

	logger.Info("application stopped")

	return runErr
}

// runSeed creates the admin and closes the connections, without starting the API.
func runSeed(ctx context.Context, adapters *ports.Adapters, db *gorm.DB) error {
	err := database.SeedInitialData(db.WithContext(ctx))
	if err != nil {
		adapters.Logger.Error("seed failed", slog.Any("error", err))
	} else {
		adapters.Logger.Info("database seeded")
	}

	return errors.Join(err, infrastructure.Shutdown(adapters, db))
}
