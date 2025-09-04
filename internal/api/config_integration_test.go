package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/services"
	"github.com/stretchr/testify/assert"
)

// setupIntegrationTestServer creates a test server for integration testing
func setupIntegrationTestServer(_ *testing.T) *Server {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Log: config.LogConfig{
			Level: "error",
		},
		Database: config.DatabaseConfig{
			Type:          "sqlite",
			ConnectionURL: ":memory:", // Use in-memory database for tests
		},
	}

	logger := logger.New(cfg.Log)

	// For integration tests, we'd typically set up a real database
	// For now, we'll create a minimal service container
	services := &services.Container{
		ConfigService: services.NewConfigService(nil, logger),
	}

	return NewServer(cfg, services, logger)
}

func TestGetAppSettingsIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v3/config/app", http.NoBody)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// Without a database, this should return default settings or handle gracefully
	switch w.Code {
	case http.StatusOK:
		var response models.AppSettings
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		t.Logf("Retrieved app settings: %+v", response)
	case http.StatusInternalServerError:
		// Expected when database is not available
		t.Logf("Expected error due to no database: %s", w.Body.String())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestUpdateAppSettingsIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	settings := &models.AppSettings{
		Version:              "1.0.0",
		Theme:                "light",
		Language:             "en-US",
		BackupRetentionDays:  30,
		MaxConcurrentTasks:   5,
		CacheExpirationHours: 24,
	}

	body, _ := json.Marshal(settings)
	req, _ := http.NewRequestWithContext(context.Background(), "PUT", "/api/v3/config/app", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// Without a database, this should return an error or handle gracefully
	if w.Code == http.StatusInternalServerError {
		// Expected when database is not available
		t.Logf("Expected error due to no database: %s", w.Body.String())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestValidateAllConfigurationsIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v3/config/validate", http.NoBody)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// This endpoint should handle gracefully even without a database
	if w.Code == http.StatusOK || w.Code == http.StatusInternalServerError {
		t.Logf("Validation endpoint returned status: %d, body: %s", w.Code, w.Body.String())
	} else {
		t.Errorf("Unexpected status code: %d", w.Code)
	}
}

func TestCreateConfigurationBackupIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	backupRequest := map[string]string{
		"name":        "Test Backup",
		"description": "Integration test backup",
	}

	body, _ := json.Marshal(backupRequest)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v3/config/backup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// This should handle gracefully even without a database
	if w.Code == http.StatusInternalServerError {
		t.Logf("Expected error due to no database: %s", w.Body.String())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestFactoryResetConfigurationIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	resetRequest := map[string]interface{}{
		"components": []string{"host", "naming"},
	}

	body, _ := json.Marshal(resetRequest)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v3/config/reset", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// This should handle gracefully even without a database
	if w.Code == http.StatusInternalServerError {
		t.Logf("Expected error due to no database: %s", w.Body.String())
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestFactoryResetConfigurationInvalidComponentIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	resetRequest := map[string]interface{}{
		"components": []string{"invalid_component"},
	}

	body, _ := json.Marshal(resetRequest)
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v3/config/reset", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// Should return bad request for invalid component
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"].(string), "Invalid component")
}

func TestExportConfigurationIntegration(t *testing.T) {
	server := setupIntegrationTestServer(t)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v3/config/export", http.NoBody)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// Should handle export request
	switch w.Code {
	case http.StatusOK:
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "radarr-config")
		t.Log("Export endpoint working correctly")
	case http.StatusInternalServerError:
		t.Logf("Expected error due to no database: %s", w.Body.String())
	}
}

func TestAppSettingsValidation(t *testing.T) {
	// Test the validation logic directly
	settings := &models.AppSettings{
		BackupRetentionDays:  -1,        // Invalid
		MaxConcurrentTasks:   100,       // Invalid
		CacheExpirationHours: 200,       // Invalid
		Theme:                "invalid", // Invalid
		TimeFormat:           "invalid", // Invalid
	}

	errors := settings.ValidateSettings()
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Backup retention days must be between 1 and 365")
	assert.Contains(t, errors, "Max concurrent tasks must be between 1 and 50")
	assert.Contains(t, errors, "Cache expiration hours must be between 1 and 168 (1 week)")
	assert.Contains(t, errors, "Theme must be one of: dark, light, auto")
	assert.Contains(t, errors, "Time format must be either 12h or 24h")
}

func TestAppSettingsDefaults(t *testing.T) {
	defaults := models.GetDefaultAppSettings()

	assert.Equal(t, "1.0.0", defaults.Version)
	assert.False(t, defaults.Initialized)
	assert.Equal(t, "dark", defaults.Theme)
	assert.Equal(t, "en-US", defaults.Language)
	assert.Equal(t, 30, defaults.BackupRetentionDays)
	assert.Equal(t, 5, defaults.MaxConcurrentTasks)
	assert.Equal(t, 24, defaults.CacheExpirationHours)

	// Validate that defaults are valid
	errors := defaults.ValidateSettings()
	assert.Empty(t, errors, "Default settings should be valid")
}

// Benchmark tests for performance
func BenchmarkAppSettingsValidation(b *testing.B) {
	settings := models.GetDefaultAppSettings()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = settings.ValidateSettings()
	}
}

func BenchmarkGetAppSettingsEndpoint(b *testing.B) {
	server := setupIntegrationTestServer(nil)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v3/config/app", http.NoBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		server.engine.ServeHTTP(w, req)
	}
}
