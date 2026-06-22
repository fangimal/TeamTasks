package httpdelivery

import (
	"log/slog"
	"net/http"

	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
	"github.com/fangimal/TeamTask/pkg/response"
)

type ProtectedHandler struct {
	logger *slog.Logger
}

type pingResponse struct {
	Status string `json:"status"`
	UserID int64  `json:"user_id"`
}

func NewProtectedHandler(logger *slog.Logger) *ProtectedHandler {
	return &ProtectedHandler{
		logger: logger,
	}
}

func (handler *ProtectedHandler) Ping(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := response.JSON(responseWriter, http.StatusOK, pingResponse{Status: "ok", UserID: userID}); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write protected ping response", slog.Any("error", err))
	}
}
