package httpdelivery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/fangimal/TeamTask/internal/usecase"
	"github.com/fangimal/TeamTask/pkg/response"
)

type AuthUseCase interface {
	Register(ctx context.Context, input usecase.RegisterInput) (*usecase.AuthUser, error)
	Login(ctx context.Context, input usecase.LoginInput) (*usecase.LoginResult, error)
}

type AuthHandler struct {
	logger *slog.Logger
	auth   AuthUseCase
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	User usecase.AuthUser `json:"user"`
}

func NewAuthHandler(logger *slog.Logger, auth AuthUseCase) *AuthHandler {
	return &AuthHandler{
		logger: logger,
		auth:   auth,
	}
}

func (handler *AuthHandler) Register(responseWriter http.ResponseWriter, request *http.Request) {
	var payload authRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	user, err := handler.auth.Register(request.Context(), usecase.RegisterInput{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		handler.writeAuthError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusCreated, registerResponse{User: *user}); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write register response", slog.Any("error", err))
	}
}

func (handler *AuthHandler) Login(responseWriter http.ResponseWriter, request *http.Request) {
	var payload authRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	result, err := handler.auth.Login(request.Context(), usecase.LoginInput{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		handler.writeAuthError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, result); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write login response", slog.Any("error", err))
	}
}

func (handler *AuthHandler) writeAuthError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid email or password")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		response.WriteError(handler.logger, responseWriter, request, http.StatusConflict, "user already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "invalid email or password")
	default:
		handler.logger.ErrorContext(request.Context(), "auth request failed", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
	}
}
