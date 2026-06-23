package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

type AnalyticsRepository struct {
	database *sql.DB
}

func NewAnalyticsRepository(database *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{
		database: database,
	}
}

func (repository *AnalyticsRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *AnalyticsRepository) GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error) {
	query := `
		SELECT
			t.id,
			t.name,
			COUNT(DISTINCT tm.user_id) AS member_count,
			COUNT(DISTINCT CASE WHEN tk.status = 'done' AND tk.created_at >= NOW() - INTERVAL 7 DAY THEN tk.id END) AS done_tasks_last_7_days
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name
		ORDER BY t.id`

	rows, err := repository.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query team stats: %w", err)
	}
	defer rows.Close()

	var stats []*domain.TeamStats

	for rows.Next() {
		stat := &domain.TeamStats{}

		if err = rows.Scan(&stat.TeamID, &stat.TeamName, &stat.MemberCount, &stat.DoneTasksLast7Days); err != nil {
			return nil, fmt.Errorf("scan team stats: %w", err)
		}

		stats = append(stats, stat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team stats: %w", err)
	}

	return stats, nil
}

func (repository *AnalyticsRepository) GetTopUsersPerTeam(ctx context.Context) ([]*domain.TopUser, error) {
	query := `
		WITH user_task_counts AS (
			SELECT
				t.team_id,
				t.created_by AS user_id,
				COUNT(*) AS task_count
			FROM tasks t
			GROUP BY t.team_id, t.created_by
		),
		ranked_users AS (
			SELECT
				utc.team_id,
				utc.user_id,
				u.email AS user_email,
				utc.task_count,
				DENSE_RANK() OVER (PARTITION BY utc.team_id ORDER BY utc.task_count DESC) AS user_rank
			FROM user_task_counts utc
			JOIN users u ON u.id = utc.user_id
		)
		SELECT team_id, user_id, user_email, task_count, user_rank
		FROM ranked_users
		WHERE user_rank <= 3
		ORDER BY team_id, user_rank`

	rows, err := repository.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query top users: %w", err)
	}
	defer rows.Close()

	var topUsers []*domain.TopUser

	for rows.Next() {
		user := &domain.TopUser{}

		if err = rows.Scan(&user.TeamID, &user.UserID, &user.UserEmail, &user.TaskCount, &user.Rank); err != nil {
			return nil, fmt.Errorf("scan top user: %w", err)
		}

		topUsers = append(topUsers, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top users: %w", err)
	}

	return topUsers, nil
}

func (repository *AnalyticsRepository) GetIntegrityViolations(ctx context.Context) ([]*domain.IntegrityViolation, error) {
	query := `
		SELECT
			tk.id,
			tk.title,
			tk.assignee_id,
			tk.team_id
		FROM tasks tk
		LEFT JOIN team_members tm ON tm.user_id = tk.assignee_id AND tm.team_id = tk.team_id
		WHERE tm.user_id IS NULL`

	rows, err := repository.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query integrity violations: %w", err)
	}
	defer rows.Close()

	var violations []*domain.IntegrityViolation

	for rows.Next() {
		violation := &domain.IntegrityViolation{}

		if err = rows.Scan(&violation.TaskID, &violation.TaskTitle, &violation.AssigneeID, &violation.TeamID); err != nil {
			return nil, fmt.Errorf("scan integrity violation: %w", err)
		}

		violations = append(violations, violation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate integrity violations: %w", err)
	}

	return violations, nil
}
