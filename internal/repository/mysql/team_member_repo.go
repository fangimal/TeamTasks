package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
)

type TeamMemberRepository struct {
	database *sql.DB
}

func NewTeamMemberRepository(database *sql.DB) *TeamMemberRepository {
	return &TeamMemberRepository{
		database: database,
	}
}

func (repository *TeamMemberRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *TeamMemberRepository) Create(ctx context.Context, member *domain.TeamMember) error {
	_, err := repository.database.ExecContext(
		ctx,
		"INSERT INTO team_members (user_id, team_id, role) VALUES (?, ?, ?)",
		member.UserID,
		member.TeamID,
		member.Role,
	)
	if err != nil {
		if isDuplicateEntryError(err) {
			return domain.ErrUserAlreadyInTeam
		}

		return fmt.Errorf("insert team member: %w", err)
	}

	return nil
}

func (repository *TeamMemberRepository) GetByUserAndTeam(ctx context.Context, userID int64, teamID int64) (*domain.TeamMember, error) {
	member := &domain.TeamMember{}

	err := repository.database.QueryRowContext(
		ctx,
		"SELECT user_id, team_id, role, joined_at FROM team_members WHERE user_id = ? AND team_id = ?",
		userID,
		teamID,
	).Scan(&member.UserID, &member.TeamID, &member.Role, &member.JoinedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}

		return nil, fmt.Errorf("select team member: %w", err)
	}

	return member, nil
}

func (repository *TeamMemberRepository) GetMembersByTeam(ctx context.Context, teamID int64) ([]*domain.TeamMember, error) {
	rows, err := repository.database.QueryContext(
		ctx,
		"SELECT user_id, team_id, role, joined_at FROM team_members WHERE team_id = ?",
		teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("select members by team: %w", err)
	}
	defer rows.Close()

	var members []*domain.TeamMember

	for rows.Next() {
		member := &domain.TeamMember{}

		if err = rows.Scan(&member.UserID, &member.TeamID, &member.Role, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return members, nil
}
