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

type CommentUseCase interface {
	CreateComment(ctx context.Context, userID int64, taskID int64, text string) (*usecase.CommentResponse, error)
	GetComments(ctx context.Context, userID int64, taskID int64, limit int, offset int) (*usecase.CommentListResponse, error)
}

type CommentHandler struct {
	logger   *slog.Logger
	comments CommentUseCase
}

type createCommentRequest struct {
	Text string `json:"text"`
}

func NewCommentHandler(logger *slog.Logger, comments CommentUseCase) *CommentHandler {
	return &CommentHandler{
		logger:   logger,
		comments: comments,
	}
}

func (handler *CommentHandler) CreateComment(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid task id")
		return
	}

	var payload createCommentRequest
	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	if payload.Text == "" {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "text is required")
		return
	}

	comment, err := handler.comments.CreateComment(request.Context(), userID, taskID, payload.Text)
	if err != nil {
		handler.writeCommentError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusCreated, comment); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write create comment response", slog.Any("error", err))
	}
}

func (handler *CommentHandler) GetComments(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid task id")
		return
	}

	limit, offset := parsePagination(request)

	result, err := handler.comments.GetComments(request.Context(), userID, taskID, limit, offset)
	if err != nil {
		handler.writeCommentError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, result); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write get comments response", slog.Any("error", err))
	}
}

func (handler *CommentHandler) writeCommentError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid input")
	case errors.Is(err, domain.ErrTaskNotFound):
		response.WriteError(handler.logger, responseWriter, request, http.StatusNotFound, "task not found")
	case errors.Is(err, domain.ErrNotTeamMember):
		response.WriteError(handler.logger, responseWriter, request, http.StatusForbidden, "you are not a member of this team")
	default:
		handler.logger.ErrorContext(request.Context(), "comment request failed", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
	}
}
