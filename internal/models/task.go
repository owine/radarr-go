// Package models defines data structures and database models for Radarr.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Task represents a scheduled or queued task in the system
type Task struct {
	ID          int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string         `json:"name" gorm:"not null;size:255;index"`
	CommandName string         `json:"commandName" gorm:"not null;size:255"`
	Message     string         `json:"message" gorm:"size:1000"`
	Body        TaskBody       `json:"body" gorm:"type:text"`
	Priority    TaskPriority   `json:"priority" gorm:"not null;default:normal;index"`
	Status      TaskStatus     `json:"status" gorm:"not null;default:queued;index"`
	QueuedAt    time.Time      `json:"queuedAt" gorm:"not null;index"`
	StartedAt   *time.Time     `json:"startedAt,omitempty" gorm:"index"`
	EndedAt     *time.Time     `json:"endedAt,omitempty" gorm:"index"`
	Duration    *time.Duration `json:"duration,omitempty"`
	Exception   string         `json:"exception,omitempty" gorm:"type:text"`
	Trigger     TaskTrigger    `json:"trigger" gorm:"not null;default:manual"`

	// Progress tracking
	Progress TaskProgress `json:"progress" gorm:"type:text"`

	// Scheduling
	Interval      *time.Duration `json:"interval,omitempty"`
	LastExecution *time.Time     `json:"lastExecution,omitempty" gorm:"index"`
	NextExecution *time.Time     `json:"nextExecution,omitempty" gorm:"index"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	// TaskStatusQueued indicates the task is waiting to be executed
	TaskStatusQueued TaskStatus = "queued"
	// TaskStatusStarted indicates the task is currently being executed
	TaskStatusStarted TaskStatus = "started"
	// TaskStatusCompleted indicates the task completed successfully
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates the task failed with an error
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusAborted indicates the task was manually cancelled
	TaskStatusAborted TaskStatus = "aborted"
	// TaskStatusCancelling indicates the task is being cancelled
	TaskStatusCancelling TaskStatus = "cancelling"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	// TaskPriorityHigh for critical tasks that need immediate attention
	TaskPriorityHigh TaskPriority = "high"
	// TaskPriorityNormal for standard tasks
	TaskPriorityNormal TaskPriority = "normal"
	// TaskPriorityLow for background maintenance tasks
	TaskPriorityLow TaskPriority = "low"
)

// TaskTrigger represents what triggered the task
type TaskTrigger string

const (
	// TaskTriggerManual for manually triggered tasks
	TaskTriggerManual TaskTrigger = "manual"
	// TaskTriggerScheduled for tasks triggered by scheduler
	TaskTriggerScheduled TaskTrigger = "scheduled"
	// TaskTriggerSystem for system-triggered tasks
	TaskTriggerSystem TaskTrigger = "system"
	// TaskTriggerAPI for API-triggered tasks
	TaskTriggerAPI TaskTrigger = "api"
)

// TaskBody holds the task-specific parameters and configuration
type TaskBody map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (tb TaskBody) Value() (driver.Value, error) {
	return json.Marshal(tb)
}

// Scan implements the sql.Scanner interface for database retrieval
func (tb *TaskBody) Scan(value interface{}) error {
	if value == nil {
		*tb = make(TaskBody)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, tb)
}

// TaskProgress tracks the progress and status messages of a running task
type TaskProgress struct {
	ProgressPercent int      `json:"progressPercent"`
	CurrentMessage  string   `json:"currentMessage"`
	ProcessedCount  int      `json:"processedCount"`
	TotalCount      int      `json:"totalCount"`
	StatusMessages  []string `json:"statusMessages"`
}

// Value implements the driver.Valuer interface for database storage
func (tp TaskProgress) Value() (driver.Value, error) {
	return json.Marshal(tp)
}

// Scan implements the sql.Scanner interface for database retrieval
func (tp *TaskProgress) Scan(value interface{}) error {
	if value == nil {
		*tp = TaskProgress{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, tp)
}

// UpdateProgress updates the task progress with a new message and percentage
func (tp *TaskProgress) UpdateProgress(percent int, message string) {
	tp.ProgressPercent = percent
	tp.CurrentMessage = message
	if message != "" {
		tp.StatusMessages = append(tp.StatusMessages, message)
		// Keep only last 100 messages to prevent bloat
		if len(tp.StatusMessages) > 100 {
			tp.StatusMessages = tp.StatusMessages[len(tp.StatusMessages)-100:]
		}
	}
}

// ScheduledTask represents a recurring task configuration
type ScheduledTask struct {
	ID          int           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string        `json:"name" gorm:"not null;size:255;uniqueIndex"`
	CommandName string        `json:"commandName" gorm:"not null;size:255"`
	Body        TaskBody      `json:"body" gorm:"type:text"`
	Interval    time.Duration `json:"interval" gorm:"not null"`
	Priority    TaskPriority  `json:"priority" gorm:"not null;default:normal"`
	Enabled     bool          `json:"enabled" gorm:"not null;default:true"`

	// Scheduling
	LastRun *time.Time `json:"lastRun,omitempty"`
	NextRun time.Time  `json:"nextRun" gorm:"not null;index"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TaskQueue represents a task execution queue
type TaskQueue struct {
	ID            int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string `json:"name" gorm:"not null;size:255;uniqueIndex"`
	MaxWorkers    int    `json:"maxWorkers" gorm:"not null;default:3"`
	ActiveWorkers int    `json:"activeWorkers" gorm:"not null;default:0"`
	QueuedTasks   int    `json:"queuedTasks" gorm:"not null;default:0"`
	RunningTasks  int    `json:"runningTasks" gorm:"not null;default:0"`
	Enabled       bool   `json:"enabled" gorm:"not null;default:true"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// BeforeCreate hook validates task data before creation
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.Name == "" {
		return gorm.ErrInvalidValue
	}
	if t.CommandName == "" {
		return gorm.ErrInvalidValue
	}
	if t.QueuedAt.IsZero() {
		t.QueuedAt = time.Now()
	}
	if t.Status == "" {
		t.Status = TaskStatusQueued
	}
	if t.Priority == "" {
		t.Priority = TaskPriorityNormal
	}
	if t.Trigger == "" {
		t.Trigger = TaskTriggerManual
	}
	if t.Body == nil {
		t.Body = make(TaskBody)
	}
	return nil
}

// BeforeUpdate hook calculates duration when task ends
func (t *Task) BeforeUpdate(_ *gorm.DB) error {
	if t.StartedAt != nil && t.EndedAt != nil && t.Duration == nil {
		duration := t.EndedAt.Sub(*t.StartedAt)
		t.Duration = &duration
	}
	return nil
}

// IsFinished returns true if the task has completed (success or failure)
func (t *Task) IsFinished() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed || t.Status == TaskStatusAborted
}

// IsRunning returns true if the task is currently executing
func (t *Task) IsRunning() bool {
	return t.Status == TaskStatusStarted || t.Status == TaskStatusCancelling
}

// CanBeCancelled returns true if the task can be cancelled
func (t *Task) CanBeCancelled() bool {
	return t.Status == TaskStatusQueued || t.Status == TaskStatusStarted
}

// BeforeCreate hook validates scheduled task data before creation
func (st *ScheduledTask) BeforeCreate(_ *gorm.DB) error {
	if st.Name == "" {
		return gorm.ErrInvalidValue
	}
	if st.CommandName == "" {
		return gorm.ErrInvalidValue
	}
	if st.Interval <= 0 {
		return gorm.ErrInvalidValue
	}
	if st.NextRun.IsZero() {
		st.NextRun = time.Now().Add(st.Interval)
	}
	if st.Priority == "" {
		st.Priority = TaskPriorityNormal
	}
	if st.Body == nil {
		st.Body = make(TaskBody)
	}
	return nil
}

// UpdateNextRun calculates and updates the next run time
func (st *ScheduledTask) UpdateNextRun() {
	st.NextRun = time.Now().Add(st.Interval)
}

// ShouldRun returns true if the scheduled task should be executed now
func (st *ScheduledTask) ShouldRun() bool {
	return st.Enabled && time.Now().After(st.NextRun)
}

// BeforeCreate hook validates task queue data before creation
func (tq *TaskQueue) BeforeCreate(_ *gorm.DB) error {
	if tq.Name == "" {
		return gorm.ErrInvalidValue
	}
	if tq.MaxWorkers <= 0 {
		tq.MaxWorkers = 3
	}
	return nil
}
