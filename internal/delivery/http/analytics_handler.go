package httpdelivery

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/fangimal/TeamTask/pkg/response"
)

type AnalyticsUseCase interface {
	GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error)
	GetTopUsersPerTeam(ctx context.Context) ([]*domain.TopUser, error)
	GetIntegrityViolations(ctx context.Context) ([]*domain.IntegrityViolation, error)
}

type AnalyticsHandler struct {
	logger    *slog.Logger
	analytics AnalyticsUseCase
}

func NewAnalyticsHandler(logger *slog.Logger, analytics AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		logger:    logger,
		analytics: analytics,
	}
}

func (handler *AnalyticsHandler) GetTeamStats(responseWriter http.ResponseWriter, request *http.Request) {
	stats, err := handler.analytics.GetTeamStats(request.Context())
	if err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to get team stats", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, stats); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write team stats response", slog.Any("error", err))
	}
}

func (handler *AnalyticsHandler) GetTopUsers(responseWriter http.ResponseWriter, request *http.Request) {
	users, err := handler.analytics.GetTopUsersPerTeam(request.Context())
	if err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to get top users", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, users); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write top users response", slog.Any("error", err))
	}
}

func (handler *AnalyticsHandler) GetIntegrityCheck(responseWriter http.ResponseWriter, request *http.Request) {
	violations, err := handler.analytics.GetIntegrityViolations(request.Context())
	if err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to get integrity violations", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, violations); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write integrity check response", slog.Any("error", err))
	}
}
