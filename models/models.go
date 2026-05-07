package models

import "time"

// Task represents a task in the system
type Task struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Points    int       `json:"points"`
	TaskType  string    `json:"task_type"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskCreateRequest is the request body for creating a task
type TaskCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Points   int    `json:"points" binding:"required"`
	TaskType string `json:"task_type" binding:"required"`
}

// ChildProgress records completion of one-time tasks
type ChildProgress struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// DailyTaskStatus records daily completion of periodic tasks
type DailyTaskStatus struct {
	ID        int64  `json:"id"`
	TaskID    int64  `json:"task_id"`
	Date      string `json:"date"`
	Completed bool   `json:"completed"`
}

// PointHistory represents a point change record
type PointHistory struct {
	ID           int64     `json:"id"`
	PointsChange int       `json:"points_change"`
	Reason       string    `json:"reason"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
}

// PointAdjustRequest is the request body for manual point adjustment
type PointAdjustRequest struct {
	PointsChange int    `json:"points_change" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
}

// RedeemItem represents a redeemable prize
type RedeemItem struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PointsRequired int    `json:"points_required"`
	Active         bool   `json:"active"`
}

// RedeemRequest is the request body for redeeming a prize
type RedeemRequest struct {
	ItemID int64 `json:"item_id" binding:"required"`
}

// ChildTask represents a task as seen by the child (available to complete)
type ChildTask struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Points    int    `json:"points"`
	TaskType  string `json:"task_type"`
	Completed bool   `json:"completed"`
}

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PointsResponse returns current points
type PointsResponse struct {
	TotalPoints int `json:"total_points"`
}

// CompleteTaskRequest is the request body for completing a task
type CompleteTaskRequest struct {
	TaskID int64 `json:"task_id" binding:"required"`
}
