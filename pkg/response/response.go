package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func JSON(response http.ResponseWriter, statusCode int, payload any) error {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	return json.NewEncoder(response).Encode(payload)
}

func Error(response http.ResponseWriter, statusCode int, message string) error {
	return JSON(response, statusCode, errorResponse{Error: message})
}

func WriteError(logger *slog.Logger, response http.ResponseWriter, request *http.Request, statusCode int, message string) {
	if err := Error(response, statusCode, message); err != nil {
		logger.ErrorContext(request.Context(), "failed to write error response", slog.Any("error", err))
	}
}
