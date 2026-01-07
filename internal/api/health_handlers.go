//nolint:revive // "api" is a standard package name for API layers
package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/models"
)

const (
	healthIssueNotFoundError = "health issue not found"
)

// Health API Handlers

// handleGetHealth returns overall system health status
func (s *Server) handleGetHealth(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	forceRefresh := c.DefaultQuery("forceRefresh", falseBoolString) == trueBoolString
	includeIssues := c.DefaultQuery("includeIssues", trueBoolString) == trueBoolString

	var types []string
	if typeParam := c.Query("types"); typeParam != "" {
		// Parse comma-separated list of types
		// This would need proper parsing logic
		types = []string{typeParam}
	}

	if forceRefresh || s.services.HealthService == nil {
		// Run health checks
		result := s.services.HealthService.RunAllChecks(ctx, types)

		response := gin.H{
			"status":    result.OverallStatus,
			"summary":   result.Summary,
			"timestamp": result.Timestamp,
			"duration":  result.Duration.Milliseconds(),
		}

		if includeIssues {
			response["issues"] = result.Issues
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Return cached status
	status := s.services.HealthService.GetHealthStatus(ctx)
	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"cached":    true,
		"timestamp": time.Now(),
	})
}

// handleGetHealthCheck runs a specific health check
func (s *Server) handleGetHealthCheck(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Health check name is required"})
		return
	}

	ctx := c.Request.Context()
	result, err := s.services.HealthService.RunCheck(ctx, name)
	if err != nil {
		s.logger.Errorw("Failed to run health check", "name", name, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleGetHealthDashboard returns comprehensive health dashboard data
func (s *Server) handleGetHealthDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	dashboard, err := s.services.HealthService.GetHealthDashboard(ctx)
	if err != nil {
		s.logger.Errorw("Failed to get health dashboard", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve health dashboard"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// handleGetHealthIssues returns health issues with filtering and pagination
func (s *Server) handleGetHealthIssues(c *gin.Context) {
	page, pageSize := s.parseHealthIssuesPagination(c)
	filterV2 := s.parseHealthIssuesFilter(c)

	issuesV2, total, err := s.services.HealthService.GetHealthIssues(filterV2, pageSize, (page-1)*pageSize)
	if err != nil {
		s.logger.Errorw("Failed to get health issues", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve health issues"})
		return
	}

	// Convert V2 issues to V1 for API response
	issues := models.ConvertHealthIssueV2SliceToV1(issuesV2)

	response := s.buildHealthIssuesPaginatedResponse(issues, total, page, pageSize)
	c.JSON(http.StatusOK, response)
}

// parseHealthIssuesPagination parses pagination parameters from request
func (s *Server) parseHealthIssuesPagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	return page, pageSize
}

// parseHealthIssuesFilter parses filter parameters from request
func (s *Server) parseHealthIssuesFilter(c *gin.Context) models.HealthIssueFilterV2 {
	filter := models.HealthIssueFilterV2{}

	s.parseHealthIssueTypes(c, &filter)
	s.parseHealthIssueSeverities(c, &filter)
	s.parseHealthIssueSources(c, &filter)
	s.parseHealthIssueBooleanFilters(c, &filter)
	s.parseHealthIssueTimeFilters(c, &filter)

	return filter
}

// parseHealthIssueTypes parses type filters
func (s *Server) parseHealthIssueTypes(c *gin.Context, filter *models.HealthIssueFilterV2) {
	if types := c.QueryArray("types"); len(types) > 0 {
		filter.Types = types
	}
}

// parseHealthIssueSeverities parses severity filters
func (s *Server) parseHealthIssueSeverities(c *gin.Context, filter *models.HealthIssueFilterV2) {
	if severities := c.QueryArray("severities"); len(severities) > 0 {
		filter.Severities = severities
	}
}

// parseHealthIssueSources parses source filters
func (s *Server) parseHealthIssueSources(c *gin.Context, filter *models.HealthIssueFilterV2) {
	if sources := c.QueryArray("sources"); len(sources) > 0 {
		filter.Sources = sources
	}
}

// parseHealthIssueBooleanFilters parses resolved and dismissed boolean filters
func (s *Server) parseHealthIssueBooleanFilters(c *gin.Context, filter *models.HealthIssueFilterV2) {
	if resolved := c.Query("resolved"); resolved != "" {
		if r := resolved == trueBoolString; resolved == trueBoolString || resolved == falseBoolString {
			filter.Resolved = &r
		}
	}

	if dismissed := c.Query("dismissed"); dismissed != "" {
		if d := dismissed == trueBoolString; dismissed == trueBoolString || dismissed == falseBoolString {
			filter.Dismissed = &d
		}
	}
}

// parseHealthIssueTimeFilters parses time-based filters
func (s *Server) parseHealthIssueTimeFilters(c *gin.Context, filter *models.HealthIssueFilterV2) {
	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			filter.Since = &t
		}
	}

	if until := c.Query("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			filter.Until = &t
		}
	}
}

// buildHealthIssuesPaginatedResponse builds paginated response for health issues
func (s *Server) buildHealthIssuesPaginatedResponse(
	issues []models.HealthIssue, total int64, page, pageSize int,
) gin.H {
	totalPages := (int(total) + pageSize - 1) / pageSize

	return gin.H{
		"records":     issues,
		"total":       total,
		"page":        page,
		"pageSize":    pageSize,
		"totalPages":  totalPages,
		"hasNextPage": page < totalPages,
		"hasPrevPage": page > 1,
	}
}

// handleGetHealthIssue returns a specific health issue
func (s *Server) handleGetHealthIssue(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	issue, err := s.services.HealthService.GetHealthIssueByID(id)
	if err != nil {
		s.logger.Errorw("Failed to get health issue", "id", id, "error", err)
		if err.Error() == healthIssueNotFoundError {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health issue not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve health issue"})
		}
		return
	}

	c.JSON(http.StatusOK, issue)
}

// handleDismissHealthIssue dismisses a health issue
func (s *Server) handleDismissHealthIssue(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.HealthService.DismissHealthIssue(id); err != nil {
		s.logger.Errorw("Failed to dismiss health issue", "id", id, "error", err)
		if err.Error() == healthIssueNotFoundError {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health issue not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dismiss health issue"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health issue dismissed successfully"})
}

// handleResolveHealthIssue resolves a health issue
func (s *Server) handleResolveHealthIssue(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.HealthService.ResolveHealthIssue(id); err != nil {
		s.logger.Errorw("Failed to resolve health issue", "id", id, "error", err)
		if err.Error() == healthIssueNotFoundError {
			c.JSON(http.StatusNotFound, gin.H{"error": "Health issue not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve health issue"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health issue resolved successfully"})
}

// handleGetSystemResources returns current system resource usage
func (s *Server) handleGetSystemResources(c *gin.Context) {
	ctx := c.Request.Context()

	resources, err := s.services.HealthService.GetSystemResources(ctx)
	if err != nil {
		s.logger.Errorw("Failed to get system resources", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system resources"})
		return
	}

	c.JSON(http.StatusOK, resources)
}

// handleGetDiskSpace returns disk space information
func (s *Server) handleGetDiskSpace(c *gin.Context) {
	ctx := c.Request.Context()

	diskSpace, err := s.services.HealthService.CheckDiskSpace(ctx)
	if err != nil {
		s.logger.Errorw("Failed to check disk space", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check disk space"})
		return
	}

	c.JSON(http.StatusOK, diskSpace)
}

// handleGetPerformanceMetrics returns performance metrics with optional time range
func (s *Server) handleGetPerformanceMetrics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	var since, until *time.Time
	if sinceParam := c.Query("since"); sinceParam != "" {
		if t, err := time.Parse(time.RFC3339, sinceParam); err == nil {
			since = &t
		}
	}

	if untilParam := c.Query("until"); untilParam != "" {
		if t, err := time.Parse(time.RFC3339, untilParam); err == nil {
			until = &t
		}
	}

	metrics, err := s.services.HealthService.GetPerformanceMetrics(since, until, limit)
	if err != nil {
		s.logger.Errorw("Failed to get performance metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve performance metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
		"count":   len(metrics),
		"since":   since,
		"until":   until,
		"limit":   limit,
	})
}

// handleRecordPerformanceMetrics manually triggers recording of performance metrics
func (s *Server) handleRecordPerformanceMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.services.HealthService.RecordPerformanceMetrics(ctx); err != nil {
		s.logger.Errorw("Failed to record performance metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record performance metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Performance metrics recorded successfully"})
}

// handleStartHealthMonitoring starts the health monitoring background process
func (s *Server) handleStartHealthMonitoring(c *gin.Context) {
	ctx := c.Request.Context()

	if err := s.services.HealthService.StartMonitoring(ctx); err != nil {
		s.logger.Errorw("Failed to start health monitoring", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health monitoring started successfully"})
}

// handleStopHealthMonitoring stops the health monitoring background process
func (s *Server) handleStopHealthMonitoring(c *gin.Context) {
	if err := s.services.HealthService.StopMonitoring(); err != nil {
		s.logger.Errorw("Failed to stop health monitoring", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health monitoring stopped successfully"})
}

// handleCleanupHealthData cleans up old health data (metrics, resolved issues)
func (s *Server) handleCleanupHealthData(c *gin.Context) {
	if err := s.services.HealthService.CleanupOldMetrics(); err != nil {
		s.logger.Errorw("Failed to cleanup old metrics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup old metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Health data cleanup completed successfully"})
}
