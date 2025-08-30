// Package models defines data structures and database models for Radarr.
// This file contains the refactored Health models with simplified validation.
package models

import (
	"encoding/json"
	"time"
)

// HealthIssueV2 represents a health issue with simplified structure and no problematic hooks
type HealthIssueV2 struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Type     string `json:"type" gorm:"not null;size:50"`
	Source   string `json:"source" gorm:"not null;size:100"`
	Severity string `json:"severity" gorm:"not null;size:20;check:severity IN ('info', 'warning', 'error', 'critical')"`
	Message  string `json:"message" gorm:"not null;type:text"`

	// Status flags
	IsResolved  bool `json:"isResolved" gorm:"default:false"`
	IsDismissed bool `json:"isDismissed" gorm:"default:false"`

	// Metadata as JSON
	Details JSONField `json:"details,omitempty" gorm:"type:json"`

	// Timestamps
	FirstSeen  time.Time  `json:"firstSeen" gorm:"autoCreateTime"`
	LastSeen   time.Time  `json:"lastSeen" gorm:"autoCreateTime"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (HealthIssueV2) TableName() string {
	return "health_issues"
}

// Validate performs basic validation
func (h *HealthIssueV2) Validate() error {
	if h.Type == "" {
		return ValidationError{Field: "type", Message: "Type is required"}
	}
	if h.Source == "" {
		return ValidationError{Field: "source", Message: "Source is required"}
	}
	if h.Severity == "" {
		return ValidationError{Field: "severity", Message: "Severity is required"}
	}
	if h.Message == "" {
		return ValidationError{Field: "message", Message: "Message is required"}
	}
	return nil
}

// Resolve marks the health issue as resolved
func (h *HealthIssueV2) Resolve() {
	now := time.Now()
	h.IsResolved = true
	h.ResolvedAt = &now
}

// Dismiss marks the health issue as dismissed
func (h *HealthIssueV2) Dismiss() {
	h.IsDismissed = true
}

// UpdateLastSeen updates the last seen timestamp
func (h *HealthIssueV2) UpdateLastSeen() {
	h.LastSeen = time.Now()
}

// IsCritical returns true if the issue is critical severity
func (h *HealthIssueV2) IsCritical() bool {
	return h.Severity == "critical"
}

// IsWarning returns true if the issue is warning severity
func (h *HealthIssueV2) IsWarning() bool {
	return h.Severity == "warning"
}

// IsError returns true if the issue is error severity
func (h *HealthIssueV2) IsError() bool {
	return h.Severity == "error"
}

// Health check types
const (
	HealthTypeDatabase        = "database"
	HealthTypeDiskSpace       = "diskSpace"
	HealthTypeDownloadClient  = "downloadClient"
	HealthTypeIndexer         = "indexer"
	HealthTypeImportList      = "importList"
	HealthTypeRootFolder      = "rootFolder"
	HealthTypeNetwork         = "network"
	HealthTypeSystem          = "system"
	HealthTypePerformance     = "performance"
	HealthTypeConfiguration   = "configuration"
	HealthTypeExternalService = "externalService"
)

// Health severities
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// HealthCheckResultV2 represents the result of running health checks (simplified)
type HealthCheckResultV2 struct {
	OverallStatus string               `json:"overallStatus"`
	Issues        []HealthIssueV2      `json:"issues"`
	Summary       HealthCheckSummaryV2 `json:"summary"`
	Timestamp     time.Time            `json:"timestamp"`
	Duration      time.Duration        `json:"duration"`
}

// HealthCheckSummaryV2 provides a summary of health check results (V2 simplified version)
type HealthCheckSummaryV2 struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Error    int `json:"error"`
	Critical int `json:"critical"`
}

// HealthIssueFilterV2 represents filters for health issues (simplified)
type HealthIssueFilterV2 struct {
	Types      []string   `json:"types,omitempty"`
	Severities []string   `json:"severities,omitempty"`
	Sources    []string   `json:"sources,omitempty"`
	Resolved   *bool      `json:"resolved,omitempty"`
	Dismissed  *bool      `json:"dismissed,omitempty"`
	Since      *time.Time `json:"since,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
}

// GetWorstSeverity returns the worst severity from a list of severities
func GetWorstSeverity(severities ...string) string {
	severityOrder := map[string]int{
		"info":     1,
		"warning":  2,
		"error":    3,
		"critical": 4,
	}

	worst := "info"
	worstScore := 0

	for _, severity := range severities {
		if score, exists := severityOrder[severity]; exists && score > worstScore {
			worst = severity
			worstScore = score
		}
	}

	return worst
}

// ToV1 converts HealthIssueV2 to HealthIssue (for API compatibility)
func (h *HealthIssueV2) ToV1() *HealthIssue {
	v1 := &HealthIssue{
		ID:          h.ID,
		Type:        HealthCheckType(h.Type),
		Source:      h.Source,
		Severity:    HealthSeverity(h.Severity),
		Message:     h.Message,
		FirstSeen:   h.FirstSeen,
		LastSeen:    h.LastSeen,
		ResolvedAt:  h.ResolvedAt,
		IsResolved:  h.IsResolved,
		IsDismissed: h.IsDismissed,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}

	// Convert details from JSON to json.RawMessage
	if h.Details != nil {
		if detailsBytes, err := json.Marshal(h.Details); err == nil {
			v1.Details = detailsBytes
		}
	}

	// Convert details to map[string]interface{} for Data field
	if h.Details != nil {
		if detailsMap, ok := h.Details["data"].(map[string]interface{}); ok {
			v1.Data = detailsMap
		}
	}

	return v1
}

// FromV1 converts HealthIssue to HealthIssueV2
func (h *HealthIssueV2) FromV1(v1 *HealthIssue) {
	h.ID = v1.ID
	h.Type = string(v1.Type)
	h.Source = v1.Source
	h.Severity = string(v1.Severity)
	h.Message = v1.Message
	h.FirstSeen = v1.FirstSeen
	h.LastSeen = v1.LastSeen
	h.ResolvedAt = v1.ResolvedAt
	h.IsResolved = v1.IsResolved
	h.IsDismissed = v1.IsDismissed
	h.CreatedAt = v1.CreatedAt
	h.UpdatedAt = v1.UpdatedAt

	// Convert details from json.RawMessage and Data to JSON
	if h.Details == nil {
		h.Details = make(JSONField)
	}

	// If we have Data, add it to Details
	if v1.Data != nil {
		h.Details["data"] = v1.Data
	}
}

// NewHealthIssueV2FromV1 creates a new HealthIssueV2 from HealthIssue
func NewHealthIssueV2FromV1(v1 *HealthIssue) *HealthIssueV2 {
	v2 := &HealthIssueV2{}
	v2.FromV1(v1)
	return v2
}

// ConvertHealthIssueV2SliceToV1 converts a slice of HealthIssueV2 to HealthIssue
func ConvertHealthIssueV2SliceToV1(v2Issues []HealthIssueV2) []HealthIssue {
	v1Issues := make([]HealthIssue, 0, len(v2Issues))
	for _, v2Issue := range v2Issues {
		v1Issues = append(v1Issues, *v2Issue.ToV1())
	}
	return v1Issues
}

// ConvertHealthIssueV1SliceToV2 converts a slice of HealthIssue to HealthIssueV2
func ConvertHealthIssueV1SliceToV2(v1Issues []HealthIssue) []HealthIssueV2 {
	v2Issues := make([]HealthIssueV2, 0, len(v1Issues))
	for _, v1Issue := range v1Issues {
		v2Issues = append(v2Issues, *NewHealthIssueV2FromV1(&v1Issue))
	}
	return v2Issues
}

// ConvertHealthCheckTypesToStrings converts HealthCheckType slice to string slice
func ConvertHealthCheckTypesToStrings(types []HealthCheckType) []string {
	strings := make([]string, 0, len(types))
	for _, t := range types {
		strings = append(strings, string(t))
	}
	return strings
}

// ConvertStringsToHealthCheckTypes converts string slice to HealthCheckType slice
func ConvertStringsToHealthCheckTypes(strings []string) []HealthCheckType {
	types := make([]HealthCheckType, 0, len(strings))
	for _, s := range strings {
		types = append(types, HealthCheckType(s))
	}
	return types
}

// ToV1Filter converts HealthIssueFilterV2 to HealthIssueFilter
func (f *HealthIssueFilterV2) ToV1Filter() *HealthIssueFilter {
	v1 := &HealthIssueFilter{
		Sources:   f.Sources,
		Resolved:  f.Resolved,
		Dismissed: f.Dismissed,
		Since:     f.Since,
		Until:     f.Until,
	}

	// Convert string slices to typed slices
	if len(f.Types) > 0 {
		v1.Types = ConvertStringsToHealthCheckTypes(f.Types)
	}

	if len(f.Severities) > 0 {
		v1.Severities = make([]HealthSeverity, 0, len(f.Severities))
		for _, s := range f.Severities {
			v1.Severities = append(v1.Severities, HealthSeverity(s))
		}
	}

	return v1
}

// FromV1Filter converts HealthIssueFilter to HealthIssueFilterV2
func (f *HealthIssueFilterV2) FromV1Filter(v1 *HealthIssueFilter) {
	f.Sources = v1.Sources
	f.Resolved = v1.Resolved
	f.Dismissed = v1.Dismissed
	f.Since = v1.Since
	f.Until = v1.Until

	// Convert typed slices to string slices
	if len(v1.Types) > 0 {
		f.Types = ConvertHealthCheckTypesToStrings(v1.Types)
	}

	if len(v1.Severities) > 0 {
		f.Severities = make([]string, 0, len(v1.Severities))
		for _, s := range v1.Severities {
			f.Severities = append(f.Severities, string(s))
		}
	}
}
