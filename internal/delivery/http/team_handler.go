package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/fangimal/TeamTask/internal/delivery/http/middleware"
	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/fangimal/TeamTask/internal/usecase"
	"github.com/fangimal/TeamTask/pkg/response"
)

type TeamUseCase interface {
	CreateTeam(ctx context.Context, userID int64, input usecase.CreateTeamInput) (*usecase.TeamResponse, error)
	GetUserTeams(ctx context.Context, userID int64, limit int, offset int) ([]*usecase.TeamResponse, error)
	InviteUser(ctx context.Context, inviterID int64, input usecase.InviteUserInput) error
}

type TeamHandler struct {
	logger *slog.Logger
	teams  TeamUseCase
}

type createTeamRequest struct {
	Name string `json:"name"`
}

type inviteUserRequest struct {
	Email string          `json:"email"`
	Role  domain.TeamRole `json:"role,omitempty"`
}

func NewTeamHandler(logger *slog.Logger, teams TeamUseCase) *TeamHandler {
	return &TeamHandler{
		logger: logger,
		teams:  teams,
	}
}

func (handler *TeamHandler) CreateTeam(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload createTeamRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	if payload.Name == "" {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "team name is required")
		return
	}

	team, err := handler.teams.CreateTeam(request.Context(), userID, usecase.CreateTeamInput{Name: payload.Name})
	if err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to create team", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
		return
	}

	if err = response.JSON(responseWriter, http.StatusCreated, team); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write create team response", slog.Any("error", err))
	}
}

func (handler *TeamHandler) GetUserTeams(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, offset := parsePagination(request)

	teams, err := handler.teams.GetUserTeams(request.Context(), userID, limit, offset)
	if err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to get user teams", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, teams); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write get teams response", slog.Any("error", err))
	}
}

func (handler *TeamHandler) InviteUser(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamID, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || teamID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid team id")
		return
	}

	var payload inviteUserRequest
	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	if payload.Email == "" {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "email is required")
		return
	}

	err = handler.teams.InviteUser(request.Context(), userID, usecase.InviteUserInput{
		TeamID: teamID,
		Email:  payload.Email,
		Role:   payload.Role,
	})
	if err != nil {
		handler.writeInviteError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "user invited"}); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write invite response", slog.Any("error", err))
	}
}

func (handler *TeamHandler) writeInviteError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrTeamNotFound):
		response.WriteError(handler.logger, responseWriter, request, http.StatusNotFound, "team not found")
	case errors.Is(err, domain.ErrUserNotFound):
		response.WriteError(handler.logger, responseWriter, request, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrForbidden):
		response.WriteError(handler.logger, responseWriter, request, http.StatusForbidden, "insufficient permissions")
	case errors.Is(err, domain.ErrUserAlreadyInTeam):
		response.WriteError(handler.logger, responseWriter, request, http.StatusConflict, "user already in team")
	case errors.Is(err, domain.ErrCannotInviteOwner):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "cannot invite user as owner")
	case errors.Is(err, domain.ErrInvalidRole):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid role")
	default:
		handler.logger.ErrorContext(request.Context(), "invite user failed", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
	}
}

func parsePagination(request *http.Request) (int, int) {
	limit := 10
	offset := 0

	if limitParam := request.URL.Query().Get("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetParam := request.URL.Query().Get("offset"); offsetParam != "" {
		if parsed, err := strconv.Atoi(offsetParam); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
