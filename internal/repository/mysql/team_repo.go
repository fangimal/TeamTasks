package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

type TeamRepository struct {
	database *sql.DB
}

func NewTeamRepository(database *sql.DB) *TeamRepository {
	return &TeamRepository{
		database: database,
	}
}

func (repository *TeamRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *TeamRepository) CreateTeamWithOwner(ctx context.Context, team *domain.Team, member *domain.TeamMember) error {
	tx, err := repository.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		"INSERT INTO teams (name, created_by) VALUES (?, ?)",
		team.Name,
		team.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	teamID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted team id: %w", err)
	}

	team.ID = teamID
	member.TeamID = teamID
	member.UserID = team.CreatedBy
	member.Role = domain.TeamRoleOwner

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO team_members (user_id, team_id, role) VALUES (?, ?, ?)",
		member.UserID,
		member.TeamID,
		member.Role,
	)
	if err != nil {
		return fmt.Errorf("insert team member: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (repository *TeamRepository) GetByID(ctx context.Context, id int64) (*domain.Team, error) {
	team := &domain.Team{}

	err := repository.database.QueryRowContext(
		ctx,
		"SELECT id, name, created_by, created_at, updated_at FROM teams WHERE id = ?",
		id,
	).Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTeamNotFound
		}

		return nil, fmt.Errorf("select team by id: %w", err)
	}

	return team, nil
}

func (repository *TeamRepository) GetByUserID(ctx context.Context, userID int64, limit int, offset int) ([]*domain.Team, error) {
	rows, err := repository.database.QueryContext(
		ctx,
		`SELECT t.id, t.name, t.created_by, t.created_at, t.updated_at
		FROM teams t
		INNER JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?`,
		userID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("select teams by user id: %w", err)
	}
	defer rows.Close()

	var teams []*domain.Team

	for rows.Next() {
		team := &domain.Team{}

		if err = rows.Scan(&team.ID, &team.Name, &team.CreatedBy, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}

		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams: %w", err)
	}

	return teams, nil
}
