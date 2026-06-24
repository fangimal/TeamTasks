package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockTeamMemberRepository struct {
	mock.Mock
}

func (m *MockTeamMemberRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTeamMemberRepository) Create(ctx context.Context, member *domain.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockTeamMemberRepository) GetByUserAndTeam(ctx context.Context, userID int64, teamID int64) (*domain.TeamMember, error) {
	args := m.Called(ctx, userID, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TeamMember), args.Error(1)
}

func (m *MockTeamMemberRepository) GetMembersByTeam(ctx context.Context, teamID int64) ([]*domain.TeamMember, error) {
	args := m.Called(ctx, teamID)
	members, _ := args.Get(0).([]*domain.TeamMember)
	return members, args.Error(1)
}

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, tx *sql.Tx, task *domain.Task) error {
	args := m.Called(ctx, tx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetList(ctx context.Context, filter domain.TaskFilter, pagination domain.Pagination) ([]*domain.Task, int64, error) {
	args := m.Called(ctx, filter, pagination)
	tasks, _ := args.Get(0).([]*domain.Task)
	total, _ := args.Get(1).(int64)
	return tasks, total, args.Error(2)
}

type MockTaskHistoryRepository struct {
	mock.Mock
}

func (m *MockTaskHistoryRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTaskHistoryRepository) CreateBatch(ctx context.Context, tx *sql.Tx, records []*domain.TaskHistory) error {
	args := m.Called(ctx, tx, records)
	return args.Error(0)
}

func (m *MockTaskHistoryRepository) GetByTaskID(ctx context.Context, taskID int64, limit, offset int) ([]*domain.TaskHistory, error) {
	args := m.Called(ctx, taskID, limit, offset)
	records, _ := args.Get(0).([]*domain.TaskHistory)
	return records, args.Error(1)
}

type MockTaskCacheRepository struct {
	mock.Mock
}

func (m *MockTaskCacheRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTaskCacheRepository) Get(ctx context.Context, key string) ([]*domain.Task, int64, error) {
	args := m.Called(ctx, key)
	tasks, _ := args.Get(0).([]*domain.Task)
	total, _ := args.Get(1).(int64)
	return tasks, total, args.Error(2)
}

func (m *MockTaskCacheRepository) Set(ctx context.Context, key string, tasks []*domain.Task, total int64, ttl time.Duration) error {
	args := m.Called(ctx, key, tasks, total, ttl)
	return args.Error(0)
}

func (m *MockTaskCacheRepository) Invalidate(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendInviteEmail(ctx context.Context, email string, teamName string) error {
	args := m.Called(ctx, email, teamName)
	return args.Error(0)
}
