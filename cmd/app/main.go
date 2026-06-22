package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fangimal/TeamTask/internal/config"
	"github.com/fangimal/TeamTask/internal/delivery/router"
	"github.com/fangimal/TeamTask/internal/logger"
)

const configPath = "configs/config.yaml"

func main() {
	cfg, err := config.Load(configPath)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	appLogger := logger.New(cfg.Logger)
	httpServer := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router.New(appLogger),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		appLogger.Info("starting http server", slog.String("address", httpServer.Addr))
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case signal := <-shutdownSignals:
		appLogger.Info("shutdown signal received", slog.String("signal", signal.String()))
	case err = <-serverErrors:
		if err != nil {
			appLogger.Error("http server failed", slog.Any("error", err))
			os.Exit(1)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err = httpServer.Shutdown(shutdownContext); err != nil {
		appLogger.Error("failed to shutdown http server gracefully", slog.Any("error", err))
		os.Exit(1)
	}

	appLogger.Info("http server stopped")
}
