// Package services provides business logic and domain services for Radarr.
package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// TaskService provides task scheduling and management functionality
type TaskService struct {
	db     *database.Database
	logger *logger.Logger

	// Task execution
	workers        map[string]*TaskWorkerPool
	handlers       map[string]TaskHandler
	scheduler      *TaskScheduler
	executionMutex sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// TaskHandler defines the interface for task execution handlers
type TaskHandler interface {
	// Execute runs the task with the given context and parameters
	Execute(ctx context.Context, task *models.Task, updateProgress func(percent int, message string)) error
	// GetName returns the command name this handler processes
	GetName() string
	// GetDescription returns a human-readable description of what this handler does
	GetDescription() string
}

// TaskWorkerPool manages a pool of workers for executing tasks
type TaskWorkerPool struct {
	name       string
	maxWorkers int
	workers    chan struct{}
	queue      chan *models.Task
	active     map[int]*models.Task
	activeMu   sync.RWMutex
	logger     *logger.Logger
}

// TaskScheduler manages recurring scheduled tasks
type TaskScheduler struct {
	service *TaskService
	ticker  *time.Ticker
	logger  *logger.Logger
}

// NewTaskService creates a new task service instance
func NewTaskService(db *database.Database, logger *logger.Logger) *TaskService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &TaskService{
		db:       db,
		logger:   logger,
		workers:  make(map[string]*TaskWorkerPool),
		handlers: make(map[string]TaskHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize scheduler
	service.scheduler = &TaskScheduler{
		service: service,
		ticker:  time.NewTicker(30 * time.Second), // Check for scheduled tasks every 30 seconds
		logger:  logger,
	}

	// Create default worker pools
	service.createWorkerPool("default", 3)
	service.createWorkerPool("high-priority", 2)
	service.createWorkerPool("background", 1)

	// Start scheduler
	go service.scheduler.run()

	return service
}

// RegisterHandler registers a task handler for a specific command
func (ts *TaskService) RegisterHandler(handler TaskHandler) {
	ts.executionMutex.Lock()
	defer ts.executionMutex.Unlock()

	ts.handlers[handler.GetName()] = handler
	ts.logger.Infow("Registered task handler", "command", handler.GetName(), "description", handler.GetDescription())
}

// QueueTask queues a new task for execution
func (ts *TaskService) QueueTask(name, commandName string, body models.TaskBody, priority models.TaskPriority, trigger models.TaskTrigger) (*models.Task, error) {
	task := &models.Task{
		Name:        name,
		CommandName: commandName,
		Body:        body,
		Priority:    priority,
		Status:      models.TaskStatusQueued,
		Trigger:     trigger,
		QueuedAt:    time.Now(),
		Progress:    models.TaskProgress{},
	}

	if err := ts.db.GORM.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to queue task: %w", err)
	}

	// Add to appropriate worker pool
	poolName := ts.getPoolNameForTask(task)
	if pool, exists := ts.workers[poolName]; exists {
		select {
		case pool.queue <- task:
			ts.logger.Infow("Task queued for execution",
				"taskId", task.ID, "command", commandName, "pool", poolName)
		default:
			ts.logger.Warnw("Task queue full, task will be picked up by scheduler",
				"taskId", task.ID, "command", commandName, "pool", poolName)
		}
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (ts *TaskService) GetTask(id int) (*models.Task, error) {
	var task models.Task
	if err := ts.db.GORM.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks retrieves tasks with optional filtering
func (ts *TaskService) ListTasks(status models.TaskStatus, commandName string, limit, offset int) ([]*models.Task, int64, error) {
	query := ts.db.GORM.Model(&models.Task{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if commandName != "" {
		query = query.Where("command_name = ?", commandName)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	var tasks []*models.Task
	if err := query.Order("queued_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

// CancelTask cancels a running or queued task
func (ts *TaskService) CancelTask(id int) error {
	task, err := ts.GetTask(id)
	if err != nil {
		return err
	}

	if !task.CanBeCancelled() {
		return fmt.Errorf("task %d cannot be cancelled (status: %s)", id, task.Status)
	}

	// Update status to cancelling
	updates := map[string]interface{}{
		"status":     models.TaskStatusCancelling,
		"updated_at": time.Now(),
	}

	if err := ts.db.GORM.Model(task).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	ts.logger.Info("Task marked for cancellation", "taskId", id)
	return nil
}

// CreateScheduledTask creates a new scheduled task
func (ts *TaskService) CreateScheduledTask(name, commandName string, body models.TaskBody, interval time.Duration, priority models.TaskPriority) (*models.ScheduledTask, error) {
	scheduledTask := &models.ScheduledTask{
		Name:        name,
		CommandName: commandName,
		Body:        body,
		Interval:    interval,
		Priority:    priority,
		Enabled:     true,
		NextRun:     time.Now().Add(interval),
	}

	if err := ts.db.GORM.Create(scheduledTask).Error; err != nil {
		return nil, fmt.Errorf("failed to create scheduled task: %w", err)
	}

	ts.logger.Infow("Scheduled task created",
		"name", name, "command", commandName, "interval", interval)

	return scheduledTask, nil
}

// GetScheduledTasks retrieves all scheduled tasks
func (ts *TaskService) GetScheduledTasks() ([]*models.ScheduledTask, error) {
	var scheduledTasks []*models.ScheduledTask
	if err := ts.db.GORM.Order("name").Find(&scheduledTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks: %w", err)
	}
	return scheduledTasks, nil
}

// UpdateScheduledTask updates a scheduled task
func (ts *TaskService) UpdateScheduledTask(id int, updates map[string]interface{}) error {
	if err := ts.db.GORM.Model(&models.ScheduledTask{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update scheduled task: %w", err)
	}
	return nil
}

// DeleteScheduledTask removes a scheduled task
func (ts *TaskService) DeleteScheduledTask(id int) error {
	if err := ts.db.GORM.Delete(&models.ScheduledTask{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete scheduled task: %w", err)
	}
	return nil
}

// GetQueueStatus returns the current status of all worker pools
func (ts *TaskService) GetQueueStatus() map[string]interface{} {
	ts.executionMutex.RLock()
	defer ts.executionMutex.RUnlock()

	status := make(map[string]interface{})
	for name, pool := range ts.workers {
		pool.activeMu.RLock()
		status[name] = map[string]interface{}{
			"maxWorkers":    pool.maxWorkers,
			"activeWorkers": len(pool.active),
			"queuedTasks":   len(pool.queue),
			"activeTasks":   pool.getActiveTasks(),
		}
		pool.activeMu.RUnlock()
	}
	return status
}

// createWorkerPool creates a new worker pool with the specified configuration
func (ts *TaskService) createWorkerPool(name string, maxWorkers int) {
	pool := &TaskWorkerPool{
		name:       name,
		maxWorkers: maxWorkers,
		workers:    make(chan struct{}, maxWorkers),
		queue:      make(chan *models.Task, 100), // Buffer 100 tasks
		active:     make(map[int]*models.Task),
		logger:     ts.logger,
	}

	ts.workers[name] = pool

	// Start worker goroutines
	for i := 0; i < maxWorkers; i++ {
		go pool.worker(ts.ctx, ts)
	}

	ts.logger.Infow("Created task worker pool",
		"pool", name, "maxWorkers", maxWorkers)
}

// getPoolNameForTask determines which worker pool should handle a task
func (ts *TaskService) getPoolNameForTask(task *models.Task) string {
	switch task.Priority {
	case models.TaskPriorityHigh:
		return "high-priority"
	case models.TaskPriorityNormal:
		return "default"
	case models.TaskPriorityLow:
		return "background"
	default:
		return "default"
	}
}

// worker processes tasks from the queue
func (pool *TaskWorkerPool) worker(ctx context.Context, service *TaskService) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-pool.queue:
			// Acquire worker slot
			pool.workers <- struct{}{}

			// Track active task
			pool.activeMu.Lock()
			pool.active[task.ID] = task
			pool.activeMu.Unlock()

			// Execute task
			pool.executeTask(ctx, service, task)

			// Release worker slot
			<-pool.workers

			// Remove from active tasks
			pool.activeMu.Lock()
			delete(pool.active, task.ID)
			pool.activeMu.Unlock()
		}
	}
}

// executeTask runs a single task
func (pool *TaskWorkerPool) executeTask(ctx context.Context, service *TaskService, task *models.Task) {
	logger := pool.logger.With("taskId", task.ID, "command", task.CommandName)

	// Check if task was cancelled before execution
	var currentStatus models.TaskStatus
	if err := service.db.GORM.Model(&models.Task{}).Where("id = ?", task.ID).
		Select("status").Scan(&currentStatus).Error; err != nil {
		logger.Errorw("Failed to check task status before execution", "error", err)
		return
	}

	if currentStatus == models.TaskStatusCancelling {
		if err := service.updateTaskStatus(task.ID, models.TaskStatusAborted, "Task was cancelled before execution", nil); err != nil {
			logger.Errorw("Failed to update task status to aborted", "taskId", task.ID, "error", err)
		}
		return
	}

	// Mark task as started
	startTime := time.Now()
	if err := service.updateTaskStatus(task.ID, models.TaskStatusStarted, "Task execution started", &startTime); err != nil {
		logger.Errorw("Failed to update task status to started", "error", err)
		return
	}

	logger.Infow("Task execution started")

	// Find handler for this command
	service.executionMutex.RLock()
	handler, exists := service.handlers[task.CommandName]
	service.executionMutex.RUnlock()

	if !exists {
		endTime := time.Now()
		err := fmt.Errorf("no handler registered for command: %s", task.CommandName)
		if updateErr := service.updateTaskStatus(task.ID, models.TaskStatusFailed, err.Error(), &endTime); updateErr != nil {
			logger.Errorw("Failed to update task status to failed", "taskId", task.ID, "error", updateErr)
		}
		logger.Errorw("Task failed: no handler", "error", err)
		return
	}

	// Create progress update function
	updateProgress := func(percent int, message string) {
		if err := service.updateTaskProgress(task.ID, percent, message); err != nil {
			logger.Errorw("Failed to update task progress", "taskId", task.ID, "error", err)
		}
	}

	// Execute the task
	var taskErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				taskErr = fmt.Errorf("task panicked: %v", r)
				logger.Errorw("Task panicked", "panic", r)
			}
		}()

		// Create task-specific context with cancellation
		taskCtx, taskCancel := context.WithCancel(ctx)
		defer taskCancel()

		// Start goroutine to monitor for cancellation
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-taskCtx.Done():
					return
				case <-ticker.C:
					var status models.TaskStatus
					if err := service.db.GORM.Model(&models.Task{}).Where("id = ?", task.ID).
						Select("status").Scan(&status).Error; err != nil {
						continue
					}
					if status == models.TaskStatusCancelling {
						taskCancel()
						return
					}
				}
			}
		}()

		taskErr = handler.Execute(taskCtx, task, updateProgress)
	}()

	// Update task status based on result
	endTime := time.Now()
	if taskErr != nil {
		if ctx.Err() != nil || errors.Is(taskErr, context.Canceled) {
			if err := service.updateTaskStatus(task.ID, models.TaskStatusAborted, "Task was cancelled", &endTime); err != nil {
				logger.Errorw("Failed to update task status to aborted", "taskId", task.ID, "error", err)
			}
			logger.Infow("Task was cancelled")
		} else {
			if err := service.updateTaskStatus(task.ID, models.TaskStatusFailed, taskErr.Error(), &endTime); err != nil {
				logger.Errorw("Failed to update task status to failed", "taskId", task.ID, "error", err)
			}
			logger.Errorw("Task failed", "error", taskErr)
		}
	} else {
		if err := service.updateTaskStatus(task.ID, models.TaskStatusCompleted, "Task completed successfully", &endTime); err != nil {
			logger.Errorw("Failed to update task status to completed", "taskId", task.ID, "error", err)
		}
		logger.Infow("Task completed successfully", "duration", endTime.Sub(startTime))
	}
}

// getActiveTasks returns a slice of active task IDs
func (pool *TaskWorkerPool) getActiveTasks() []int {
	pool.activeMu.RLock()
	defer pool.activeMu.RUnlock()

	var tasks []int
	for id := range pool.active {
		tasks = append(tasks, id)
	}
	return tasks
}

// updateTaskStatus updates the status and timing of a task
func (ts *TaskService) updateTaskStatus(taskID int, status models.TaskStatus, message string, timestamp *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if message != "" {
		updates["message"] = message
		if status == models.TaskStatusFailed {
			updates["exception"] = message
		}
	}

	if timestamp != nil {
		switch status {
		case models.TaskStatusStarted:
			updates["started_at"] = *timestamp
		case models.TaskStatusCompleted, models.TaskStatusFailed, models.TaskStatusAborted:
			updates["ended_at"] = *timestamp
		case models.TaskStatusQueued:
			// Queued tasks don't need timestamp updates
		case models.TaskStatusCancelling:
			// Cancelling tasks don't need timestamp updates
		}
	}

	return ts.db.GORM.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

// updateTaskProgress updates the progress information of a task
func (ts *TaskService) updateTaskProgress(taskID int, percent int, message string) error {
	// Get current progress
	var currentProgress models.TaskProgress
	if err := ts.db.GORM.Model(&models.Task{}).Where("id = ?", taskID).
		Select("progress").Scan(&currentProgress).Error; err != nil {
		return err
	}

	// Update progress
	currentProgress.UpdateProgress(percent, message)

	// Save to database
	return ts.db.GORM.Model(&models.Task{}).Where("id = ?", taskID).
		Update("progress", currentProgress).Error
}

// run executes the task scheduler loop
func (scheduler *TaskScheduler) run() {
	for {
		select {
		case <-scheduler.service.ctx.Done():
			scheduler.ticker.Stop()
			return
		case <-scheduler.ticker.C:
			scheduler.processScheduledTasks()
		}
	}
}

// processScheduledTasks checks for and executes scheduled tasks
func (scheduler *TaskScheduler) processScheduledTasks() {
	var scheduledTasks []*models.ScheduledTask
	if err := scheduler.service.db.GORM.Where("enabled = ? AND next_run <= ?", true, time.Now()).
		Find(&scheduledTasks).Error; err != nil {
		scheduler.logger.Errorw("Failed to fetch scheduled tasks", "error", err)
		return
	}

	for _, scheduledTask := range scheduledTasks {
		// Queue the task
		_, err := scheduler.service.QueueTask(
			scheduledTask.Name,
			scheduledTask.CommandName,
			scheduledTask.Body,
			scheduledTask.Priority,
			models.TaskTriggerScheduled,
		)

		if err != nil {
			scheduler.logger.Errorw("Failed to queue scheduled task",
				"error", err, "scheduledTaskId", scheduledTask.ID, "command", scheduledTask.CommandName)
			continue
		}

		// Update next run time
		scheduledTask.LastRun = &scheduledTask.NextRun
		scheduledTask.UpdateNextRun()

		if err := scheduler.service.db.GORM.Save(scheduledTask).Error; err != nil {
			scheduler.logger.Errorw("Failed to update scheduled task next run time",
				"error", err, "scheduledTaskId", scheduledTask.ID)
		}

		scheduler.logger.Infow("Scheduled task queued",
			"scheduledTaskId", scheduledTask.ID, "command", scheduledTask.CommandName, "nextRun", scheduledTask.NextRun)
	}
}

// Shutdown gracefully shuts down the task service
func (ts *TaskService) Shutdown() {
	ts.logger.Infow("Shutting down task service")
	ts.cancel()
}
