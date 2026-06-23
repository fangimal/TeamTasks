package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

type TaskCommentRepository struct {
	database *sql.DB
}

func NewTaskCommentRepository(database *sql.DB) *TaskCommentRepository {
	return &TaskCommentRepository{
		database: database,
	}
}

func (repository *TaskCommentRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *TaskCommentRepository) Create(ctx context.Context, comment *domain.TaskComment) error {
	result, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO task_comments (task_id, user_id, text) VALUES (?, ?, ?)`,
		comment.TaskID,
		comment.UserID,
		comment.Text,
	)
	if err != nil {
		return fmt.Errorf("insert task comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted comment id: %w", err)
	}

	comment.ID = id

	return nil
}

func (repository *TaskCommentRepository) GetByTaskID(ctx context.Context, taskID int64, limit, offset int) ([]*domain.TaskComment, error) {
	rows, err := repository.database.QueryContext(
		ctx,
		`SELECT id, task_id, user_id, text, created_at, updated_at
		FROM task_comments
		WHERE task_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		taskID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("select task comments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.TaskComment

	for rows.Next() {
		comment := &domain.TaskComment{}

		if err = rows.Scan(&comment.ID, &comment.TaskID, &comment.UserID,
			&comment.Text, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan task comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}

func (repository *TaskCommentRepository) CountByTaskID(ctx context.Context, taskID int64) (int64, error) {
	var count int64

	err := repository.database.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM task_comments WHERE task_id = ?",
		taskID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count task comments: %w", err)
	}

	return count, nil
}
