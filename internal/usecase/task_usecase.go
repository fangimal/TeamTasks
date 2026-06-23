package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
)

const cacheTTL = 5 * time.Minute

const (
	defaultTaskLimit  = 10
	defaultTaskOffset = 0
)

type TaskUseCase struct {
	tasks    domain.TaskRepository
	members  domain.TeamMemberRepository
	history  domain.TaskHistoryRepository
	cache    domain.TaskCacheRepository
	database *sql.DB
	logger   *slog.Logger
}

type CreateTaskInput struct {
	Title       string
	Description string
	Status      domain.TaskStatus
	AssigneeID  int64
	TeamID      int64
}

type UpdateTaskInput struct {
	Title       string
	Description string
	Status      domain.TaskStatus
	AssigneeID  int64
}

type TaskResponse struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      domain.TaskStatus `json:"status"`
	AssigneeID  int64             `json:"assignee_id"`
	TeamID      int64             `json:"team_id"`
	CreatedBy   int64             `json:"created_by"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

type TaskListResponse struct {
	Data   []*TaskResponse `json:"data"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type TaskHistoryResponse struct {
	ID        int64           `json:"id"`
	TaskID    int64           `json:"task_id"`
	ChangedBy int64           `json:"changed_by"`
	ChangedAt string          `json:"changed_at"`
	OldValue  json.RawMessage `json:"old_value"`
	NewValue  json.RawMessage `json:"new_value"`
}

type TaskHistoryListResponse struct {
	Data   []*TaskHistoryResponse `json:"data"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

func NewTaskUseCase(tasks domain.TaskRepository, members domain.TeamMemberRepository, history domain.TaskHistoryRepository, cache domain.TaskCacheRepository, database *sql.DB, logger *slog.Logger) *TaskUseCase {
	return &TaskUseCase{
		tasks:    tasks,
		members:  members,
		history:  history,
		cache:    cache,
		database: database,
		logger:   logger,
	}
}

func (useCase *TaskUseCase) CreateTask(ctx context.Context, userID int64, input CreateTaskInput) (*TaskResponse, error) {
	member, err := useCase.members.GetByUserAndTeam(ctx, userID, input.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotTeamMember
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}
	_ = member

	if input.AssigneeID > 0 {
		_, err = useCase.members.GetByUserAndTeam(ctx, input.AssigneeID, input.TeamID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrAssigneeNotInTeam
			}

			return nil, fmt.Errorf("check assignee membership: %w", err)
		}
	}

	if input.Status == "" {
		input.Status = domain.TaskStatusTodo
	}

	task := &domain.Task{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
		AssigneeID:  input.AssigneeID,
		TeamID:      input.TeamID,
		CreatedBy:   userID,
	}

	if err = useCase.tasks.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	if useCase.cache != nil {
		cachePattern := fmt.Sprintf("tasks:team:%d:*", input.TeamID)
		if err = useCase.cache.Invalidate(ctx, cachePattern); err != nil {
			useCase.logger.WarnContext(ctx, "cache invalidation failed after create task",
				slog.Any("error", err),
				slog.Int64("team_id", input.TeamID),
			)
		}
	}

	return taskToResponse(task), nil
}

func (useCase *TaskUseCase) GetTaskByID(ctx context.Context, userID int64, taskID int64) (*TaskResponse, error) {
	task, err := useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	_, err = useCase.members.GetByUserAndTeam(ctx, userID, task.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrForbidden
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	return taskToResponse(task), nil
}

func (useCase *TaskUseCase) GetTasks(ctx context.Context, userID int64, filter domain.TaskFilter, pagination domain.Pagination) (*TaskListResponse, error) {
	if pagination.Limit <= 0 {
		pagination.Limit = defaultTaskLimit
	}
	if pagination.Offset < 0 {
		pagination.Offset = defaultTaskOffset
	}

	_, err := useCase.members.GetByUserAndTeam(ctx, userID, filter.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotTeamMember
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	if useCase.cache != nil {
		cacheKey := buildCacheKey(filter, pagination)
		cachedTasks, total, cacheErr := useCase.cache.Get(ctx, cacheKey)
		if cacheErr == nil {
			response := &TaskListResponse{
				Data:   make([]*TaskResponse, 0, len(cachedTasks)),
				Total:  total,
				Limit:  pagination.Limit,
				Offset: pagination.Offset,
			}

			for _, task := range cachedTasks {
				response.Data = append(response.Data, taskToResponse(task))
			}

			return response, nil
		}

		if !errors.Is(cacheErr, domain.ErrCacheMiss) {
			useCase.logger.WarnContext(ctx, "cache read failed", slog.Any("error", cacheErr))
		}
	}

	tasks, total, err := useCase.tasks.GetList(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("get task list: %w", err)
	}

	response := &TaskListResponse{
		Data:   make([]*TaskResponse, 0, len(tasks)),
		Total:  total,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	}

	for _, task := range tasks {
		response.Data = append(response.Data, taskToResponse(task))
	}

	if useCase.cache != nil {
		cacheKey := buildCacheKey(filter, pagination)
		if err = useCase.cache.Set(ctx, cacheKey, tasks, total, cacheTTL); err != nil {
			useCase.logger.WarnContext(ctx, "cache write failed", slog.Any("error", err))
		}
	}

	return response, nil
}

func buildCacheKey(filter domain.TaskFilter, pagination domain.Pagination) string {
	key := fmt.Sprintf("tasks:team:%d", filter.TeamID)

	if filter.Status != "" {
		key += fmt.Sprintf(":status:%s", filter.Status)
	}

	if filter.AssigneeID > 0 {
		key += fmt.Sprintf(":assignee:%d", filter.AssigneeID)
	}

	key += fmt.Sprintf(":l:%d:o:%d", pagination.Limit, pagination.Offset)

	return key
}

func (useCase *TaskUseCase) UpdateTask(ctx context.Context, userID int64, taskID int64, input UpdateTaskInput) (*TaskResponse, error) {
	oldTask, err := useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	member, err := useCase.members.GetByUserAndTeam(ctx, userID, oldTask.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrForbidden
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	if member.Role == domain.TeamRoleMember {
		if oldTask.AssigneeID != userID && oldTask.CreatedBy != userID {
			return nil, domain.ErrForbidden
		}
	}

	newTask := *oldTask

	if input.Title != "" {
		newTask.Title = input.Title
	}
	if input.Description != "" {
		newTask.Description = input.Description
	}
	if input.Status != "" {
		newTask.Status = input.Status
	}
	if input.AssigneeID > 0 {
		if input.AssigneeID != oldTask.AssigneeID {
			_, err = useCase.members.GetByUserAndTeam(ctx, input.AssigneeID, oldTask.TeamID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, domain.ErrAssigneeNotInTeam
				}

				return nil, fmt.Errorf("check new assignee: %w", err)
			}
		}

		newTask.AssigneeID = input.AssigneeID
	}

	records := buildHistoryRecords(oldTask, &newTask, userID)
	if records == nil {
		return taskToResponse(oldTask), nil
	}

	tx, err := useCase.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err = useCase.tasks.Update(ctx, tx, &newTask); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	if err = useCase.history.CreateBatch(ctx, tx, records); err != nil {
		return nil, fmt.Errorf("create task history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	if useCase.cache != nil {
		cachePattern := fmt.Sprintf("tasks:team:%d:*", oldTask.TeamID)
		if err = useCase.cache.Invalidate(ctx, cachePattern); err != nil {
			useCase.logger.WarnContext(ctx, "cache invalidation failed after update task",
				slog.Any("error", err),
				slog.Int64("team_id", oldTask.TeamID),
			)
		}
	}

	task, err := useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("reload task: %w", err)
	}

	return taskToResponse(task), nil
}

func (useCase *TaskUseCase) GetTaskHistory(ctx context.Context, userID int64, taskID int64, limit int, offset int) (*TaskHistoryListResponse, error) {
	if limit <= 0 {
		limit = defaultTaskLimit
	}
	if offset < 0 {
		offset = defaultTaskOffset
	}

	task, err := useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	_, err = useCase.members.GetByUserAndTeam(ctx, userID, task.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrForbidden
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	records, err := useCase.history.GetByTaskID(ctx, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get task history: %w", err)
	}

	response := &TaskHistoryListResponse{
		Data:   make([]*TaskHistoryResponse, 0, len(records)),
		Limit:  limit,
		Offset: offset,
	}

	for _, record := range records {
		response.Data = append(response.Data, &TaskHistoryResponse{
			ID:        record.ID,
			TaskID:    record.TaskID,
			ChangedBy: record.ChangedBy,
			ChangedAt: record.ChangedAt.Format("2006-01-02T15:04:05Z07:00"),
			OldValue:  record.OldValue,
			NewValue:  record.NewValue,
		})
	}

	return response, nil
}

func buildHistoryRecords(oldTask, newTask *domain.Task, changedBy int64) []*domain.TaskHistory {
	oldValues := make(map[string]any)
	newValues := make(map[string]any)

	if oldTask.Title != newTask.Title {
		oldValues["title"] = oldTask.Title
		newValues["title"] = newTask.Title
	}

	if oldTask.Description != newTask.Description {
		oldValues["description"] = oldTask.Description
		newValues["description"] = newTask.Description
	}

	if oldTask.Status != newTask.Status {
		oldValues["status"] = oldTask.Status
		newValues["status"] = newTask.Status
	}

	if oldTask.AssigneeID != newTask.AssigneeID {
		oldValues["assignee_id"] = oldTask.AssigneeID
		newValues["assignee_id"] = newTask.AssigneeID
	}

	if len(oldValues) == 0 {
		return nil
	}

	oldJSON, err := json.Marshal(oldValues)
	if err != nil {
		return nil
	}

	newJSON, err := json.Marshal(newValues)
	if err != nil {
		return nil
	}

	return []*domain.TaskHistory{
		{
			TaskID:    newTask.ID,
			ChangedBy: changedBy,
			ChangedAt: time.Now(),
			OldValue:  oldJSON,
			NewValue:  newJSON,
		},
	}
}

func taskToResponse(task *domain.Task) *TaskResponse {
	return &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		AssigneeID:  task.AssigneeID,
		TeamID:      task.TeamID,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
