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

type TaskUseCase interface {
	CreateTask(ctx context.Context, userID int64, input usecase.CreateTaskInput) (*usecase.TaskResponse, error)
	GetTaskByID(ctx context.Context, userID int64, taskID int64) (*usecase.TaskResponse, error)
	GetTasks(ctx context.Context, userID int64, filter domain.TaskFilter, pagination domain.Pagination) (*usecase.TaskListResponse, error)
	UpdateTask(ctx context.Context, userID int64, taskID int64, input usecase.UpdateTaskInput) (*usecase.TaskResponse, error)
}

type TaskHandler struct {
	logger *slog.Logger
	tasks  TaskUseCase
}

type createTaskRequest struct {
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Status      domain.TaskStatus `json:"status,omitempty"`
	AssigneeID  int64            `json:"assignee_id"`
	TeamID      int64            `json:"team_id"`
}

type updateTaskRequest struct {
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
	Status      domain.TaskStatus `json:"status,omitempty"`
	AssigneeID  int64            `json:"assignee_id,omitempty"`
}

func NewTaskHandler(logger *slog.Logger, tasks TaskUseCase) *TaskHandler {
	return &TaskHandler{
		logger: logger,
		tasks:  tasks,
	}
}

func (handler *TaskHandler) CreateTask(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload createTaskRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	if payload.Title == "" {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "title is required")
		return
	}

	if payload.TeamID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "team_id is required")
		return
	}

	if payload.AssigneeID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "assignee_id is required")
		return
	}

	task, err := handler.tasks.CreateTask(request.Context(), userID, usecase.CreateTaskInput{
		Title:       payload.Title,
		Description: payload.Description,
		Status:      payload.Status,
		AssigneeID:  payload.AssigneeID,
		TeamID:      payload.TeamID,
	})
	if err != nil {
		handler.writeTaskError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusCreated, task); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write create task response", slog.Any("error", err))
	}
}

func (handler *TaskHandler) GetTaskByID(responseWriter http.ResponseWriter, request *http.Request) {
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

	task, err := handler.tasks.GetTaskByID(request.Context(), userID, taskID)
	if err != nil {
		handler.writeTaskError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, task); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write get task response", slog.Any("error", err))
	}
}

func (handler *TaskHandler) GetTasks(responseWriter http.ResponseWriter, request *http.Request) {
	userID, ok := middleware.UserIDFromContext(request.Context())
	if !ok {
		response.WriteError(handler.logger, responseWriter, request, http.StatusUnauthorized, "unauthorized")
		return
	}

	teamID, err := strconv.ParseInt(request.URL.Query().Get("team_id"), 10, 64)
	if err != nil || teamID <= 0 {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "team_id is required")
		return
	}

	limit, offset := parsePagination(request)

	filter := domain.TaskFilter{
		TeamID: teamID,
	}

	if statusParam := request.URL.Query().Get("status"); statusParam != "" {
		filter.Status = domain.TaskStatus(statusParam)
	}

	if assigneeParam := request.URL.Query().Get("assignee_id"); assigneeParam != "" {
		if parsed, err := strconv.ParseInt(assigneeParam, 10, 64); err == nil && parsed > 0 {
			filter.AssigneeID = parsed
		}
	}

	result, err := handler.tasks.GetTasks(request.Context(), userID, filter, domain.Pagination{Limit: limit, Offset: offset})
	if err != nil {
		handler.writeTaskError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, result); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write get tasks response", slog.Any("error", err))
	}
}

func (handler *TaskHandler) UpdateTask(responseWriter http.ResponseWriter, request *http.Request) {
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

	var payload updateTaskRequest
	if err = json.NewDecoder(request.Body).Decode(&payload); err != nil {
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid json body")
		return
	}

	task, err := handler.tasks.UpdateTask(request.Context(), userID, taskID, usecase.UpdateTaskInput{
		Title:       payload.Title,
		Description: payload.Description,
		Status:      payload.Status,
		AssigneeID:  payload.AssigneeID,
	})
	if err != nil {
		handler.writeTaskError(responseWriter, request, err)
		return
	}

	if err = response.JSON(responseWriter, http.StatusOK, task); err != nil {
		handler.logger.ErrorContext(request.Context(), "failed to write update task response", slog.Any("error", err))
	}
}

func (handler *TaskHandler) writeTaskError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "invalid input")
	case errors.Is(err, domain.ErrTaskNotFound):
		response.WriteError(handler.logger, responseWriter, request, http.StatusNotFound, "task not found")
	case errors.Is(err, domain.ErrForbidden):
		response.WriteError(handler.logger, responseWriter, request, http.StatusForbidden, "insufficient permissions")
	case errors.Is(err, domain.ErrNotTeamMember):
		response.WriteError(handler.logger, responseWriter, request, http.StatusForbidden, "you are not a member of this team")
	case errors.Is(err, domain.ErrAssigneeNotInTeam):
		response.WriteError(handler.logger, responseWriter, request, http.StatusBadRequest, "assignee is not a member of this team")
	default:
		handler.logger.ErrorContext(request.Context(), "task request failed", slog.Any("error", err))
		response.WriteError(handler.logger, responseWriter, request, http.StatusInternalServerError, "internal server error")
	}
}
