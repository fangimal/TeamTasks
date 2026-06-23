package httpdelivery

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	logger       *slog.Logger
	dbChecker    HealthChecker
	cacheChecker HealthChecker
}

func NewHealthHandler(logger *slog.Logger, dbChecker HealthChecker, cacheChecker HealthChecker) *HealthHandler {
	return &HealthHandler{
		logger:       logger,
		dbChecker:    dbChecker,
		cacheChecker: cacheChecker,
	}
}

func (handler *HealthHandler) Check(response http.ResponseWriter, request *http.Request) {
	statusCode := http.StatusOK
	databaseStatus := "connected"
	cacheStatus := "connected"
	applicationStatus := "ok"

	if err := handler.dbChecker.Ping(request.Context()); err != nil {
		handler.logger.ErrorContext(request.Context(), "database healthcheck failed", slog.Any("error", err))
		statusCode = http.StatusServiceUnavailable
		databaseStatus = "disconnected"
		applicationStatus = "error"
	}

	if err := handler.cacheChecker.Ping(request.Context()); err != nil {
		handler.logger.ErrorContext(request.Context(), "cache healthcheck failed", slog.Any("error", err))
		cacheStatus = "disconnected"
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	payload := map[string]string{
		"status":    applicationStatus,
		"database":  databaseStatus,
		"cache":     cacheStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(response).Encode(payload); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write health response", slog.Any("error", err))
	}
}
