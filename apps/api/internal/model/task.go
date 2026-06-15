package model

import (
	"encoding/json"
	"fmt"
)

// TaskType, Asynq task type string'i
type TaskType string

const (
	TaskTypeLoginGoogle      TaskType = "login_google"
	TaskTypeTestStart        TaskType = "test_start"
	TaskTypeDailyEngagement  TaskType = "daily_engagement"
	TaskTypeWriteReview      TaskType = "write_review"
	TaskTypeHealthcheck      TaskType = "healthcheck"
)

// Payload — ortak alanlar
type Payload struct {
	TestID         string `json:"test_id"`
	AssignmentID   string `json:"assignment_id"`
	PackageName    string `json:"package_name,omitempty"`
}

// LoginGoogle: Google hesabı ekleme (warming sonrası, TestStart öncesi)
// Email + App Password kullanılır (2FA yok çünkü hesaplarda 2FA kapalı)
type LoginGooglePayload struct {
	Payload
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TestStart: opt-in → download → initial 5dk engage
type TestStartPayload struct {
	Payload
	TesterID    string `json:"tester_id"`
	TestLink    string `json:"test_link"`
}

// DailyEngagement: günlük 2-5dk rastgele etkileşim
type DailyEngagementPayload struct {
	Payload
	Day int `json:"day"`
}

// WriteReview: 14. gün, 5 yıldız + comment + post
type WriteReviewPayload struct {
	Payload
	TesterID  string `json:"tester_id"`
	Stars     int    `json:"stars"`
	Comment   string `json:"comment"`
	Language  string `json:"language"`
}

// Healthcheck: günlük 23:00, emulator/hesap/storage sağlık kontrolü
type HealthcheckPayload struct {
	Payload
}

// ============================================
// Encode / Decode helpers
// ============================================

func (p LoginGooglePayload) Encode() ([]byte, error)      { return json.Marshal(p) }
func (p TestStartPayload) Encode() ([]byte, error)        { return json.Marshal(p) }
func (p DailyEngagementPayload) Encode() ([]byte, error)  { return json.Marshal(p) }
func (p WriteReviewPayload) Encode() ([]byte, error)      { return json.Marshal(p) }
func (p HealthcheckPayload) Encode() ([]byte, error)      { return json.Marshal(p) }

func DecodeLoginGoogle(b []byte) (LoginGooglePayload, error) {
	var p LoginGooglePayload
	err := json.Unmarshal(b, &p)
	return p, err
}
func DecodeTestStart(b []byte) (TestStartPayload, error) {
	var p TestStartPayload
	err := json.Unmarshal(b, &p)
	return p, err
}
func DecodeDailyEngagement(b []byte) (DailyEngagementPayload, error) {
	var p DailyEngagementPayload
	err := json.Unmarshal(b, &p)
	return p, err
}
func DecodeWriteReview(b []byte) (WriteReviewPayload, error) {
	var p WriteReviewPayload
	err := json.Unmarshal(b, &p)
	return p, err
}
func DecodeHealthcheck(b []byte) (HealthcheckPayload, error) {
	var p HealthcheckPayload
	err := json.Unmarshal(b, &p)
	return p, err
}

// JobID, Asynq scheduler için stable unique ID
// Format: {testID}:{type}:{day}
// Aynı job iki kez enqueue edilirse Asynq no-op yapar
func JobID(testID string, taskType TaskType, day int) string {
	if day < 0 {
		day = 0
	}
	return fmt.Sprintf("%s:%s:%d", testID, taskType, day)
}
