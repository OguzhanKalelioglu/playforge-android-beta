package model

import (
	"time"

	"github.com/google/uuid"
)

// TestStatus
type TestStatus string

const (
	TestStatusPending   TestStatus = "pending"
	TestStatusActive    TestStatus = "active"
	TestStatusCompleted TestStatus = "completed"
	TestStatusFailed    TestStatus = "failed"
	TestStatusCancelled TestStatus = "cancelled"
)

// TestDTO, API ↔ orchestrator arasında test bilgisi
type TestDTO struct {
	ID             uuid.UUID   `json:"id"`
	PackageName    string      `json:"package_name"`
	TestLink       string      `json:"test_link"`
	Status         TestStatus  `json:"status"`
	StartedAt      *time.Time  `json:"started_at,omitempty"`
	EndsAt         *time.Time  `json:"ends_at,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	Assignments    []AssignmentDTO `json:"assignments,omitempty"`
}

// AssignmentStatus
type AssignmentStatus string

const (
	AssignmentStatusPending    AssignmentStatus = "pending"
	AssignmentStatusInProgress AssignmentStatus = "in_progress"
	AssignmentStatusCompleted  AssignmentStatus = "completed"
	AssignmentStatusFailed     AssignmentStatus = "failed"
	AssignmentStatusSkipped    AssignmentStatus = "skipped"
)

// AssignmentDTO, bir tester'ın teste atanması
type AssignmentDTO struct {
	ID               uuid.UUID        `json:"id"`
	TestID           uuid.UUID        `json:"test_id"`
	TesterID         uuid.UUID        `json:"tester_id"`
	TesterEmail      string           `json:"tester_email,omitempty"`
	Status           AssignmentStatus `json:"status"`
	OptInAt          *time.Time       `json:"opt_in_at,omitempty"`
	InstallAt        *time.Time       `json:"install_at,omitempty"`
	LastEngagementAt *time.Time       `json:"last_engagement_at,omitempty"`
	EmulatorSerial   string           `json:"emulator_serial,omitempty"`
}
