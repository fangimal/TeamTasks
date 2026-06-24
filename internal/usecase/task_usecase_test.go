package usecase

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestCreateTask_Success(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(2), int64(10)).
		Return(&domain.TeamMember{UserID: 2, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockTasks.On("Create", mock.Anything, mock.MatchedBy(func(task *domain.Task) bool {
		return task.Title == "Test Task" && task.TeamID == 10
	})).Return(nil).Run(func(args mock.Arguments) {
		task := args.Get(1).(*domain.Task)
		task.ID = 100
	})
	mockCache.On("Invalidate", mock.Anything, "tasks:team:10:*").Return(nil)

	response, err := taskUseCase.CreateTask(context.Background(), 1, CreateTaskInput{
		Title:       "Test Task",
		Description: "Test Description",
		AssigneeID:  2,
		TeamID:      10,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(100), response.ID)
	assert.Equal(t, "Test Task", response.Title)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCreateTask_NotTeamMember(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(nil, sql.ErrNoRows)

	response, err := taskUseCase.CreateTask(context.Background(), 1, CreateTaskInput{
		Title:       "Test Task",
		Description: "Test Description",
		AssigneeID:  2,
		TeamID:      10,
	})

	assert.ErrorIs(t, err, domain.ErrNotTeamMember)
	assert.Nil(t, response)
	mockMembers.AssertExpectations(t)
}

func TestCreateTask_CacheInvalidationError(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockTasks.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		task := args.Get(1).(*domain.Task)
		task.ID = 100
	})
	mockCache.On("Invalidate", mock.Anything, "tasks:team:10:*").Return(assert.AnError)

	response, err := taskUseCase.CreateTask(context.Background(), 1, CreateTaskInput{
		Title:       "Test Task",
		Description: "Test Description",
		AssigneeID:  1,
		TeamID:      10,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(100), response.ID)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCreateTask_AssigneeNotInTeam(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(2), int64(10)).
		Return(nil, sql.ErrNoRows)

	response, err := taskUseCase.CreateTask(context.Background(), 1, CreateTaskInput{
		Title:       "Test Task",
		Description: "Test Description",
		AssigneeID:  2,
		TeamID:      10,
	})

	assert.ErrorIs(t, err, domain.ErrAssigneeNotInTeam)
	assert.Nil(t, response)
	mockMembers.AssertExpectations(t)
}

func TestUpdateTask_MemberCannotUpdateOtherTask(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	existingTask := &domain.Task{
		ID:          100,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  2,
		TeamID:      10,
		CreatedBy:   3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)

	response, err := taskUseCase.UpdateTask(context.Background(), 1, 100, UpdateTaskInput{
		Title: "Updated Title",
	})

	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestUpdateTask_MemberCanUpdateOwnAssignedTask(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:          100,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  1,
		TeamID:      10,
		CreatedBy:   2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)

	// No changes → early return without transaction
	response, err := taskUseCase.UpdateTask(context.Background(), 1, 100, UpdateTaskInput{})

	assert.NoError(t, err)
	assert.Equal(t, "Original Title", response.Title)
	assert.Equal(t, "Original Description", response.Description)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestUpdateTask_MemberCanUpdateOwnCreatedTask(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:          100,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  2,
		TeamID:      10,
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)

	// No changes → early return without transaction
	response, err := taskUseCase.UpdateTask(context.Background(), 1, 100, UpdateTaskInput{})

	assert.NoError(t, err)
	assert.Equal(t, "Original Title", response.Title)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestUpdateTask_TaskNotFound(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockTasks.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.ErrTaskNotFound)

	response, err := taskUseCase.UpdateTask(context.Background(), 1, 999, UpdateTaskInput{
		Title: "Updated Title",
	})

	assert.ErrorIs(t, err, domain.ErrTaskNotFound)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
}

func TestUpdateTask_MemberCannotUpdateWithInvalidAssignee(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:          100,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  1,
		TeamID:      10,
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(999), int64(10)).
		Return(nil, sql.ErrNoRows)

	response, err := taskUseCase.UpdateTask(context.Background(), 1, 100, UpdateTaskInput{
		AssigneeID: 999,
	})

	assert.ErrorIs(t, err, domain.ErrAssigneeNotInTeam)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestUpdateTask_NoChanges(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:          100,
		Title:       "Same Title",
		Description: "Same Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  1,
		TeamID:      10,
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleAdmin}, nil)

	// All input values match existing → no history records → early return
	response, err := taskUseCase.UpdateTask(context.Background(), 1, 100, UpdateTaskInput{
		Title:       "Same Title",
		Description: "Same Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  1,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(100), response.ID)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestGetTaskByID_TaskNotFound(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockTasks.On("GetByID", mock.Anything, int64(999)).Return(nil, domain.ErrTaskNotFound)

	response, err := taskUseCase.GetTaskByID(context.Background(), 1, 999)

	assert.ErrorIs(t, err, domain.ErrTaskNotFound)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
}

func TestGetTaskByID_UnexpectedTaskError(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockTasks.On("GetByID", mock.Anything, int64(999)).Return(nil, assert.AnError)

	response, err := taskUseCase.GetTaskByID(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
}

func TestGetTaskByID_Success(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:          100,
		Title:       "My Task",
		Description: "Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  1,
		TeamID:      10,
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)

	response, err := taskUseCase.GetTaskByID(context.Background(), 1, 100)

	assert.NoError(t, err)
	assert.Equal(t, int64(100), response.ID)
	assert.Equal(t, "My Task", response.Title)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestGetTaskByID_Forbidden(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	now := time.Now()
	existingTask := &domain.Task{
		ID:        100,
		Title:     "My Task",
		TeamID:    10,
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockTasks.On("GetByID", mock.Anything, int64(100)).Return(existingTask, nil)
	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(nil, sql.ErrNoRows)

	response, err := taskUseCase.GetTaskByID(context.Background(), 1, 100)

	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.Nil(t, response)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
}

func TestGetTasks_Success(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockCache.On("Get", mock.Anything, "tasks:team:10:l:10:o:0").
		Return(nil, int64(0), domain.ErrCacheMiss)

	now := time.Now()
	tasks := []*domain.Task{
		{ID: 1, Title: "Task 1", Status: domain.TaskStatusTodo, AssigneeID: 1, TeamID: 10, CreatedBy: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Task 2", Status: domain.TaskStatusDone, AssigneeID: 2, TeamID: 10, CreatedBy: 2, CreatedAt: now, UpdatedAt: now},
	}
	mockTasks.On("GetList", mock.Anything, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0}).
		Return(tasks, int64(2), nil)
	mockCache.On("Set", mock.Anything, "tasks:team:10:l:10:o:0", tasks, int64(2), 5*time.Minute).
		Return(nil)

	response, err := taskUseCase.GetTasks(context.Background(), 1, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0})

	assert.NoError(t, err)
	assert.Len(t, response.Data, 2)
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 0, response.Offset)
	mockTasks.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetTasks_NotTeamMember(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(nil, sql.ErrNoRows)

	response, err := taskUseCase.GetTasks(context.Background(), 1, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0})

	assert.ErrorIs(t, err, domain.ErrNotTeamMember)
	assert.Nil(t, response)
	mockMembers.AssertExpectations(t)
}

func TestGetTasks_CacheReadError_FallsbackToDB(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)
	mockCache.On("Get", mock.Anything, "tasks:team:10:l:10:o:0").
		Return(nil, int64(0), assert.AnError)

	now := time.Now()
	tasks := []*domain.Task{
		{ID: 1, Title: "Task 1", Status: domain.TaskStatusTodo, AssigneeID: 1, TeamID: 10, CreatedBy: 1, CreatedAt: now, UpdatedAt: now},
	}
	mockTasks.On("GetList", mock.Anything, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0}).
		Return(tasks, int64(1), nil)
	mockCache.On("Set", mock.Anything, "tasks:team:10:l:10:o:0", tasks, int64(1), 5*time.Minute).
		Return(nil)

	response, err := taskUseCase.GetTasks(context.Background(), 1, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0})

	assert.NoError(t, err)
	assert.Len(t, response.Data, 1)
	mockTasks.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestGetTasks_FromCache(t *testing.T) {
	mockTasks := new(MockTaskRepository)
	mockMembers := new(MockTeamMemberRepository)
	mockHistory := new(MockTaskHistoryRepository)
	mockCache := new(MockTaskCacheRepository)
	logger := newTestLogger()

	taskUseCase := NewTaskUseCase(mockTasks, mockMembers, mockHistory, mockCache, nil, logger)

	mockMembers.On("GetByUserAndTeam", mock.Anything, int64(1), int64(10)).
		Return(&domain.TeamMember{UserID: 1, TeamID: 10, Role: domain.TeamRoleMember}, nil)

	now := time.Now()
	cachedTasks := []*domain.Task{
		{ID: 1, Title: "Cached Task", Status: domain.TaskStatusTodo, AssigneeID: 1, TeamID: 10, CreatedBy: 1, CreatedAt: now, UpdatedAt: now},
	}
	mockCache.On("Get", mock.Anything, "tasks:team:10:l:10:o:0").
		Return(cachedTasks, int64(1), nil)

	response, err := taskUseCase.GetTasks(context.Background(), 1, domain.TaskFilter{TeamID: 10}, domain.Pagination{Limit: 10, Offset: 0})

	assert.NoError(t, err)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "Cached Task", response.Data[0].Title)
	assert.Equal(t, int64(1), response.Total)
	mockCache.AssertExpectations(t)
	mockMembers.AssertExpectations(t)
	mockTasks.AssertNotCalled(t, "GetList", mock.Anything, mock.Anything, mock.Anything)
}
