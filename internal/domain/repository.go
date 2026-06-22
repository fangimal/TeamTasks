package domain

import "context"

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type UserRepository interface {
	HealthRepository
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
