package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// DownloadService provides operations for managing download clients and download queue.
type DownloadService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewDownloadService creates a new instance of DownloadService with the provided database and logger.
func NewDownloadService(db *database.Database, logger *logger.Logger) *DownloadService {
	return &DownloadService{
		db:     db,
		logger: logger,
	}
}

// GetDownloadClients retrieves all configured download clients.
func (s *DownloadService) GetDownloadClients() ([]models.DownloadClient, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var clients []models.DownloadClient

	if err := s.db.GORM.Find(&clients).Error; err != nil {
		s.logger.Error("Failed to fetch download clients", "error", err)
		return nil, fmt.Errorf("failed to fetch download clients: %w", err)
	}

	s.logger.Debug("Retrieved download clients", "count", len(clients))
	return clients, nil
}

// GetDownloadClientByID retrieves a specific download client by its ID.
func (s *DownloadService) GetDownloadClientByID(id int) (*models.DownloadClient, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var client models.DownloadClient

	if err := s.db.GORM.First(&client, id).Error; err != nil {
		s.logger.Error("Failed to fetch download client", "id", id, "error", err)
		return nil, fmt.Errorf("download client not found: %w", err)
	}

	return &client, nil
}

// CreateDownloadClient creates a new download client configuration.
func (s *DownloadService) CreateDownloadClient(client *models.DownloadClient) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateDownloadClient(client); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Create(client).Error; err != nil {
		s.logger.Error("Failed to create download client", "name", client.Name, "error", err)
		return fmt.Errorf("failed to create download client: %w", err)
	}

	s.logger.Info("Created download client", "id", client.ID, "name", client.Name, "type", client.Type)
	return nil
}

// UpdateDownloadClient updates an existing download client configuration.
func (s *DownloadService) UpdateDownloadClient(client *models.DownloadClient) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateDownloadClient(client); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Save(client).Error; err != nil {
		s.logger.Error("Failed to update download client", "id", client.ID, "error", err)
		return fmt.Errorf("failed to update download client: %w", err)
	}

	s.logger.Info("Updated download client", "id", client.ID, "name", client.Name)
	return nil
}

// DeleteDownloadClient removes a download client configuration.
func (s *DownloadService) DeleteDownloadClient(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Check if client exists first
	_, err := s.GetDownloadClientByID(id)
	if err != nil {
		return fmt.Errorf("download client not found: %w", err)
	}

	result := s.db.GORM.Delete(&models.DownloadClient{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete download client", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete download client: %w", result.Error)
	}

	s.logger.Info("Deleted download client", "id", id)
	return nil
}

// GetEnabledDownloadClients retrieves all enabled download clients.
func (s *DownloadService) GetEnabledDownloadClients() ([]models.DownloadClient, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var clients []models.DownloadClient

	if err := s.db.GORM.Where("enable = ?", true).Find(&clients).Error; err != nil {
		s.logger.Error("Failed to fetch enabled download clients", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled download clients: %w", err)
	}

	s.logger.Debug("Retrieved enabled download clients", "count", len(clients))
	return clients, nil
}

// TestDownloadClient tests the connection to a download client.
func (s *DownloadService) TestDownloadClient(client *models.DownloadClient) (*models.DownloadClientTestResult, error) {
	// Basic validation
	errors := []string{}

	if client.Name == "" {
		errors = append(errors, "Name is required")
	}

	if client.Host == "" {
		errors = append(errors, "Host is required")
	}

	if client.Port <= 0 || client.Port > 65535 {
		errors = append(errors, "Port must be between 1 and 65535")
	}

	result := &models.DownloadClientTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// Test actual connection
	if err := s.testClientConnection(client); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	}

	s.logger.Info("Tested download client connection", "name", client.Name, "valid", result.IsValid)
	return result, nil
}

// GetDownloadHistory retrieves download history.
func (s *DownloadService) GetDownloadHistory(limit int) ([]models.DownloadHistory, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var history []models.DownloadHistory

	query := s.db.GORM.Preload("Movie").Preload("DownloadClient").Order("date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&history).Error; err != nil {
		s.logger.Error("Failed to fetch download history", "error", err)
		return nil, fmt.Errorf("failed to fetch download history: %w", err)
	}

	s.logger.Debug("Retrieved download history", "count", len(history))
	return history, nil
}

// AddToHistory adds a download to the history.
func (s *DownloadService) AddToHistory(history *models.DownloadHistory) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	if err := s.db.GORM.Create(history).Error; err != nil {
		s.logger.Error("Failed to add to download history", "title", history.SourceTitle, "error", err)
		return fmt.Errorf("failed to add to download history: %w", err)
	}

	s.logger.Info("Added to download history", "id", history.ID, "title", history.SourceTitle,
		"successful", history.Successful)
	return nil
}

// GetDownloads retrieves all active downloads from the download queue (legacy method for compatibility).
func (s *DownloadService) GetDownloads() ([]models.QueueItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var items []models.QueueItem

	if err := s.db.GORM.Preload("Movie").Preload("DownloadClient").Find(&items).Error; err != nil {
		s.logger.Error("Failed to fetch queue items", "error", err)
		return nil, fmt.Errorf("failed to fetch queue items: %w", err)
	}

	s.logger.Debug("Retrieved queue items", "count", len(items))
	return items, nil
}

// GetDownloadClientsByProtocol retrieves download clients that support a specific protocol
func (s *DownloadService) GetDownloadClientsByProtocol(
	protocol models.DownloadProtocol) ([]models.DownloadClient, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var clients []models.DownloadClient

	if err := s.db.GORM.Where("protocol = ? AND enable = ?", protocol, true).Find(&clients).Error; err != nil {
		s.logger.Error("Failed to fetch download clients by protocol", "protocol", protocol, "error", err)
		return nil, fmt.Errorf("failed to fetch download clients by protocol: %w", err)
	}

	s.logger.Debug("Retrieved download clients by protocol", "protocol", protocol, "count", len(clients))
	return clients, nil
}

// validateDownloadClient performs validation on download client configuration
func (s *DownloadService) validateDownloadClient(client *models.DownloadClient) error {
	if client.Name == "" {
		return fmt.Errorf("name is required")
	}

	if client.Host == "" {
		return fmt.Errorf("host is required")
	}

	if client.Port <= 0 || client.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if client.Type == "" {
		return fmt.Errorf("client type is required")
	}

	if client.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}

	// Validate protocol matches client type
	if !s.isValidProtocolForClientType(client.Type, client.Protocol) {
		return fmt.Errorf("protocol %s is not supported by client type %s", client.Protocol, client.Type)
	}

	return nil
}

// isValidProtocolForClientType checks if a protocol is valid for a given client type
func (s *DownloadService) isValidProtocolForClientType(
	clientType models.DownloadClientType, protocol models.DownloadProtocol) bool {
	switch clientType {
	case models.DownloadClientTypeQBittorrent, models.DownloadClientTypeTransmission,
		models.DownloadClientTypeDeluge, models.DownloadClientTypeRTorrent,
		models.DownloadClientTypeUtorrent:
		return protocol == models.DownloadProtocolTorrent
	case models.DownloadClientTypeSABnzbd, models.DownloadClientTypeNZBGet:
		return protocol == models.DownloadProtocolUsenet
	default:
		return false
	}
}

// testClientConnection tests the actual connection to a download client
func (s *DownloadService) testClientConnection(client *models.DownloadClient) error {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build test URL based on client type
	testURL := client.GetBaseURL()
	switch client.Type {
	case models.DownloadClientTypeQBittorrent:
		testURL += "/api/v2/app/version"
	case models.DownloadClientTypeTransmission:
		testURL += "/transmission/rpc"
	case models.DownloadClientTypeSABnzbd:
		testURL += "/api?mode=version"
	case models.DownloadClientTypeNZBGet:
		testURL += "/jsonrpc"
	case models.DownloadClientTypeDeluge, models.DownloadClientTypeRTorrent,
		models.DownloadClientTypeUtorrent:
		// Generic test for other torrent clients
		testURL += "/"
	default:
		// Generic test - just try to connect to the host
		testURL += "/"
	}

	// Make test request with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	// Check if we get a reasonable response
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return nil
}

// GetDownloadClientStats returns statistics about download clients
func (s *DownloadService) GetDownloadClientStats() (map[string]any, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]any)

	// Count total clients
	var totalClients int64
	if err := s.db.GORM.Model(&models.DownloadClient{}).Count(&totalClients).Error; err != nil {
		return nil, fmt.Errorf("failed to count total clients: %w", err)
	}
	stats["total"] = totalClients

	// Count enabled clients
	var enabledClients int64
	err := s.db.GORM.Model(&models.DownloadClient{}).Where("enable = ?", true).Count(&enabledClients).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count enabled clients: %w", err)
	}
	stats["enabled"] = enabledClients

	// Count by protocol
	var torrentClients int64
	if err := s.db.GORM.Model(&models.DownloadClient{}).
		Where("protocol = ?", models.DownloadProtocolTorrent).Count(&torrentClients).Error; err != nil {
		return nil, fmt.Errorf("failed to count torrent clients: %w", err)
	}
	stats["torrent"] = torrentClients

	var usenetClients int64
	if err := s.db.GORM.Model(&models.DownloadClient{}).
		Where("protocol = ?", models.DownloadProtocolUsenet).Count(&usenetClients).Error; err != nil {
		return nil, fmt.Errorf("failed to count usenet clients: %w", err)
	}
	stats["usenet"] = usenetClients

	return stats, nil
}
