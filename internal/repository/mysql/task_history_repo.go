package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

type TaskHistoryRepository struct {
	database *sql.DB
}

func NewTaskHistoryRepository(database *sql.DB) *TaskHistoryRepository {
	return &TaskHistoryRepository{
		database: database,
	}
}

func (repository *TaskHistoryRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *TaskHistoryRepository) CreateBatch(ctx context.Context, tx *sql.Tx, records []*domain.TaskHistory) error {
	query := `INSERT INTO task_history (task_id, changed_by, old_value, new_value) VALUES (?, ?, ?, ?)`

	for _, record := range records {
		if _, err := tx.ExecContext(
			ctx,
			query,
			record.TaskID,
			record.ChangedBy,
			record.OldValue,
			record.NewValue,
		); err != nil {
			return fmt.Errorf("insert task history: %w", err)
		}
	}

	return nil
}

func (repository *TaskHistoryRepository) GetByTaskID(ctx context.Context, taskID int64, limit int, offset int) ([]*domain.TaskHistory, error) {
	rows, err := repository.database.QueryContext(
		ctx,
		`SELECT id, task_id, changed_by, changed_at, old_value, new_value
		FROM task_history
		WHERE task_id = ?
		ORDER BY changed_at DESC
		LIMIT ? OFFSET ?`,
		taskID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("select task history: %w", err)
	}
	defer rows.Close()

	var records []*domain.TaskHistory

	for rows.Next() {
		record := &domain.TaskHistory{}

		if err = rows.Scan(&record.ID, &record.TaskID, &record.ChangedBy, &record.ChangedAt, &record.OldValue, &record.NewValue); err != nil {
			return nil, fmt.Errorf("scan task history: %w", err)
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task history: %w", err)
	}

	return records, nil
}
