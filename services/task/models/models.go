package models

import "time"

// TaskStatus статусы задач
type TaskStatus string

const (
	StatusOpen       TaskStatus = "open"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusConfirmed  TaskStatus = "confirmed"
)

// Task задача
type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatorID   int        `json:"creator_id"`
	AssigneeID  int        `json:"assignee_id"`
	RoomID      int        `json:"room_id"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// CreateTaskRequest запрос создания задачи
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	AssigneeID  int    `json:"assignee_id"`
	RoomID      int    `json:"room_id"`
}

// UpdateTaskRequest запрос обновления задачи
type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AssigneeID  int        `json:"assignee_id"`
	Status      TaskStatus `json:"status"`
}
