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
