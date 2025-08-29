package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthIssueService_CreateIssue(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create a test health issue
	issue := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Database Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Database connection slow",
	}

	// Create the issue
	err := service.CreateIssue(issue)
	require.NoError(t, err)

	// Verify the issue was created
	assert.NotZero(t, issue.ID)
	assert.False(t, issue.FirstSeen.IsZero())
	assert.False(t, issue.LastSeen.IsZero())

	// Verify it's in the database
	var dbIssue models.HealthIssue
	err = db.GORM.First(&dbIssue, issue.ID).Error
	require.NoError(t, err)
	assert.Equal(t, issue.Type, dbIssue.Type)
	assert.Equal(t, issue.Source, dbIssue.Source)
	assert.Equal(t, issue.Severity, dbIssue.Severity)
	assert.Equal(t, issue.Message, dbIssue.Message)
}

func TestHealthIssueService_CreateDuplicateIssue(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create initial issue
	issue1 := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Database Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Database connection slow",
	}
	err := service.CreateIssue(issue1)
	require.NoError(t, err)
	originalID := issue1.ID
	originalFirstSeen := issue1.FirstSeen

	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Try to create same issue again
	issue2 := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Database Checker",
		Severity: models.HealthSeverityError, // Different severity
		Message:  "Database connection slow", // Same message
	}
	err = service.CreateIssue(issue2)
	require.NoError(t, err)

	// Verify it updated the existing issue instead of creating new one
	var dbIssue models.HealthIssue
	err = db.GORM.First(&dbIssue, originalID).Error
	require.NoError(t, err)
	assert.Equal(t, models.HealthSeverityError, dbIssue.Severity)       // Should be updated
	assert.Equal(t, originalFirstSeen.Unix(), dbIssue.FirstSeen.Unix()) // FirstSeen should remain
	assert.True(t, dbIssue.LastSeen.After(originalFirstSeen))           // LastSeen should be updated

	// Verify only one issue exists
	var count int64
	db.GORM.Model(&models.HealthIssue{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestHealthIssueService_GetIssues(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create test issues and run tests
	testGetIssuesWithTestData(t, service)
}

// testGetIssuesWithTestData creates test issues and runs get issues tests
func testGetIssuesWithTestData(t *testing.T, service *HealthIssueService) {
	issues := createTestHealthIssues()

	for _, issue := range issues {
		err := service.CreateIssue(issue)
		require.NoError(t, err)
	}

	testGetAllIssues(t, service)
	testGetIssuesFiltering(t, service)
	testGetIssuesPagination(t, service)
}

// createTestHealthIssues creates a set of test health issues
func createTestHealthIssues() []*models.HealthIssue {
	return []*models.HealthIssue{
		{
			Type:     models.HealthCheckTypeDatabase,
			Source:   "Database Checker",
			Severity: models.HealthSeverityCritical,
			Message:  "Database down",
		},
		{
			Type:     models.HealthCheckTypeDiskSpace,
			Source:   "Disk Space Checker",
			Severity: models.HealthSeverityWarning,
			Message:  "Low disk space",
		},
		{
			Type:       models.HealthCheckTypeSystem,
			Source:     "System Checker",
			Severity:   models.HealthSeverityInfo,
			Message:    "High memory usage",
			IsResolved: true,
		},
	}
}

// testGetAllIssues tests getting all issues without filters
func testGetAllIssues(t *testing.T, service *HealthIssueService) {
	allIssues, total, err := service.GetIssues(models.HealthIssueFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, allIssues, 3)
}

// testGetIssuesFiltering tests filtering issues by various criteria
func testGetIssuesFiltering(t *testing.T, service *HealthIssueService) {
	// Test filtering by type
	dbIssues, total, err := service.GetIssues(models.HealthIssueFilter{
		Types: []models.HealthCheckType{models.HealthCheckTypeDatabase},
	}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, dbIssues, 1)
	assert.Equal(t, models.HealthCheckTypeDatabase, dbIssues[0].Type)

	// Test filtering by severity
	criticalIssues, total, err := service.GetIssues(models.HealthIssueFilter{
		Severities: []models.HealthSeverity{models.HealthSeverityCritical},
	}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, criticalIssues, 1)
	assert.Equal(t, models.HealthSeverityCritical, criticalIssues[0].Severity)

	// Test filtering by resolved status
	unresolvedIssues, total, err := service.GetIssues(models.HealthIssueFilter{
		Resolved: &[]bool{false}[0],
	}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, unresolvedIssues, 2)
}

// testGetIssuesPagination tests pagination functionality
func testGetIssuesPagination(t *testing.T, service *HealthIssueService) {
	page1, total, err := service.GetIssues(models.HealthIssueFilter{}, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, page1, 2)

	page2, total, err := service.GetIssues(models.HealthIssueFilter{}, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, page2, 1)
}

func TestHealthIssueService_GetIssueByID(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create a test issue
	issue := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Test issue",
	}
	err := service.CreateIssue(issue)
	require.NoError(t, err)

	// Get issue by ID
	retrievedIssue, err := service.GetIssueByID(issue.ID)
	require.NoError(t, err)
	assert.Equal(t, issue.ID, retrievedIssue.ID)
	assert.Equal(t, issue.Type, retrievedIssue.Type)
	assert.Equal(t, issue.Source, retrievedIssue.Source)
	assert.Equal(t, issue.Severity, retrievedIssue.Severity)
	assert.Equal(t, issue.Message, retrievedIssue.Message)

	// Try to get non-existent issue
	_, err = service.GetIssueByID(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthIssueService_ResolveIssue(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create a test issue
	issue := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Test issue",
	}
	err := service.CreateIssue(issue)
	require.NoError(t, err)

	// Resolve the issue
	err = service.ResolveIssue(issue.ID)
	require.NoError(t, err)

	// Verify it was resolved
	resolvedIssue, err := service.GetIssueByID(issue.ID)
	require.NoError(t, err)
	assert.True(t, resolvedIssue.IsResolved)
	assert.NotNil(t, resolvedIssue.ResolvedAt)

	// Try to resolve non-existent issue
	err = service.ResolveIssue(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthIssueService_DismissIssue(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create a test issue
	issue := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Test issue",
	}
	err := service.CreateIssue(issue)
	require.NoError(t, err)

	// Dismiss the issue
	err = service.DismissIssue(issue.ID)
	require.NoError(t, err)

	// Verify it was dismissed
	dismissedIssue, err := service.GetIssueByID(issue.ID)
	require.NoError(t, err)
	assert.True(t, dismissedIssue.IsDismissed)

	// Try to dismiss non-existent issue
	err = service.DismissIssue(99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthIssueService_CheckForDuplicates(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create a test issue
	issue1 := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Test Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Test issue",
	}
	err := service.CreateIssue(issue1)
	require.NoError(t, err)

	// Check for duplicate of the same issue
	duplicate, err := service.CheckForDuplicates(issue1)
	require.NoError(t, err)
	assert.NotNil(t, duplicate)
	assert.Equal(t, issue1.ID, duplicate.ID)

	// Check for duplicate of a different issue
	issue2 := &models.HealthIssue{
		Type:     models.HealthCheckTypeDatabase,
		Source:   "Different Checker",
		Severity: models.HealthSeverityWarning,
		Message:  "Different issue",
	}
	duplicate, err = service.CheckForDuplicates(issue2)
	require.NoError(t, err)
	assert.Nil(t, duplicate) // Should be no duplicate

	// Resolve the first issue and check again
	err = service.ResolveIssue(issue1.ID)
	require.NoError(t, err)

	duplicate, err = service.CheckForDuplicates(issue1)
	require.NoError(t, err)
	assert.Nil(t, duplicate) // Should be no duplicate since it's resolved
}

func TestHealthIssueService_CleanupResolvedIssues(t *testing.T) {
	db, log := setupTestDB(t)
	service := NewHealthIssueService(db, log)

	// Create test issues
	oldIssue := &models.HealthIssue{
		Type:       models.HealthCheckTypeDatabase,
		Source:     "Test Checker",
		Severity:   models.HealthSeverityWarning,
		Message:    "Old issue",
		IsResolved: true,
	}
	err := service.CreateIssue(oldIssue)
	require.NoError(t, err)

	// Manually set resolved_at to an old date
	oldTime := time.Now().AddDate(0, 0, -60) // 60 days ago
	db.GORM.Model(oldIssue).Update("resolved_at", oldTime)

	newIssue := &models.HealthIssue{
		Type:       models.HealthCheckTypeDiskSpace,
		Source:     "Disk Checker",
		Severity:   models.HealthSeverityInfo,
		Message:    "New issue",
		IsResolved: true,
	}
	err = service.CreateIssue(newIssue)
	require.NoError(t, err)

	// Manually set resolved_at to a recent date
	recentTime := time.Now().AddDate(0, 0, -1) // 1 day ago
	db.GORM.Model(newIssue).Update("resolved_at", recentTime)

	unresolvedIssue := &models.HealthIssue{
		Type:     models.HealthCheckTypeSystem,
		Source:   "System Checker",
		Severity: models.HealthSeverityError,
		Message:  "Unresolved issue",
	}
	err = service.CreateIssue(unresolvedIssue)
	require.NoError(t, err)

	// Count issues before cleanup
	var beforeCount int64
	db.GORM.Model(&models.HealthIssue{}).Count(&beforeCount)
	assert.Equal(t, int64(3), beforeCount)

	// Cleanup issues older than 30 days
	cutoff := time.Now().AddDate(0, 0, -30)
	err = service.CleanupResolvedIssues(cutoff)
	require.NoError(t, err)

	// Count issues after cleanup
	var afterCount int64
	db.GORM.Model(&models.HealthIssue{}).Count(&afterCount)
	assert.Equal(t, int64(2), afterCount) // Old resolved issue should be removed

	// Verify the correct issue was removed
	_, err = service.GetIssueByID(oldIssue.ID)
	assert.Error(t, err) // Should not be found

	// Verify other issues still exist
	_, err = service.GetIssueByID(newIssue.ID)
	assert.NoError(t, err) // Should still exist

	_, err = service.GetIssueByID(unresolvedIssue.ID)
	assert.NoError(t, err) // Should still exist
}
