package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

const (
	defaultTaskLimit  = 10
	defaultTaskOffset = 0
)

type TaskUseCase struct {
	tasks  domain.TaskRepository
	members domain.TeamMemberRepository
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
	ID          int64              `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      domain.TaskStatus  `json:"status"`
	AssigneeID  int64              `json:"assignee_id"`
	TeamID      int64              `json:"team_id"`
	CreatedBy   int64              `json:"created_by"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type TaskListResponse struct {
	Data   []*TaskResponse `json:"data"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

func NewTaskUseCase(tasks domain.TaskRepository, members domain.TeamMemberRepository) *TaskUseCase {
	return &TaskUseCase{
		tasks:  tasks,
		members: members,
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

	return response, nil
}

func (useCase *TaskUseCase) UpdateTask(ctx context.Context, userID int64, taskID int64, input UpdateTaskInput) (*TaskResponse, error) {
	task, err := useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			return nil, domain.ErrTaskNotFound
		}

		return nil, fmt.Errorf("get task: %w", err)
	}

	member, err := useCase.members.GetByUserAndTeam(ctx, userID, task.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrForbidden
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	if member.Role == domain.TeamRoleMember {
		if task.AssigneeID != userID && task.CreatedBy != userID {
			return nil, domain.ErrForbidden
		}
	}

	if input.Title != "" {
		task.Title = input.Title
	}
	if input.Description != "" {
		task.Description = input.Description
	}
	if input.Status != "" {
		task.Status = input.Status
	}
	if input.AssigneeID > 0 {
		if input.AssigneeID != task.AssigneeID {
			_, err = useCase.members.GetByUserAndTeam(ctx, input.AssigneeID, task.TeamID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, domain.ErrAssigneeNotInTeam
				}

				return nil, fmt.Errorf("check new assignee: %w", err)
			}
		}

		task.AssigneeID = input.AssigneeID
	}

	if err = useCase.tasks.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}

	task, err = useCase.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("reload task: %w", err)
	}

	return taskToResponse(task), nil
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
