package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fangimal/TeamTask/internal/domain"
)

const (
	defaultPageLimit  = 10
	defaultPageOffset = 0
)

type NotificationService interface {
	SendInviteEmail(ctx context.Context, email string, teamName string) error
}

type TeamUseCase struct {
	teams    domain.TeamRepository
	members  domain.TeamMemberRepository
	users    domain.UserRepository
	notifier NotificationService
}

type teamMemberResponse struct {
	UserID int64           `json:"user_id"`
	Role   domain.TeamRole `json:"role"`
}

type TeamResponse struct {
	ID        int64                `json:"id"`
	Name      string               `json:"name"`
	CreatedBy int64                `json:"created_by"`
	Members   []teamMemberResponse `json:"members,omitempty"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
}

type CreateTeamInput struct {
	Name string
}

type InviteUserInput struct {
	TeamID int64
	Email  string
	Role   domain.TeamRole
}

func NewTeamUseCase(
	teams domain.TeamRepository,
	members domain.TeamMemberRepository,
	users domain.UserRepository,
) *TeamUseCase {
	return &TeamUseCase{
		teams:    teams,
		members:  members,
		users:    users,
		notifier: &consoleNotifier{},
	}
}

func (useCase *TeamUseCase) CreateTeam(ctx context.Context, userID int64, input CreateTeamInput) (*TeamResponse, error) {
	team := &domain.Team{
		Name:      input.Name,
		CreatedBy: userID,
	}

	member := &domain.TeamMember{}

	if err := useCase.teams.CreateTeamWithOwner(ctx, team, member); err != nil {
		return nil, fmt.Errorf("create team: %w", err)
	}

	return teamToResponse(team, nil), nil
}

func (useCase *TeamUseCase) GetUserTeams(ctx context.Context, userID int64, limit int, offset int) ([]*TeamResponse, error) {
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if offset < 0 {
		offset = defaultPageOffset
	}

	teams, err := useCase.teams.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get user teams: %w", err)
	}

	result := make([]*TeamResponse, 0, len(teams))
	for _, team := range teams {
		result = append(result, teamToResponse(team, nil))
	}

	return result, nil
}

func (useCase *TeamUseCase) InviteUser(ctx context.Context, inviterID int64, input InviteUserInput) error {
	if _, err := useCase.teams.GetByID(ctx, input.TeamID); err != nil {
		if errors.Is(err, domain.ErrTeamNotFound) {
			return domain.ErrTeamNotFound
		}

		return fmt.Errorf("get team: %w", err)
	}

	inviterMember, err := useCase.members.GetByUserAndTeam(ctx, inviterID, input.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrForbidden
		}

		return fmt.Errorf("get inviter membership: %w", err)
	}

	if inviterMember.Role == domain.TeamRoleMember {
		return domain.ErrForbidden
	}

	invitedUser, err := useCase.users.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrUserNotFound
		}

		return fmt.Errorf("get invited user: %w", err)
	}

	existingMember, err := useCase.members.GetByUserAndTeam(ctx, invitedUser.ID, input.TeamID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check existing membership: %w", err)
	}
	if existingMember != nil {
		return domain.ErrUserAlreadyInTeam
	}

	role := domain.TeamRoleMember
	if input.Role != "" {
		if input.Role == domain.TeamRoleOwner {
			return domain.ErrCannotInviteOwner
		}
		if input.Role != domain.TeamRoleAdmin && input.Role != domain.TeamRoleMember {
			return domain.ErrInvalidRole
		}
		role = input.Role
	}

	member := &domain.TeamMember{
		UserID: invitedUser.ID,
		TeamID: input.TeamID,
		Role:   role,
	}

	if err = useCase.members.Create(ctx, member); err != nil {
		return err
	}

	if err = useCase.notifier.SendInviteEmail(ctx, input.Email, "dummy team name"); err != nil {
		slog.WarnContext(ctx, "failed to send invite email", slog.Any("error", err))
	}

	return nil
}

func teamToResponse(team *domain.Team, members []*domain.TeamMember) *TeamResponse {
	response := &TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		CreatedBy: team.CreatedBy,
		CreatedAt: team.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: team.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if members != nil {
		response.Members = make([]teamMemberResponse, 0, len(members))
		for _, member := range members {
			response.Members = append(response.Members, teamMemberResponse{
				UserID: member.UserID,
				Role:   member.Role,
			})
		}
	}

	return response
}

type consoleNotifier struct{}

func (notifier *consoleNotifier) SendInviteEmail(ctx context.Context, email string, teamName string) error {
	slog.InfoContext(ctx, "invite email sent",
		slog.String("email", email),
		slog.String("team_name", teamName),
	)

	return nil
}
