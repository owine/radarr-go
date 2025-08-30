package services

import (
	"context"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthService_NewHealthService(t *testing.T) {
	// Test creating a new health service
	db, log := setupTestDB(t)
	cfg := &config.Config{
		Health: config.HealthConfig{
			Enabled:                    true,
			Interval:                   "5m",
			DiskSpaceWarningThreshold:  5 * 1024 * 1024 * 1024,
			DiskSpaceCriticalThreshold: 1 * 1024 * 1024 * 1024,
			DatabaseTimeoutThreshold:   "10s",
			ExternalServiceTimeout:     "15s",
			MetricsRetentionDays:       7,
			NotifyCriticalIssues:       true,
			NotifyWarningIssues:        false,
		},
	}

	healthService := NewHealthService(db, cfg, log)

	assert.NotNil(t, healthService)
	assert.Equal(t, cfg.Health.Enabled, healthService.healthConfig.Enabled)
	assert.Equal(t, 5*time.Minute, healthService.healthConfig.Interval)
	assert.Equal(t, cfg.Health.DiskSpaceWarningThreshold, healthService.healthConfig.DiskSpaceWarningThreshold)
	assert.Equal(t, cfg.Health.DiskSpaceCriticalThreshold, healthService.healthConfig.DiskSpaceCriticalThreshold)
	assert.Equal(t, 10*time.Second, healthService.healthConfig.DatabaseTimeoutThreshold)
	assert.Equal(t, 15*time.Second, healthService.healthConfig.ExternalServiceTimeout)
	assert.Equal(t, cfg.Health.MetricsRetentionDays, healthService.healthConfig.MetricsRetentionDays)
	assert.Equal(t, cfg.Health.NotifyCriticalIssues, healthService.healthConfig.NotifyCriticalIssues)
	assert.Equal(t, cfg.Health.NotifyWarningIssues, healthService.healthConfig.NotifyWarningIssues)
}

func TestHealthService_RegisterChecker(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Create a mock checker
	mockChecker := &MockHealthChecker{
		name:      "Test Checker",
		checkType: models.HealthCheckTypeSystem,
		enabled:   true,
		interval:  1 * time.Minute,
	}

	// Register the checker
	healthService.RegisterChecker(mockChecker)

	// Verify it was registered
	assert.Contains(t, healthService.checkers, mockChecker.Name())
	assert.Contains(t, healthService.checkerTypes[string(models.HealthCheckTypeSystem)], mockChecker.Name())
}

func TestHealthService_UnregisterChecker(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Create and register a mock checker
	mockChecker := &MockHealthChecker{
		name:      "Test Checker",
		checkType: models.HealthCheckTypeSystem,
		enabled:   true,
		interval:  1 * time.Minute,
	}
	healthService.RegisterChecker(mockChecker)

	// Verify it was registered
	assert.Contains(t, healthService.checkers, mockChecker.Name())

	// Unregister the checker
	healthService.UnregisterChecker(mockChecker.Name())

	// Verify it was removed
	assert.NotContains(t, healthService.checkers, mockChecker.Name())
	assert.NotContains(t, healthService.checkerTypes[string(models.HealthCheckTypeSystem)], mockChecker.Name())
}

func TestHealthService_RunAllChecks(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Create mock checkers with different results
	healthyChecker := &MockHealthChecker{
		name:      "Healthy Checker",
		checkType: models.HealthCheckTypeDatabase,
		enabled:   true,
		interval:  1 * time.Minute,
		result: models.HealthCheckExecution{
			Type:      models.HealthCheckTypeDatabase,
			Source:    "Healthy Checker",
			Status:    models.HealthStatusHealthy,
			Message:   "All good",
			Timestamp: time.Now(),
		},
	}

	warningChecker := &MockHealthChecker{
		name:      "Warning Checker",
		checkType: models.HealthCheckTypeSystem,
		enabled:   true,
		interval:  1 * time.Minute,
		result: models.HealthCheckExecution{
			Type:      models.HealthCheckTypeSystem,
			Source:    "Warning Checker",
			Status:    models.HealthStatusWarning,
			Message:   "Warning detected",
			Timestamp: time.Now(),
			Issues: []models.HealthIssue{
				{
					Type:     models.HealthCheckTypeSystem,
					Source:   "Warning Checker",
					Severity: models.HealthSeverityWarning,
					Message:  "Warning issue",
				},
			},
		},
	}

	healthService.RegisterChecker(healthyChecker)
	healthService.RegisterChecker(warningChecker)

	// Run all checks
	ctx := context.Background()
	result := healthService.RunAllChecks(ctx, nil)

	// Verify results
	assert.Equal(t, models.HealthStatusWarning, result.OverallStatus) // Worst status
	assert.Len(t, result.Issues, 2)
	assert.Equal(t, 2, result.Summary.Total)
	assert.Equal(t, 1, result.Summary.Healthy)
	assert.Equal(t, 1, result.Summary.Warning)
	assert.Len(t, result.Issues, 1)
}

func TestHealthService_RunCheck(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Create and register a mock checker
	mockChecker := &MockHealthChecker{
		name:      "Test Checker",
		checkType: models.HealthCheckTypeDatabase,
		enabled:   true,
		interval:  1 * time.Minute,
		result: models.HealthCheckExecution{
			Type:      models.HealthCheckTypeDatabase,
			Source:    "Test Checker",
			Status:    models.HealthStatusHealthy,
			Message:   "Database is healthy",
			Timestamp: time.Now(),
		},
	}
	healthService.RegisterChecker(mockChecker)

	// Run specific check
	ctx := context.Background()
	result, err := healthService.RunCheck(ctx, "Test Checker")

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, models.HealthCheckTypeDatabase, result.Type)
	assert.Equal(t, "Test Checker", result.Source)
	assert.Equal(t, models.HealthStatusHealthy, result.Status)
	assert.Equal(t, "Database is healthy", result.Message)
}

func TestHealthService_RunCheck_NotFound(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Try to run a non-existent check
	ctx := context.Background()
	result, err := healthService.RunCheck(ctx, "Non-existent Checker")

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthService_RunCheck_Disabled(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Create and register a disabled mock checker
	mockChecker := &MockHealthChecker{
		name:      "Disabled Checker",
		checkType: models.HealthCheckTypeSystem,
		enabled:   false, // Disabled
		interval:  1 * time.Minute,
	}
	healthService.RegisterChecker(mockChecker)

	// Try to run the disabled check
	ctx := context.Background()
	result, err := healthService.RunCheck(ctx, "Disabled Checker")

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "is disabled")
}

func TestHealthService_GetHealthStatus(t *testing.T) {
	db, log := setupTestDB(t)
	cfg := &config.Config{}
	healthService := NewHealthService(db, cfg, log)

	// Mock a checker
	mockChecker := &MockHealthChecker{
		name:      "Test Checker",
		checkType: models.HealthCheckTypeDatabase,
		enabled:   true,
		interval:  1 * time.Minute,
		result: models.HealthCheckExecution{
			Type:      models.HealthCheckTypeDatabase,
			Source:    "Test Checker",
			Status:    models.HealthStatusHealthy,
			Message:   "Database is healthy",
			Timestamp: time.Now(),
		},
	}
	healthService.RegisterChecker(mockChecker)

	// Get health status
	ctx := context.Background()
	status := healthService.GetHealthStatus(ctx)

	// Verify status
	assert.Equal(t, models.HealthStatusHealthy, status)
}

// MockHealthChecker is a mock implementation of HealthChecker for testing
type MockHealthChecker struct {
	name      string
	checkType models.HealthCheckType
	enabled   bool
	interval  time.Duration
	result    models.HealthCheckExecution
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

func (m *MockHealthChecker) Type() models.HealthCheckType {
	return m.checkType
}

func (m *MockHealthChecker) Check(_ context.Context) models.HealthCheckExecution {
	return m.result
}

func (m *MockHealthChecker) IsEnabled() bool {
	return m.enabled
}

func (m *MockHealthChecker) GetInterval() time.Duration {
	return m.interval
}
