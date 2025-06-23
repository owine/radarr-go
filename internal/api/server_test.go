package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create test configuration
	cfg := &config.Config{
		Log: config.LogConfig{
			Level: "error", // Reduce noise in tests
		},
	}
	
	// Create test logger
	logger := logger.New(cfg.Log)
	
	// Create test services (can be nil for this test)
	services := &services.Container{}
	
	// Create test server
	server := NewServer(cfg, services, logger)
	
	// Create test request
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.engine.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

func TestSystemStatusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create test configuration
	cfg := &config.Config{
		Log: config.LogConfig{
			Level: "error",
		},
		Server: config.ServerConfig{
			URLBase: "/radarr",
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
		},
		Auth: config.AuthConfig{
			Method: "none",
		},
		Storage: config.StorageConfig{
			DataDirectory: "./test-data",
		},
	}
	
	// Create test logger
	logger := logger.New(cfg.Log)
	
	// Create test services
	services := &services.Container{}
	
	// Create test server
	server := NewServer(cfg, services, logger)
	
	// Create test request
	req, _ := http.NewRequest("GET", "/api/v3/system/status", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.engine.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "1.0.0-go")
	assert.Contains(t, w.Body.String(), "sqlite")
}