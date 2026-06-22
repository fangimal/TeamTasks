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
	logger        *slog.Logger
	healthChecker HealthChecker
}

func NewHealthHandler(logger *slog.Logger, healthChecker HealthChecker) *HealthHandler {
	return &HealthHandler{
		logger:        logger,
		healthChecker: healthChecker,
	}
}

func (handler *HealthHandler) Check(response http.ResponseWriter, request *http.Request) {
	statusCode := http.StatusOK
	databaseStatus := "connected"
	applicationStatus := "ok"

	if err := handler.healthChecker.Ping(request.Context()); err != nil {
		handler.logger.ErrorContext(request.Context(), "database healthcheck failed", slog.Any("error", err))
		statusCode = http.StatusServiceUnavailable
		databaseStatus = "disconnected"
		applicationStatus = "error"
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	payload := map[string]string{
		"status":    applicationStatus,
		"database":  databaseStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(response).Encode(payload); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write health response", slog.Any("error", err))
	}
}
