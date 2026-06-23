package domain

import (
	"context"
	"database/sql"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type UserRepository interface {
	HealthRepository
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type TeamRepository interface {
	HealthRepository
	CreateTeamWithOwner(ctx context.Context, team *Team, member *TeamMember) error
	GetByID(ctx context.Context, id int64) (*Team, error)
	GetByUserID(ctx context.Context, userID int64, limit int, offset int) ([]*Team, error)
}

type TeamMemberRepository interface {
	HealthRepository
	Create(ctx context.Context, member *TeamMember) error
	GetByUserAndTeam(ctx context.Context, userID int64, teamID int64) (*TeamMember, error)
	GetMembersByTeam(ctx context.Context, teamID int64) ([]*TeamMember, error)
}

type TaskRepository interface {
	HealthRepository
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id int64) (*Task, error)
	Update(ctx context.Context, tx *sql.Tx, task *Task) error
	GetList(ctx context.Context, filter TaskFilter, pagination Pagination) ([]*Task, int64, error)
}

type TaskHistoryRepository interface {
	HealthRepository
	CreateBatch(ctx context.Context, tx *sql.Tx, records []*TaskHistory) error
	GetByTaskID(ctx context.Context, taskID int64, limit, offset int) ([]*TaskHistory, error)
}

type TaskCommentRepository interface {
	HealthRepository
	Create(ctx context.Context, comment *TaskComment) error
	GetByTaskID(ctx context.Context, taskID int64, limit, offset int) ([]*TaskComment, error)
	CountByTaskID(ctx context.Context, taskID int64) (int64, error)
}
