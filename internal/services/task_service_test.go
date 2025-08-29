package services

import (
	"context"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTaskServiceForTesting sets up the task service with required migrations
func setupTaskServiceForTesting(t *testing.T, db *database.Database, logger *logger.Logger) *TaskService {
	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	return NewTaskService(db, logger)
}

// createTestTasks creates standard test tasks for testing
func createTestTasks(t *testing.T, db *database.Database) {
	tasks := []*models.Task{
		{
			Name:        "Task 1",
			CommandName: "Command1",
			Status:      models.TaskStatusQueued,
			Priority:    models.TaskPriorityHigh,
			Trigger:     models.TaskTriggerAPI,
			QueuedAt:    time.Now(),
			Body:        models.TaskBody{},
			Progress:    models.TaskProgress{},
		},
		{
			Name:        "Task 2",
			CommandName: "Command2",
			Status:      models.TaskStatusCompleted,
			Priority:    models.TaskPriorityNormal,
			Trigger:     models.TaskTriggerScheduled,
			QueuedAt:    time.Now(),
			Body:        models.TaskBody{},
			Progress:    models.TaskProgress{},
		},
		{
			Name:        "Task 3",
			CommandName: "Command1",
			Status:      models.TaskStatusFailed,
			Priority:    models.TaskPriorityLow,
			Trigger:     models.TaskTriggerManual,
			QueuedAt:    time.Now(),
			Body:        models.TaskBody{},
			Progress:    models.TaskProgress{},
		},
	}

	for _, task := range tasks {
		err := db.GORM.Create(task).Error
		require.NoError(t, err)
	}
}

// testListAllTasks tests listing all tasks without filters
func testListAllTasks(t *testing.T, service *TaskService) {
	allTasks, total, err := service.ListTasks("", "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, allTasks, 3)
}

// testFilterByStatus tests filtering tasks by status
func testFilterByStatus(t *testing.T, service *TaskService) {
	queuedTasks, total, err := service.ListTasks(models.TaskStatusQueued, "", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, queuedTasks, 1)
	assert.Equal(t, "Task 1", queuedTasks[0].Name)
}

// testFilterByCommand tests filtering tasks by command name
func testFilterByCommand(t *testing.T, service *TaskService) {
	command1Tasks, total, err := service.ListTasks("", "Command1", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, command1Tasks, 2)
}

// testTaskPagination tests task list pagination
func testTaskPagination(t *testing.T, service *TaskService) {
	paginatedTasks, total, err := service.ListTasks("", "", 2, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, paginatedTasks, 2)
}

// TestTaskHandler is a mock task handler for testing
type TestTaskHandler struct {
	name        string
	description string
	executed    bool
	shouldFail  bool
	delay       time.Duration
}

func NewTestTaskHandler(name, description string) *TestTaskHandler {
	return &TestTaskHandler{
		name:        name,
		description: description,
	}
}

func (h *TestTaskHandler) Execute(
	ctx context.Context, _ *models.Task, updateProgress func(percent int, message string),
) error {
	h.executed = true

	updateProgress(0, "Starting test task")

	if h.delay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(h.delay):
		}
	}

	updateProgress(50, "Halfway done")

	if h.shouldFail {
		return assert.AnError
	}

	updateProgress(100, "Completed test task")
	return nil
}

func (h *TestTaskHandler) GetName() string {
	return h.name
}

func (h *TestTaskHandler) GetDescription() string {
	return h.description
}

func TestTaskService_QueueTask(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Register test handler
	handler := NewTestTaskHandler("TestCommand", "Test handler description")
	service.RegisterHandler(handler)

	// Queue a task
	task, err := service.QueueTask(
		"Test Task",
		"TestCommand",
		models.TaskBody{"key": "value"},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)

	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Greater(t, task.ID, 0)
	assert.Equal(t, "Test Task", task.Name)
	assert.Equal(t, "TestCommand", task.CommandName)
	assert.Equal(t, models.TaskStatusQueued, task.Status)
	assert.Equal(t, models.TaskPriorityNormal, task.Priority)
	assert.Equal(t, models.TaskTriggerAPI, task.Trigger)
	assert.NotZero(t, task.QueuedAt)

	// Give task time to execute
	time.Sleep(100 * time.Millisecond)

	// Verify task was executed
	updatedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.True(t, handler.executed)
	assert.Equal(t, models.TaskStatusCompleted, updatedTask.Status)
}

func TestTaskService_GetTask(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Create a task directly in database
	task := &models.Task{
		Name:        "Test Task",
		CommandName: "TestCommand",
		Status:      models.TaskStatusQueued,
		Priority:    models.TaskPriorityNormal,
		Trigger:     models.TaskTriggerManual,
		QueuedAt:    time.Now(),
		Body:        models.TaskBody{"test": "data"},
		Progress:    models.TaskProgress{},
	}

	err = db.GORM.Create(task).Error
	require.NoError(t, err)

	// Retrieve task
	retrievedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.Name, retrievedTask.Name)
	assert.Equal(t, task.CommandName, retrievedTask.CommandName)
}

func TestTaskService_ListTasks(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := setupTaskServiceForTesting(t, db, logger)
	defer service.Shutdown()

	createTestTasks(t, db)

	// Test listing all tasks
	testListAllTasks(t, service)

	// Test filtering by status
	testFilterByStatus(t, service)

	// Test filtering by command
	testFilterByCommand(t, service)

	// Test pagination
	testTaskPagination(t, service)
}

func TestTaskService_CancelTask(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Create a long-running task handler
	handler := NewTestTaskHandler("SlowCommand", "Slow test handler")
	handler.delay = 5 * time.Second // Long delay
	service.RegisterHandler(handler)

	// Queue a task
	task, err := service.QueueTask(
		"Slow Task",
		"SlowCommand",
		models.TaskBody{},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	require.NoError(t, err)

	// Wait a bit for task to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the task
	err = service.CancelTask(task.ID)
	require.NoError(t, err)

	// Wait for cancellation to take effect
	time.Sleep(100 * time.Millisecond)

	// Verify task was cancelled
	updatedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.True(t, updatedTask.Status == models.TaskStatusAborted || updatedTask.Status == models.TaskStatusCancelling)
}

func TestTaskService_ScheduledTasks(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Create a scheduled task
	scheduledTask, err := service.CreateScheduledTask(
		"Test Scheduled Task",
		"TestCommand",
		models.TaskBody{"scheduled": true},
		1*time.Minute,
		models.TaskPriorityLow,
	)

	require.NoError(t, err)
	assert.NotNil(t, scheduledTask)
	assert.Greater(t, scheduledTask.ID, 0)
	assert.Equal(t, "Test Scheduled Task", scheduledTask.Name)
	assert.Equal(t, "TestCommand", scheduledTask.CommandName)
	assert.Equal(t, time.Minute, scheduledTask.Interval)
	assert.True(t, scheduledTask.Enabled)

	// Get scheduled tasks
	scheduledTasks, err := service.GetScheduledTasks()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(scheduledTasks), 1)

	// Update scheduled task
	updates := map[string]interface{}{
		"enabled":  false,
		"interval": 2 * time.Minute,
	}
	err = service.UpdateScheduledTask(scheduledTask.ID, updates)
	require.NoError(t, err)

	// Verify update
	updatedTasks, err := service.GetScheduledTasks()
	require.NoError(t, err)
	found := false
	for _, task := range updatedTasks {
		if task.ID == scheduledTask.ID {
			assert.False(t, task.Enabled)
			assert.Equal(t, 2*time.Minute, task.Interval)
			found = true
			break
		}
	}
	assert.True(t, found, "Updated scheduled task not found")

	// Delete scheduled task
	err = service.DeleteScheduledTask(scheduledTask.ID)
	require.NoError(t, err)

	// Verify deletion
	finalTasks, err := service.GetScheduledTasks()
	require.NoError(t, err)
	for _, task := range finalTasks {
		assert.NotEqual(t, scheduledTask.ID, task.ID)
	}
}

func TestTaskService_QueueStatus(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Get queue status
	status := service.GetQueueStatus()
	require.NotNil(t, status)

	// Verify expected worker pools exist
	assert.Contains(t, status, "default")
	assert.Contains(t, status, "high-priority")
	assert.Contains(t, status, "background")

	// Verify status structure
	defaultQueue, ok := status["default"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, defaultQueue, "maxWorkers")
	assert.Contains(t, defaultQueue, "activeWorkers")
	assert.Contains(t, defaultQueue, "queuedTasks")
	assert.Contains(t, defaultQueue, "activeTasks")
}

func TestTaskService_TaskProgress(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Register test handler
	handler := NewTestTaskHandler("ProgressCommand", "Progress test handler")
	service.RegisterHandler(handler)

	// Queue a task
	task, err := service.QueueTask(
		"Progress Task",
		"ProgressCommand",
		models.TaskBody{},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	require.NoError(t, err)

	// Wait for task to complete
	time.Sleep(100 * time.Millisecond)

	// Verify progress was updated
	updatedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.Greater(t, updatedTask.Progress.ProgressPercent, 0)
	assert.NotEmpty(t, updatedTask.Progress.StatusMessages)
	assert.Contains(t, updatedTask.Progress.StatusMessages, "Starting test task")
	assert.Contains(t, updatedTask.Progress.StatusMessages, "Completed test task")
}

func TestTaskService_TaskFailure(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Register failing test handler
	handler := NewTestTaskHandler("FailCommand", "Failing test handler")
	handler.shouldFail = true
	service.RegisterHandler(handler)

	// Queue a task
	task, err := service.QueueTask(
		"Failing Task",
		"FailCommand",
		models.TaskBody{},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	require.NoError(t, err)

	// Wait for task to complete
	time.Sleep(100 * time.Millisecond)

	// Verify task failed
	updatedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStatusFailed, updatedTask.Status)
	assert.NotEmpty(t, updatedTask.Exception)
}

func TestTaskService_UnknownHandler(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	// Auto-migrate task tables
	err := db.GORM.AutoMigrate(&models.Task{}, &models.ScheduledTask{}, &models.TaskQueue{})
	require.NoError(t, err)

	service := NewTaskService(db, logger)
	defer service.Shutdown()

	// Queue a task with unknown command
	task, err := service.QueueTask(
		"Unknown Task",
		"UnknownCommand",
		models.TaskBody{},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	require.NoError(t, err)

	// Wait for task processing
	time.Sleep(100 * time.Millisecond)

	// Verify task failed due to no handler
	updatedTask, err := service.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.TaskStatusFailed, updatedTask.Status)
	assert.Contains(t, updatedTask.Exception, "no handler registered")
}
