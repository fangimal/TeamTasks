package mysql

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedTestData(t *testing.T, database *sql.DB) (int64, int64, int64) {
	t.Helper()

	result, err := database.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", "owner@test.com", "hash1")
	require.NoError(t, err)
	ownerID, err := result.LastInsertId()
	require.NoError(t, err)

	result, err = database.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", "member@test.com", "hash2")
	require.NoError(t, err)
	memberID, err := result.LastInsertId()
	require.NoError(t, err)

	result, err = database.Exec("INSERT INTO teams (name, created_by) VALUES (?, ?)", "Test Team", ownerID)
	require.NoError(t, err)
	teamID, err := result.LastInsertId()
	require.NoError(t, err)

	_, err = database.Exec("INSERT INTO team_members (user_id, team_id, role) VALUES (?, ?, ?), (?, ?, ?)",
		ownerID, teamID, "owner", memberID, teamID, "member")
	require.NoError(t, err)

	return ownerID, memberID, teamID
}

func TestIntegration_TaskRepository_CreateAndGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tc := setupTestContainer(t)
	defer tc.cleanup()

	taskRepo := NewTaskRepository(tc.database)
	ownerID, _, teamID := seedTestData(t, tc.database)

	task := &domain.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  ownerID,
		TeamID:      teamID,
		CreatedBy:   ownerID,
	}

	err := taskRepo.Create(context.Background(), task)
	require.NoError(t, err)
	assert.Greater(t, task.ID, int64(0))

	loaded, err := taskRepo.GetByID(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, loaded.ID)
	assert.Equal(t, "Test Task", loaded.Title)
	assert.Equal(t, "Test Description", loaded.Description)
	assert.Equal(t, domain.TaskStatusTodo, loaded.Status)
	assert.Equal(t, ownerID, loaded.AssigneeID)
	assert.Equal(t, teamID, loaded.TeamID)
	assert.Equal(t, ownerID, loaded.CreatedBy)
	assert.False(t, loaded.CreatedAt.IsZero())
	assert.False(t, loaded.UpdatedAt.IsZero())
}

func TestIntegration_TaskRepository_GetList_WithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tc := setupTestContainer(t)
	defer tc.cleanup()

	taskRepo := NewTaskRepository(tc.database)
	ownerID, memberID, teamID := seedTestData(t, tc.database)

	tasks := []*domain.Task{
		{Title: "Task 1", Description: "D1", Status: domain.TaskStatusTodo, AssigneeID: ownerID, TeamID: teamID, CreatedBy: ownerID},
		{Title: "Task 2", Description: "D2", Status: domain.TaskStatusInProgress, AssigneeID: memberID, TeamID: teamID, CreatedBy: ownerID},
		{Title: "Task 3", Description: "D3", Status: domain.TaskStatusDone, AssigneeID: ownerID, TeamID: teamID, CreatedBy: ownerID},
		{Title: "Task 4", Description: "D4", Status: domain.TaskStatusTodo, AssigneeID: memberID, TeamID: teamID, CreatedBy: memberID},
	}
	for _, task := range tasks {
		err := taskRepo.Create(context.Background(), task)
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond) // ensure distinct created_at for ordering
	}

	t.Run("all tasks with limit 2", func(t *testing.T) {
		result, total, err := taskRepo.GetList(context.Background(), domain.TaskFilter{TeamID: teamID}, domain.Pagination{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(4), total)
	})

	t.Run("filter by status", func(t *testing.T) {
		result, total, err := taskRepo.GetList(context.Background(), domain.TaskFilter{TeamID: teamID, Status: domain.TaskStatusTodo}, domain.Pagination{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
	})

	t.Run("filter by assignee", func(t *testing.T) {
		result, total, err := taskRepo.GetList(context.Background(), domain.TaskFilter{TeamID: teamID, AssigneeID: memberID}, domain.Pagination{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
	})

	t.Run("filter by status and assignee", func(t *testing.T) {
		result, total, err := taskRepo.GetList(context.Background(), domain.TaskFilter{TeamID: teamID, Status: domain.TaskStatusTodo, AssigneeID: memberID}, domain.Pagination{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), total)
	})

	t.Run("pagination offset", func(t *testing.T) {
		result, total, err := taskRepo.GetList(context.Background(), domain.TaskFilter{TeamID: teamID}, domain.Pagination{Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(4), total)
	})
}

func TestIntegration_TaskRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tc := setupTestContainer(t)
	defer tc.cleanup()

	taskRepo := NewTaskRepository(tc.database)
	ownerID, _, teamID := seedTestData(t, tc.database)

	task := &domain.Task{
		Title:       "Original",
		Description: "Original Desc",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  ownerID,
		TeamID:      teamID,
		CreatedBy:   ownerID,
	}
	err := taskRepo.Create(context.Background(), task)
	require.NoError(t, err)

	tx, err := tc.database.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	task.Title = "Updated Title"
	task.Description = "Updated Desc"
	task.Status = domain.TaskStatusDone
	task.AssigneeID = ownerID

	err = taskRepo.Update(context.Background(), tx, task)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	loaded, err := taskRepo.GetByID(context.Background(), task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", loaded.Title)
	assert.Equal(t, "Updated Desc", loaded.Description)
	assert.Equal(t, domain.TaskStatusDone, loaded.Status)
}

func TestIntegration_TaskHistoryRepository_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tc := setupTestContainer(t)
	defer tc.cleanup()

	taskRepo := NewTaskRepository(tc.database)
	historyRepo := NewTaskHistoryRepository(tc.database)
	ownerID, _, teamID := seedTestData(t, tc.database)

	task := &domain.Task{
		Title:       "Task for History",
		Description: "Desc",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  ownerID,
		TeamID:      teamID,
		CreatedBy:   ownerID,
	}
	err := taskRepo.Create(context.Background(), task)
	require.NoError(t, err)

	tx, err := tc.database.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	records := []*domain.TaskHistory{
		{
			TaskID:    task.ID,
			ChangedBy: ownerID,
			ChangedAt: time.Now(),
			OldValue:  []byte(`{"title":"Old Title"}`),
			NewValue:  []byte(`{"title":"Task for History"}`),
		},
	}

	err = historyRepo.CreateBatch(context.Background(), tx, records)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	loaded, err := historyRepo.GetByTaskID(context.Background(), task.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, task.ID, loaded[0].TaskID)
	assert.Equal(t, ownerID, loaded[0].ChangedBy)
	assert.Equal(t, `{"title":"Old Title"}`, string(loaded[0].OldValue))
	assert.Equal(t, `{"title":"Task for History"}`, string(loaded[0].NewValue))
	assert.False(t, loaded[0].ChangedAt.IsZero())
}
