package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
)

const (
	defaultCommentLimit  = 10
	defaultCommentOffset = 0
)

type CommentUseCase struct {
	comments domain.TaskCommentRepository
	tasks    domain.TaskRepository
	members  domain.TeamMemberRepository
}

type CommentResponse struct {
	ID        int64  `json:"id"`
	TaskID    int64  `json:"task_id"`
	UserID    int64  `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CommentListResponse struct {
	Data   []*CommentResponse `json:"data"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

func NewCommentUseCase(
	comments domain.TaskCommentRepository,
	tasks domain.TaskRepository,
	members domain.TeamMemberRepository,
) *CommentUseCase {
	return &CommentUseCase{
		comments: comments,
		tasks:    tasks,
		members:  members,
	}
}

func (useCase *CommentUseCase) CreateComment(ctx context.Context, userID int64, taskID int64, text string) (*CommentResponse, error) {
	if text == "" {
		return nil, domain.ErrInvalidInput
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
			return nil, domain.ErrNotTeamMember
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	comment := &domain.TaskComment{
		TaskID:    taskID,
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err = useCase.comments.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	return commentToResponse(comment), nil
}

func (useCase *CommentUseCase) GetComments(ctx context.Context, userID int64, taskID int64, limit int, offset int) (*CommentListResponse, error) {
	if limit <= 0 {
		limit = defaultCommentLimit
	}
	if offset < 0 {
		offset = defaultCommentOffset
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
			return nil, domain.ErrNotTeamMember
		}

		return nil, fmt.Errorf("check membership: %w", err)
	}

	comments, err := useCase.comments.GetByTaskID(ctx, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get comments: %w", err)
	}

	total, err := useCase.comments.CountByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("count comments: %w", err)
	}

	response := &CommentListResponse{
		Data:   make([]*CommentResponse, 0, len(comments)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	for _, comment := range comments {
		response.Data = append(response.Data, commentToResponse(comment))
	}

	return response, nil
}

func commentToResponse(comment *domain.TaskComment) *CommentResponse {
	return &CommentResponse{
		ID:        comment.ID,
		TaskID:    comment.TaskID,
		UserID:    comment.UserID,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
