package httpdelivery

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type HealthHandler struct {
	logger *slog.Logger
}

func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

func (handler *HealthHandler) Check(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	payload := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(response).Encode(payload); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write health response", slog.Any("error", err))
	}
}
