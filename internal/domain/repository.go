package domain

import "context"

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
}

type TaskRepository interface {
	HealthRepository
}

type TaskHistoryRepository interface {
	HealthRepository
}

type TaskCommentRepository interface {
	HealthRepository
}
