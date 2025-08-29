package services

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// HealthIssueService implements health issue management
type HealthIssueService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewHealthIssueService creates a new health issue service
func NewHealthIssueService(db *database.Database, logger *logger.Logger) *HealthIssueService {
	return &HealthIssueService{
		db:     db,
		logger: logger,
	}
}

// CreateIssue implements HealthIssueServiceInterface
func (his *HealthIssueService) CreateIssue(issue *models.HealthIssue) error {
	// Check for duplicate issues first
	existing, err := his.CheckForDuplicates(issue)
	if err != nil {
		his.logger.Warnw("Failed to check for duplicate issues", "error", err)
	}

	if existing != nil {
		// Update existing issue instead of creating duplicate
		existing.LastSeen = time.Now()
		existing.Severity = issue.Severity // Update severity if changed
		existing.Details = issue.Details   // Update details
		existing.Data = issue.Data         // Update additional data
		return his.UpdateIssue(existing)
	}

	// Set timestamps for new issue
	now := time.Now()
	if issue.FirstSeen.IsZero() {
		issue.FirstSeen = now
	}
	if issue.LastSeen.IsZero() {
		issue.LastSeen = now
	}

	if err := his.db.GORM.Create(issue).Error; err != nil {
		his.logger.Errorw("Failed to create health issue", "error", err, "type", issue.Type, "source", issue.Source)
		return fmt.Errorf("failed to create health issue: %w", err)
	}

	his.logger.Infow("Health issue created",
		"id", issue.ID,
		"type", issue.Type,
		"source", issue.Source,
		"severity", issue.Severity,
		"message", issue.Message)

	return nil
}

// UpdateIssue implements HealthIssueServiceInterface
func (his *HealthIssueService) UpdateIssue(issue *models.HealthIssue) error {
	if err := his.db.GORM.Save(issue).Error; err != nil {
		his.logger.Errorw("Failed to update health issue", "error", err, "id", issue.ID)
		return fmt.Errorf("failed to update health issue: %w", err)
	}

	his.logger.Debugw("Health issue updated", "id", issue.ID)
	return nil
}

// GetIssues implements HealthIssueServiceInterface
func (his *HealthIssueService) GetIssues(filter models.HealthIssueFilter, limit, offset int) ([]models.HealthIssue, int64, error) {
	query := his.db.GORM.Model(&models.HealthIssue{})

	// Apply filters
	query = his.applyFilters(query, filter)

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		his.logger.Errorw("Failed to count health issues", "error", err)
		return nil, 0, fmt.Errorf("failed to count health issues: %w", err)
	}

	// Get paginated results
	var issues []models.HealthIssue
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&issues).Error; err != nil {
		his.logger.Errorw("Failed to fetch health issues", "error", err)
		return nil, 0, fmt.Errorf("failed to fetch health issues: %w", err)
	}

	return issues, total, nil
}

// GetIssueByID implements HealthIssueServiceInterface
func (his *HealthIssueService) GetIssueByID(id int) (*models.HealthIssue, error) {
	var issue models.HealthIssue
	if err := his.db.GORM.First(&issue, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("health issue not found")
		}
		his.logger.Errorw("Failed to fetch health issue", "error", err, "id", id)
		return nil, fmt.Errorf("failed to fetch health issue: %w", err)
	}
	return &issue, nil
}

// ResolveIssue implements HealthIssueServiceInterface
func (his *HealthIssueService) ResolveIssue(id int) error {
	now := time.Now()
	result := his.db.GORM.Model(&models.HealthIssue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_resolved": true,
		"resolved_at": now,
	})

	if result.Error != nil {
		his.logger.Errorw("Failed to resolve health issue", "error", result.Error, "id", id)
		return fmt.Errorf("failed to resolve health issue: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("health issue not found")
	}

	his.logger.Infow("Health issue resolved", "id", id)
	return nil
}

// DismissIssue implements HealthIssueServiceInterface
func (his *HealthIssueService) DismissIssue(id int) error {
	result := his.db.GORM.Model(&models.HealthIssue{}).Where("id = ?", id).Update("is_dismissed", true)

	if result.Error != nil {
		his.logger.Errorw("Failed to dismiss health issue", "error", result.Error, "id", id)
		return fmt.Errorf("failed to dismiss health issue: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("health issue not found")
	}

	his.logger.Infow("Health issue dismissed", "id", id)
	return nil
}

// CheckForDuplicates implements HealthIssueServiceInterface
func (his *HealthIssueService) CheckForDuplicates(issue *models.HealthIssue) (*models.HealthIssue, error) {
	var existing models.HealthIssue

	// Look for unresolved issues with same type, source, and message
	err := his.db.GORM.Where(
		"type = ? AND source = ? AND message = ? AND is_resolved = false",
		issue.Type, issue.Source, issue.Message,
	).First(&existing).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No duplicate found
		}
		return nil, fmt.Errorf("failed to check for duplicate issues: %w", err)
	}

	return &existing, nil
}

// CleanupResolvedIssues implements HealthIssueServiceInterface
func (his *HealthIssueService) CleanupResolvedIssues(olderThan time.Time) error {
	result := his.db.GORM.Where("is_resolved = true AND resolved_at < ?", olderThan).Delete(&models.HealthIssue{})

	if result.Error != nil {
		his.logger.Errorw("Failed to cleanup resolved issues", "error", result.Error)
		return fmt.Errorf("failed to cleanup resolved issues: %w", result.Error)
	}

	his.logger.Infow("Cleaned up resolved health issues", "count", result.RowsAffected, "older_than", olderThan)
	return nil
}

// GetIssueHistory implements HealthIssueServiceInterface
func (his *HealthIssueService) GetIssueHistory(since time.Time, groupBy string) (map[string]int, error) {
	var results []struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}

	query := his.db.GORM.Model(&models.HealthIssue{}).
		Select("? as category, count(*) as count", gorm.Expr(groupBy)).
		Where("created_at >= ?", since).
		Group(groupBy)

	if err := query.Scan(&results).Error; err != nil {
		his.logger.Errorw("Failed to get issue history", "error", err, "since", since, "groupBy", groupBy)
		return nil, fmt.Errorf("failed to get issue history: %w", err)
	}

	history := make(map[string]int)
	for _, result := range results {
		history[result.Category] = result.Count
	}

	return history, nil
}

// applyFilters applies the provided filters to a GORM query
func (his *HealthIssueService) applyFilters(query *gorm.DB, filter models.HealthIssueFilter) *gorm.DB {
	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}

	if len(filter.Severities) > 0 {
		query = query.Where("severity IN ?", filter.Severities)
	}

	if len(filter.Sources) > 0 {
		query = query.Where("source IN ?", filter.Sources)
	}

	if filter.Resolved != nil {
		query = query.Where("is_resolved = ?", *filter.Resolved)
	}

	if filter.Dismissed != nil {
		query = query.Where("is_dismissed = ?", *filter.Dismissed)
	}

	if filter.Since != nil {
		query = query.Where("created_at >= ?", *filter.Since)
	}

	if filter.Until != nil {
		query = query.Where("created_at <= ?", *filter.Until)
	}

	return query
}

// GetActiveIssues returns currently active (unresolved, non-dismissed) issues
func (his *HealthIssueService) GetActiveIssues(limit, offset int) ([]models.HealthIssue, int64, error) {
	filter := models.HealthIssueFilter{
		Resolved:  boolPointer(false),
		Dismissed: boolPointer(false),
	}

	return his.GetIssues(filter, limit, offset)
}

// GetIssuesBySeverity returns issues filtered by severity level
func (his *HealthIssueService) GetIssuesBySeverity(severity models.HealthSeverity, limit, offset int) ([]models.HealthIssue, int64, error) {
	filter := models.HealthIssueFilter{
		Severities: []models.HealthSeverity{severity},
		Resolved:   boolPointer(false),
	}

	return his.GetIssues(filter, limit, offset)
}

// GetCriticalIssues returns all critical issues that are unresolved
func (his *HealthIssueService) GetCriticalIssues(limit, offset int) ([]models.HealthIssue, int64, error) {
	return his.GetIssuesBySeverity(models.HealthSeverityCritical, limit, offset)
}

// GetIssuesStats returns statistics about health issues
func (his *HealthIssueService) GetIssuesStats() (*HealthIssuesStats, error) {
	var stats HealthIssuesStats

	if err := his.collectBasicStats(&stats); err != nil {
		return nil, err
	}

	if err := his.collectSeverityStats(&stats); err != nil {
		return nil, err
	}

	if err := his.collectTypeStats(&stats); err != nil {
		return nil, err
	}

	if err := his.collectTimeBasedStats(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// collectBasicStats collects total and active issue counts
func (his *HealthIssueService) collectBasicStats(stats *HealthIssuesStats) error {
	// Total issues
	if err := his.db.GORM.Model(&models.HealthIssue{}).Count(&stats.Total).Error; err != nil {
		return fmt.Errorf("failed to count total issues: %w", err)
	}

	// Active issues
	if err := his.db.GORM.Model(&models.HealthIssue{}).
		Where("is_resolved = false AND is_dismissed = false").
		Count(&stats.Active).Error; err != nil {
		return fmt.Errorf("failed to count active issues: %w", err)
	}

	return nil
}

// collectSeverityStats collects issues grouped by severity
func (his *HealthIssueService) collectSeverityStats(stats *HealthIssuesStats) error {
	var severityStats []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}

	if err := his.db.GORM.Model(&models.HealthIssue{}).
		Select("severity, count(*) as count").
		Where("is_resolved = false AND is_dismissed = false").
		Group("severity").
		Scan(&severityStats).Error; err != nil {
		return fmt.Errorf("failed to count issues by severity: %w", err)
	}

	stats.BySeverity = make(map[models.HealthSeverity]int64)
	for _, stat := range severityStats {
		stats.BySeverity[models.HealthSeverity(stat.Severity)] = stat.Count
	}

	return nil
}

// collectTypeStats collects issues grouped by type
func (his *HealthIssueService) collectTypeStats(stats *HealthIssuesStats) error {
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	if err := his.db.GORM.Model(&models.HealthIssue{}).
		Select("type, count(*) as count").
		Where("is_resolved = false AND is_dismissed = false").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return fmt.Errorf("failed to count issues by type: %w", err)
	}

	stats.ByType = make(map[models.HealthCheckType]int64)
	for _, stat := range typeStats {
		stats.ByType[models.HealthCheckType(stat.Type)] = stat.Count
	}

	return nil
}

// collectTimeBasedStats collects recent and resolved issue counts
func (his *HealthIssueService) collectTimeBasedStats(stats *HealthIssuesStats) error {
	// Recent issues (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	if err := his.db.GORM.Model(&models.HealthIssue{}).
		Where("created_at >= ?", since).
		Count(&stats.Recent24h).Error; err != nil {
		return fmt.Errorf("failed to count recent issues: %w", err)
	}

	// Resolved issues (last 30 days)
	since30d := time.Now().Add(-30 * 24 * time.Hour)
	if err := his.db.GORM.Model(&models.HealthIssue{}).
		Where("is_resolved = true AND resolved_at >= ?", since30d).
		Count(&stats.Resolved30d).Error; err != nil {
		return fmt.Errorf("failed to count resolved issues: %w", err)
	}

	return nil
}

// HealthIssuesStats represents statistics about health issues
type HealthIssuesStats struct {
	Total       int64                            `json:"total"`
	Active      int64                            `json:"active"`
	BySeverity  map[models.HealthSeverity]int64  `json:"bySeverity"`
	ByType      map[models.HealthCheckType]int64 `json:"byType"`
	Recent24h   int64                            `json:"recent24h"`
	Resolved30d int64                            `json:"resolved30d"`
}

// boolPointer returns a pointer to a bool value
func boolPointer(b bool) *bool {
	return &b
}
