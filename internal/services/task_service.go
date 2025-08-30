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
	"go.uber.org/zap"
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
	Execute(ctx context.Context, task *models.TaskV2, updateProgress func(percent int, message string)) error
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
	queue      chan *models.TaskV2
	active     map[int]*models.TaskV2
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
func (ts *TaskService) QueueTask(
	name, commandName string,
	body models.JSONField,
	priority string,
) (*models.TaskV2, error) {
	task := &models.TaskV2{
		Name:        name,
		CommandName: commandName,
		Body:        body,
		Priority:    priority,
		Status:      "queued",
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
func (ts *TaskService) GetTask(id int) (*models.TaskV2, error) {
	var task models.TaskV2
	if err := ts.db.GORM.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// ListTasks retrieves tasks with optional filtering
func (ts *TaskService) ListTasks(
	status string,
	commandName string,
	limit, offset int,
) ([]*models.TaskV2, int64, error) {
	query := ts.db.GORM.Model(&models.TaskV2{})

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

	var tasks []*models.TaskV2
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
		"status":     "cancelling",
		"updated_at": time.Now(),
	}

	if err := ts.db.GORM.Model(task).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	ts.logger.Info("Task marked for cancellation", "taskId", id)
	return nil
}

// CreateScheduledTask creates a new scheduled task
func (ts *TaskService) CreateScheduledTask(
	name, commandName string,
	body models.JSONField,
	interval time.Duration,
	priority string,
) (*models.ScheduledTaskV2, error) {
	scheduledTask := &models.ScheduledTaskV2{
		Name:        name,
		CommandName: commandName,
		Body:        body,
		IntervalMs:  interval.Milliseconds(),
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
func (ts *TaskService) GetScheduledTasks() ([]*models.ScheduledTaskV2, error) {
	var scheduledTasks []*models.ScheduledTaskV2
	if err := ts.db.GORM.Order("name").Find(&scheduledTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks: %w", err)
	}
	return scheduledTasks, nil
}

// UpdateScheduledTask updates a scheduled task
func (ts *TaskService) UpdateScheduledTask(id int, updates map[string]interface{}) error {
	if err := ts.db.GORM.Model(&models.ScheduledTaskV2{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update scheduled task: %w", err)
	}
	return nil
}

// DeleteScheduledTask removes a scheduled task
func (ts *TaskService) DeleteScheduledTask(id int) error {
	if err := ts.db.GORM.Delete(&models.ScheduledTaskV2{}, id).Error; err != nil {
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
		queue:      make(chan *models.TaskV2, 100), // Buffer 100 tasks
		active:     make(map[int]*models.TaskV2),
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
func (ts *TaskService) getPoolNameForTask(task *models.TaskV2) string {
	switch task.Priority {
	case "high":
		return "high-priority"
	case "normal":
		return "default"
	case "low":
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
func (pool *TaskWorkerPool) executeTask(ctx context.Context, service *TaskService, task *models.TaskV2) {
	if pool.shouldAbortTaskBeforeExecution(service, task) {
		return
	}

	startTime := pool.markTaskAsStarted(service, task)
	if startTime.IsZero() {
		return
	}

	handler := pool.findTaskHandler(service, task)
	if handler == nil {
		return
	}

	taskErr := pool.executeTaskWithHandler(ctx, service, task, handler)

	pool.finalizeTaskExecution(service, task, taskErr, startTime)
}

// shouldAbortTaskBeforeExecution checks if task should be aborted before execution
func (pool *TaskWorkerPool) shouldAbortTaskBeforeExecution(service *TaskService, task *models.TaskV2) bool {
	logger := pool.logger.With("taskId", task.ID)
	var currentStatus string
	if err := service.db.GORM.Model(&models.TaskV2{}).Where("id = ?", task.ID).
		Select("status").Scan(&currentStatus).Error; err != nil {
		logger.Errorw("Failed to check task status before execution", "error", err)
		return true
	}

	if currentStatus == "cancelling" {
		err := service.updateTaskStatus(task.ID, "aborted",
			"Task was cancelled before execution", nil)
		if err != nil {
			logger.Errorw("Failed to update task status to aborted", "taskId", task.ID, "error", err)
		}
		return true
	}

	return false
}

// markTaskAsStarted marks task as started and returns start time
func (pool *TaskWorkerPool) markTaskAsStarted(service *TaskService, task *models.TaskV2) time.Time {
	logger := pool.logger.With("taskId", task.ID)
	startTime := time.Now()
	err := service.updateTaskStatus(task.ID, "started",
		"Task execution started", &startTime)
	if err != nil {
		logger.Errorw("Failed to update task status to started", "error", err)
		return time.Time{} // Return zero time to indicate failure
	}

	logger.Infow("Task execution started")
	return startTime
}

// findTaskHandler finds and validates the handler for a task
func (pool *TaskWorkerPool) findTaskHandler(service *TaskService, task *models.TaskV2) TaskHandler {
	logger := pool.logger.With("taskId", task.ID)
	service.executionMutex.RLock()
	handler, exists := service.handlers[task.CommandName]
	service.executionMutex.RUnlock()

	if !exists {
		endTime := time.Now()
		err := fmt.Errorf("no handler registered for command: %s", task.CommandName)
		if updateErr := service.updateTaskStatus(task.ID, "failed", err.Error(), &endTime); updateErr != nil {
			logger.Errorw("Failed to update task status to failed", "taskId", task.ID, "error", updateErr)
		}
		logger.Errorw("Task failed: no handler", "error", err)
		return nil
	}

	return handler
}

// executeTaskWithHandler executes the task using the provided handler
func (pool *TaskWorkerPool) executeTaskWithHandler(
	ctx context.Context, service *TaskService, task *models.TaskV2,
	handler TaskHandler,
) error {
	logger := pool.logger.With("taskId", task.ID)
	// Create progress update function
	updateProgress := func(percent int, message string) {
		if err := service.updateTaskProgress(task.ID, percent, message); err != nil {
			logger.Errorw("Failed to update task progress", "taskId", task.ID, "error", err)
		}
	}

	// Execute the task with panic recovery
	var taskErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				taskErr = fmt.Errorf("task panicked: %v", r)
				logger.Errorw("Task panicked", "panic", r)
			}
		}()

		taskCtx := pool.createCancellableTaskContext(ctx, service, task)
		taskErr = handler.Execute(taskCtx, task, updateProgress)
	}()

	return taskErr
}

// createCancellableTaskContext creates a context that can be cancelled by monitoring task status
func (pool *TaskWorkerPool) createCancellableTaskContext(
	ctx context.Context, service *TaskService, task *models.TaskV2,
) context.Context {
	taskCtx, taskCancel := context.WithCancel(ctx)

	// Start goroutine to monitor for cancellation
	go func() {
		defer taskCancel()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-taskCtx.Done():
				return
			case <-ticker.C:
				var status string
				if err := service.db.GORM.Model(&models.TaskV2{}).Where("id = ?", task.ID).
					Select("status").Scan(&status).Error; err != nil {
					continue
				}
				if status == "cancelling" {
					return
				}
			}
		}
	}()

	return taskCtx
}

// finalizeTaskExecution updates final task status based on execution result
func (pool *TaskWorkerPool) finalizeTaskExecution(
	service *TaskService, task *models.TaskV2, taskErr error, startTime time.Time,
) {
	endTime := time.Now()
	logger := pool.logger.With("taskId", task.ID)

	if taskErr != nil {
		pool.handleTaskFailure(service, task, taskErr, endTime, logger)
	} else {
		pool.handleTaskSuccess(service, task, endTime, startTime, logger)
	}
}

// handleTaskFailure handles task failure scenarios
func (pool *TaskWorkerPool) handleTaskFailure(
	service *TaskService, task *models.TaskV2, taskErr error,
	endTime time.Time, logger *zap.SugaredLogger,
) {
	if errors.Is(taskErr, context.Canceled) {
		if err := service.updateTaskStatus(task.ID, "aborted", "Task was cancelled", &endTime); err != nil {
			logger.Errorw("Failed to update task status to aborted", "taskId", task.ID, "error", err)
		}
		logger.Infow("Task was cancelled")
	} else {
		if err := service.updateTaskStatus(task.ID, "failed", taskErr.Error(), &endTime); err != nil {
			logger.Errorw("Failed to update task status to failed", "taskId", task.ID, "error", err)
		}
		logger.Errorw("Task failed", "error", taskErr)
	}
}

// handleTaskSuccess handles successful task completion
func (pool *TaskWorkerPool) handleTaskSuccess(
	service *TaskService, task *models.TaskV2, endTime, startTime time.Time,
	logger *zap.SugaredLogger,
) {
	if err := service.updateTaskStatus(
		task.ID, "completed", "Task completed successfully", &endTime,
	); err != nil {
		logger.Errorw("Failed to update task status to completed", "taskId", task.ID, "error", err)
	}
	logger.Infow("Task completed successfully", "duration", endTime.Sub(startTime))
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
func (ts *TaskService) updateTaskStatus(
	taskID int,
	status string,
	message string,
	timestamp *time.Time,
) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if message != "" {
		if status == "failed" {
			updates["error_message"] = message
		}
	}

	if timestamp != nil {
		switch status {
		case "started":
			updates["started_at"] = *timestamp
		case "completed", "failed", "aborted":
			updates["ended_at"] = *timestamp
		case "queued":
			// Queued tasks don't need timestamp updates
		case "cancelling":
			// Cancelling tasks don't need timestamp updates
		}
	}

	return ts.db.GORM.Model(&models.TaskV2{}).Where("id = ?", taskID).Updates(updates).Error
}

// updateTaskProgress is a no-op in V2 since progress tracking is simplified
func (ts *TaskService) updateTaskProgress(taskID int, percent int, message string) error {
	// TaskV2 doesn't have complex progress tracking, so this is a no-op
	// Progress information can be stored in the Result field if needed
	return nil
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
	var scheduledTasks []*models.ScheduledTaskV2
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
