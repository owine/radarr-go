// Package models defines data structures and database models for Radarr.
// This file contains the refactored Task models with simplified validation and proper constraints.
package models

import (
	"time"
)

// TaskV2 represents a task with simplified structure and no problematic GORM hooks
type TaskV2 struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"not null;size:255"`
	CommandName string `json:"commandName" gorm:"not null;size:255"`

	// Status and execution (simplified with check constraints)
	Status   string `json:"status" gorm:"not null;default:'queued';size:20"`
	Priority string `json:"priority" gorm:"not null;default:'normal';size:20"`

	// Timing
	QueuedAt   time.Time  `json:"queuedAt" gorm:"not null;autoCreateTime"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	EndedAt    *time.Time `json:"endedAt,omitempty"`
	DurationMs *int64     `json:"durationMs,omitempty"` // Duration in milliseconds

	// Data and results as JSON
	Body         JSONField `json:"body" gorm:"type:json"`
	Result       JSONField `json:"result,omitempty" gorm:"type:json"`
	ErrorMessage string    `json:"errorMessage,omitempty" gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (TaskV2) TableName() string {
	return "tasks"
}

// Validate performs basic validation
func (t *TaskV2) Validate() error {
	if t.Name == "" {
		return ValidationError{Field: "name", Message: "Name is required"}
	}
	if t.CommandName == "" {
		return ValidationError{Field: "command_name", Message: "Command name is required"}
	}
	return nil
}

// Start marks the task as started
func (t *TaskV2) Start() {
	now := time.Now()
	t.Status = string(TaskStatusStarted)
	t.StartedAt = &now
}

// Complete marks the task as completed and calculates duration
func (t *TaskV2) Complete() {
	now := time.Now()
	t.Status = string(TaskStatusCompleted)
	t.EndedAt = &now
	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt).Milliseconds()
		t.DurationMs = &duration
	}
}

// Fail marks the task as failed with an error message
func (t *TaskV2) Fail(errorMsg string) {
	now := time.Now()
	t.Status = string(TaskStatusFailed)
	t.EndedAt = &now
	t.ErrorMessage = errorMsg
	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt).Milliseconds()
		t.DurationMs = &duration
	}
}

// Abort marks the task as aborted
func (t *TaskV2) Abort() {
	now := time.Now()
	t.Status = string(TaskStatusAborted)
	t.EndedAt = &now
	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt).Milliseconds()
		t.DurationMs = &duration
	}
}

// IsFinished returns true if the task has completed (success or failure)
func (t *TaskV2) IsFinished() bool {
	return t.Status == string(TaskStatusCompleted) ||
		t.Status == string(TaskStatusFailed) ||
		t.Status == string(TaskStatusAborted)
}

// IsRunning returns true if the task is currently executing
func (t *TaskV2) IsRunning() bool {
	return t.Status == string(TaskStatusStarted)
}

// CanBeCancelled returns true if the task can be cancelled
func (t *TaskV2) CanBeCancelled() bool {
	return t.Status == string(TaskStatusQueued) || t.Status == string(TaskStatusStarted)
}

// ScheduledTaskV2 represents a recurring task configuration with simplified structure
type ScheduledTaskV2 struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"not null;size:255;uniqueIndex"`
	CommandName string `json:"commandName" gorm:"not null;size:255"`

	// Configuration
	Enabled    bool   `json:"enabled" gorm:"not null;default:true"`
	IntervalMs int64  `json:"intervalMs" gorm:"not null"` // Interval in milliseconds
	Priority   string `json:"priority" gorm:"not null;default:'normal';size:20"`

	// Scheduling
	LastRun *time.Time `json:"lastRun,omitempty"`
	NextRun time.Time  `json:"nextRun" gorm:"not null"`

	// Data
	Body JSONField `json:"body" gorm:"type:json"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (ScheduledTaskV2) TableName() string {
	return "scheduled_tasks"
}

// Validate performs basic validation
func (st *ScheduledTaskV2) Validate() error {
	if st.Name == "" {
		return ValidationError{Field: "name", Message: "Name is required"}
	}
	if st.CommandName == "" {
		return ValidationError{Field: "command_name", Message: "Command name is required"}
	}
	if st.IntervalMs <= 0 {
		return ValidationError{Field: "interval_ms", Message: "Interval must be positive"}
	}
	return nil
}

// UpdateNextRun calculates and updates the next run time based on interval
func (st *ScheduledTaskV2) UpdateNextRun() {
	st.NextRun = time.Now().Add(time.Duration(st.IntervalMs) * time.Millisecond)
	now := time.Now()
	st.LastRun = &now
}

// ShouldRun returns true if the scheduled task should be executed now
func (st *ScheduledTaskV2) ShouldRun() bool {
	return st.Enabled && time.Now().After(st.NextRun)
}

// Note: Task status and priority constants are defined in task.go

// AppConfigV2 represents application configuration key-value pairs
type AppConfigV2 struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string    `json:"key" gorm:"not null;size:100;uniqueIndex"`
	Value       JSONField `json:"value" gorm:"type:json"`
	Description string    `json:"description,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (AppConfigV2) TableName() string {
	return "app_config"
}
