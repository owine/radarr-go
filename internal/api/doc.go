// Package api provides the HTTP API layer for the Radarr application.
//
// This package implements a RESTful API server that maintains 100% compatibility
// with Radarr's v3 API while providing significant performance improvements through
// Go's efficient HTTP handling and the Gin web framework.
//
// API Features:
//
//   - RESTful endpoints matching Radarr v3 API specification
//   - JSON request/response handling with proper error responses
//   - API key authentication (header and query parameter support)
//   - CORS support for web client integration
//   - Request logging and monitoring
//   - Health check endpoints for monitoring
//   - Graceful shutdown with proper resource cleanup
//
// Server Architecture:
//
// The API server is built on Gin framework with middleware for:
//   - Request logging
//   - CORS handling
//   - API key authentication
//   - Error recovery and handling
//   - Response formatting
//
// Endpoint Categories:
//
//   - Movies: CRUD operations, search, and metadata
//   - Quality: Quality profiles and definitions
//   - Indexers: Search provider management
//   - Download Clients: Download automation
//   - Import Lists: Automatic movie discovery
//   - History: Event tracking and audit logs
//   - Queue: Background task management
//   - System: Health checks, status, and configuration
//
// Authentication:
//
// The API supports optional API key authentication:
//   - X-API-Key header
//   - apikey query parameter
//   - Configurable per-endpoint requirements
//
// Example usage:
//
//	server := api.NewServer(config, services, logger)
//	if err := server.Start(); err != nil {
//		log.Fatal("Failed to start server:", err)
//	}
//
// Health Monitoring:
//
// The server provides health check endpoints for monitoring:
//   - GET /ping - Basic connectivity test
//   - GET /api/v3/system/status - Detailed system information
//
// Error Handling:
//
// All endpoints return consistent JSON error responses with appropriate
// HTTP status codes and detailed error messages for debugging.
//
//nolint:revive // "api" is a standard package name for API layers
package api
