package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamRole string

const (
	TeamRoleOwner  TeamRole = "owner"
	TeamRoleAdmin  TeamRole = "admin"
	TeamRoleMember TeamRole = "member"
)

type TeamMember struct {
	UserID   int64
	TeamID   int64
	Role     TeamRole
	JoinedAt time.Time
}

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      TaskStatus
	AssigneeID  int64
	TeamID      int64
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskFilter struct {
	TeamID     int64
	Status     TaskStatus
	AssigneeID int64
}

type Pagination struct {
	Limit  int
	Offset int
}

type TaskHistory struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	ChangedAt time.Time
	OldValue  json.RawMessage
	NewValue  json.RawMessage
}

type TaskComment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TeamStats struct {
	TeamID             int64 `json:"team_id"`
	TeamName           string `json:"team_name"`
	MemberCount        int64 `json:"member_count"`
	DoneTasksLast7Days int64 `json:"done_tasks_last_7_days"`
}

type TopUser struct {
	TeamID    int64  `json:"team_id"`
	UserID    int64  `json:"user_id"`
	UserEmail string `json:"user_email"`
	TaskCount int64  `json:"task_count"`
	Rank      int64  `json:"rank"`
}

type IntegrityViolation struct {
	TaskID     int64  `json:"task_id"`
	TaskTitle  string `json:"task_title"`
	AssigneeID int64  `json:"assignee_id"`
	TeamID     int64  `json:"team_id"`
}
