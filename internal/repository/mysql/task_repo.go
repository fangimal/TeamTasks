package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/fangimal/TeamTask/internal/domain"
)

var errTxRequired = errors.New("transaction is required for update")

type TaskRepository struct {
	database *sql.DB
}

func NewTaskRepository(database *sql.DB) *TaskRepository {
	return &TaskRepository{
		database: database,
	}
}

func (repository *TaskRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	result, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO tasks (title, description, status, assignee_id, team_id, created_by)
		VALUES (?, ?, ?, ?, ?, ?)`,
		task.Title,
		task.Description,
		task.Status,
		task.AssigneeID,
		task.TeamID,
		task.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted task id: %w", err)
	}

	task.ID = id

	return nil
}

func (repository *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	task := &domain.Task{}

	err := repository.database.QueryRowContext(
		ctx,
		`SELECT id, title, description, status, assignee_id, team_id, created_by, created_at, updated_at
		FROM tasks WHERE id = ?`,
		id,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status,
		&task.AssigneeID, &task.TeamID, &task.CreatedBy,
		&task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTaskNotFound
		}

		return nil, fmt.Errorf("select task by id: %w", err)
	}

	return task, nil
}

func (repository *TaskRepository) Update(ctx context.Context, tx *sql.Tx, task *domain.Task) error {
	if tx == nil {
		return errTxRequired
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE tasks SET title = ?, description = ?, status = ?, assignee_id = ?
		WHERE id = ?`,
		task.Title,
		task.Description,
		task.Status,
		task.AssigneeID,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (repository *TaskRepository) GetList(ctx context.Context, filter domain.TaskFilter, pagination domain.Pagination) ([]*domain.Task, int64, error) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	conditions = append(conditions, "team_id = ?")
	args = append(args, filter.TeamID)

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.AssigneeID > 0 {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, filter.AssigneeID)
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(
		`SELECT id, title, description, status, assignee_id, team_id, created_by, created_at, updated_at,
		COUNT(*) OVER() as total_count
		FROM tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		whereClause,
	)

	args = append(args, pagination.Limit, pagination.Offset)

	rows, err := repository.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	var totalCount int64

	for rows.Next() {
		task := &domain.Task{}

		if err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status,
			&task.AssigneeID, &task.TeamID, &task.CreatedBy,
			&task.CreatedAt, &task.UpdatedAt, &totalCount); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, totalCount, nil
}
