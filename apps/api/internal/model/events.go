package model

import (
	"encoding/json"
	"time"
)

// EventType, ActivityEvent tipi
type EventType string

const (
	EventStepStart    EventType = "step_start"
	EventStepComplete EventType = "step_complete"
	EventStepError    EventType = "step_error"
	EventTaskStart    EventType = "task_start"
	EventTaskComplete EventType = "task_complete"
	EventTaskError    EventType = "task_error"
	EventLog          EventType = "log"
)

// ActivityEvent, orchestrator → API POST edilen event
// activity_logs tablosuna INSERT edilir
type ActivityEvent struct {
	EventType      EventType       `json:"event_type"`
	TestID         string          `json:"test_id"`
	AssignmentID   string          `json:"assignment_id"`
	StepName       string          `json:"step_name,omitempty"`
	Status         string          `json:"status"` // "ok" | "error" | "in_progress"
	ErrorMessage   string          `json:"error_message,omitempty"`
	ScreenshotPath string          `json:"screenshot_path,omitempty"` // URL veya local path
	Metadata       json.RawMessage `json:"metadata,omitempty"`        // ek bilgi (gesture count, vs.)
	OccurredAt     time.Time       `json:"occurred_at"`
}

// BatchActivityEvents, birden fazla event'i tek seferde göndermek için
type BatchActivityEvents struct {
	Events []ActivityEvent `json:"events"`
}

// TaskStatus, bir task'ın genel durumu
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRetrying  TaskStatus = "retrying"
)

// TaskStateUpdate, task durumu değiştiğinde API'ye bildirilir (opsiyonel, real-time dashboard için)
type TaskStateUpdate struct {
	TaskType    TaskType   `json:"task_type"`
	TestID      string     `json:"test_id"`
	AssignmentID string    `json:"assignment_id"`
	Status      TaskStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}
