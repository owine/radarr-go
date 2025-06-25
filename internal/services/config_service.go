package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// ConfigService provides operations for managing system configuration
type ConfigService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewConfigService creates a new instance of ConfigService
func NewConfigService(db *database.Database, logger *logger.Logger) *ConfigService {
	return &ConfigService{
		db:     db,
		logger: logger,
	}
}

// Host Configuration Management

// GetHostConfig retrieves the host configuration
func (s *ConfigService) GetHostConfig() (*models.HostConfig, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var config models.HostConfig
	if err := s.db.GORM.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default configuration
			return models.GetDefaultHostConfig(), nil
		}
		s.logger.Error("Failed to fetch host config", "error", err)
		return nil, fmt.Errorf("failed to fetch host config: %w", err)
	}

	return &config, nil
}

// UpdateHostConfig updates the host configuration
func (s *ConfigService) UpdateHostConfig(config *models.HostConfig) error {
	return s.updateConfig(config, "host config", func() []string {
		return config.ValidateConfiguration()
	})
}

// Naming Configuration Management

// GetNamingConfig retrieves the naming configuration
func (s *ConfigService) GetNamingConfig() (*models.NamingConfig, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var config models.NamingConfig
	if err := s.db.GORM.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default configuration
			return models.GetDefaultNamingConfig(), nil
		}
		s.logger.Error("Failed to fetch naming config", "error", err)
		return nil, fmt.Errorf("failed to fetch naming config: %w", err)
	}

	return &config, nil
}

// UpdateNamingConfig updates the naming configuration
func (s *ConfigService) UpdateNamingConfig(config *models.NamingConfig) error {
	return s.updateConfig(config, "naming config", func() []string {
		return config.ValidateConfiguration()
	})
}

// GetNamingTokens returns available naming tokens
func (s *ConfigService) GetNamingTokens() []models.NamingToken {
	return models.GetAvailableTokens()
}

// Media Management Configuration

// GetMediaManagementConfig retrieves the media management configuration
func (s *ConfigService) GetMediaManagementConfig() (*models.MediaManagementConfig, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var config models.MediaManagementConfig
	if err := s.db.GORM.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default configuration
			return models.GetDefaultMediaManagementConfig(), nil
		}
		s.logger.Error("Failed to fetch media management config", "error", err)
		return nil, fmt.Errorf("failed to fetch media management config: %w", err)
	}

	return &config, nil
}

// UpdateMediaManagementConfig updates the media management configuration
func (s *ConfigService) UpdateMediaManagementConfig(config *models.MediaManagementConfig) error {
	return s.updateConfig(config, "media management config", func() []string {
		return config.ValidateConfiguration()
	})
}

// Root Folder Management

// GetRootFolders retrieves all root folders
func (s *ConfigService) GetRootFolders() ([]models.RootFolder, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var rootFolders []models.RootFolder
	if err := s.db.GORM.Find(&rootFolders).Error; err != nil {
		s.logger.Error("Failed to fetch root folders", "error", err)
		return nil, fmt.Errorf("failed to fetch root folders: %w", err)
	}

	// Update accessibility and space information
	for i := range rootFolders {
		s.updateRootFolderStats(&rootFolders[i])
	}

	s.logger.Debug("Retrieved root folders", "count", len(rootFolders))
	return rootFolders, nil
}

// GetRootFolderByID retrieves a specific root folder by ID
func (s *ConfigService) GetRootFolderByID(id int) (*models.RootFolder, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var rootFolder models.RootFolder
	if err := s.db.GORM.First(&rootFolder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("root folder not found")
		}
		s.logger.Error("Failed to fetch root folder", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch root folder: %w", err)
	}

	// Update accessibility and space information
	s.updateRootFolderStats(&rootFolder)

	return &rootFolder, nil
}

// CreateRootFolder creates a new root folder
func (s *ConfigService) CreateRootFolder(rootFolder *models.RootFolder) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate root folder
	if errors := models.ValidateRootFolder(rootFolder); len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	// Check if path already exists
	var existingFolder models.RootFolder
	err := s.db.GORM.Where("path = ?", rootFolder.Path).First(&existingFolder).Error
	if err == nil {
		return fmt.Errorf("root folder with path '%s' already exists", rootFolder.Path)
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing root folder: %w", err)
	}

	// Update stats before creating
	s.updateRootFolderStats(rootFolder)

	if err := s.db.GORM.Create(rootFolder).Error; err != nil {
		s.logger.Error("Failed to create root folder", "path", rootFolder.Path, "error", err)
		return fmt.Errorf("failed to create root folder: %w", err)
	}

	s.logger.Info("Created root folder", "id", rootFolder.ID, "path", rootFolder.Path)
	return nil
}

// UpdateRootFolder updates an existing root folder
func (s *ConfigService) UpdateRootFolder(rootFolder *models.RootFolder) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate root folder
	if errors := models.ValidateRootFolder(rootFolder); len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	// Update stats before saving
	s.updateRootFolderStats(rootFolder)

	if err := s.db.GORM.Save(rootFolder).Error; err != nil {
		s.logger.Error("Failed to update root folder", "id", rootFolder.ID, "error", err)
		return fmt.Errorf("failed to update root folder: %w", err)
	}

	s.logger.Info("Updated root folder", "id", rootFolder.ID, "path", rootFolder.Path)
	return nil
}

// DeleteRootFolder removes a root folder
func (s *ConfigService) DeleteRootFolder(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Check if root folder exists
	_, err := s.GetRootFolderByID(id)
	if err != nil {
		return fmt.Errorf("root folder not found: %w", err)
	}

	result := s.db.GORM.Delete(&models.RootFolder{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete root folder", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete root folder: %w", result.Error)
	}

	s.logger.Info("Deleted root folder", "id", id)
	return nil
}

// Helper Methods

// updateRootFolderStats updates the accessibility and space information for a root folder
func (s *ConfigService) updateRootFolderStats(rootFolder *models.RootFolder) {
	// Check accessibility
	if _, err := os.Stat(rootFolder.Path); err != nil {
		rootFolder.Accessible = false
		s.logger.Debug("Root folder not accessible", "path", rootFolder.Path, "error", err)
		return
	}

	rootFolder.Accessible = true

	// Get disk space information
	if usage, err := s.getDiskUsage(rootFolder.Path); err == nil {
		rootFolder.FreeSpace = usage.Free / (1024 * 1024)   // Convert to MB
		rootFolder.TotalSpace = usage.Total / (1024 * 1024) // Convert to MB
	} else {
		s.logger.Debug("Failed to get disk usage", "path", rootFolder.Path, "error", err)
	}
}

// DiskUsage represents disk usage information
type DiskUsage struct {
	Free  int64
	Total int64
}

// getDiskUsage returns disk usage information for the given path
func (s *ConfigService) getDiskUsage(path string) (*DiskUsage, error) {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Find the mount point by walking up the directory tree
	for {
		if _, err := os.Stat(absPath); err == nil {
			break
		}
		parent := filepath.Dir(absPath)
		if parent == absPath {
			return nil, fmt.Errorf("could not find accessible parent directory")
		}
		absPath = parent
	}

	// Get disk usage for the path
	usage, err := getDiskUsageForPath(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	return usage, nil
}

// GetConfigurationStats returns statistics about the configuration
func (s *ConfigService) GetConfigurationStats() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]interface{})

	// Count root folders
	var rootFolderCount int64
	if err := s.db.GORM.Model(&models.RootFolder{}).Count(&rootFolderCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count root folders: %w", err)
	}
	stats["rootFolders"] = rootFolderCount

	// Count accessible root folders
	var accessibleCount int64
	err := s.db.GORM.Model(&models.RootFolder{}).Where("accessible = ?", true).Count(&accessibleCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count accessible root folders: %w", err)
	}
	stats["accessibleRootFolders"] = accessibleCount

	// Get host config status
	hostConfig, err := s.GetHostConfig()
	if err == nil {
		stats["authenticationEnabled"] = hostConfig.IsAuthenticationEnabled()
		stats["sslEnabled"] = hostConfig.IsSSLEnabled()
		stats["analyticsEnabled"] = hostConfig.AnalyticsEnabled
	}

	// Get naming config status
	namingConfig, err := s.GetNamingConfig()
	if err == nil {
		stats["renamingEnabled"] = namingConfig.RenameMovies
		stats["mediaInfoEnabled"] = namingConfig.EnableMediaInfo
	}

	// Get media management config status
	mediaConfig, err := s.GetMediaManagementConfig()
	if err == nil {
		stats["recycleBinEnabled"] = mediaConfig.IsRecycleBinEnabled()
		stats["permissionsEnabled"] = mediaConfig.IsPermissionsEnabled()
		stats["libraryWatchingEnabled"] = mediaConfig.WatchLibraryForChanges
	}

	return stats, nil
}

// updateConfig is a generic helper for updating configuration objects
func (s *ConfigService) updateConfig(config interface{}, configType string, validate func() []string) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate configuration
	if errors := validate(); len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	// Use reflection to check if config exists and handle create/update
	err := s.db.GORM.Where("id = ?", 1).First(config).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing config: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Create new configuration
		if err := s.db.GORM.Create(config).Error; err != nil {
			s.logger.Error("Failed to create "+configType, "error", err)
			return fmt.Errorf("failed to create %s: %w", configType, err)
		}
	} else {
		// Update existing configuration
		if err := s.db.GORM.Save(config).Error; err != nil {
			s.logger.Error("Failed to update "+configType, "error", err)
			return fmt.Errorf("failed to update %s: %w", configType, err)
		}
	}

	s.logger.Info("Updated " + configType + " configuration")
	return nil
}
