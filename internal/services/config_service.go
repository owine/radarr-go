package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// ConfigService provides operations for managing system configuration
type ConfigService struct {
	db       *database.Database
	logger   *logger.Logger
	services *Container
}

// NewConfigService creates a new instance of ConfigService
func NewConfigService(db *database.Database, logger *logger.Logger) *ConfigService {
	return &ConfigService{
		db:       db,
		logger:   logger,
		services: nil, // Will be set later via SetServiceContainer
	}
}

// SetServiceContainer sets the service container reference for accessing other services
func (s *ConfigService) SetServiceContainer(services *Container) {
	s.services = services
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

// Application Settings Management

// GetAppSettings retrieves consolidated application settings
func (s *ConfigService) GetAppSettings() (*models.AppSettings, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	settings := models.GetDefaultAppSettings()

	// Retrieve all app config entries
	var configs []models.AppConfig
	if err := s.db.GORM.Find(&configs).Error; err != nil {
		s.logger.Error("Failed to fetch app configs", "error", err)
		return settings, nil // Return defaults if no configs exist
	}

	// Map config values to settings struct
	configMap := make(map[string]interface{})
	for _, config := range configs {
		var value interface{}
		if err := config.Value.Scan(config.Value); err == nil {
			configMap[config.Key] = value
		}
	}

	// Apply config values to settings
	s.applyConfigToSettings(configMap, settings)

	return settings, nil
}

// UpdateAppSettings updates application settings
func (s *ConfigService) UpdateAppSettings(settings *models.AppSettings) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate settings
	if errors := settings.ValidateSettings(); len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	// Convert settings to config entries
	configs := s.convertSettingsToConfigs(settings)

	// Update each configuration entry
	for _, config := range configs {
		if err := s.db.GORM.Save(&config).Error; err != nil {
			s.logger.Error("Failed to update app config", "key", config.Key, "error", err)
			return fmt.Errorf("failed to update app config %s: %w", config.Key, err)
		}
	}

	s.logger.Info("Updated application settings")
	return nil
}

// Configuration Backup and Restore

// CreateConfigurationBackup creates a complete backup of all configuration
func (s *ConfigService) CreateConfigurationBackup(name, description string) (*models.ConfigurationBackup, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	backup := &models.ConfigurationBackup{
		BackupName:  name,
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Collect all configuration data
	hostConfig, _ := s.GetHostConfig()
	namingConfig, _ := s.GetNamingConfig()
	mediaConfig, _ := s.GetMediaManagementConfig()
	appSettings, _ := s.GetAppSettings()

	backup.HostConfig = hostConfig
	backup.NamingConfig = namingConfig
	backup.MediaManagementConfig = mediaConfig
	backup.AppSettings = appSettings

	// Get additional configurations if available
	if s.services != nil {
		if qualityProfiles, err := s.services.QualityService.GetQualityProfiles(); err == nil {
			// Convert from []*models.QualityProfile to []models.QualityProfile
			for _, profile := range qualityProfiles {
				if profile != nil {
					backup.QualityProfiles = append(backup.QualityProfiles, *profile)
				}
			}
		}
		if rootFolders, err := s.GetRootFolders(); err == nil {
			backup.RootFolders = rootFolders
		}
	}

	// Calculate backup statistics
	backup.ConfigurationCount = s.countConfigurations(backup)

	s.logger.Info("Created configuration backup", "name", name)
	return backup, nil
}

// RestoreConfigurationBackup restores configuration from a backup
func (s *ConfigService) RestoreConfigurationBackup(
	backup *models.ConfigurationBackup, components []string,
) (*models.ConfigurationImportResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	result := &models.ConfigurationImportResult{
		Success:             true,
		ImportedComponents:  []string{},
		SkippedComponents:   []string{},
		ValidationErrors:    make(map[string][]string),
		ConflictResolutions: make(map[string]string),
		ImportedAt:          time.Now(),
	}

	// Import each requested component
	for _, component := range components {
		s.restoreConfigurationComponent(backup, component, result)
	}

	s.logger.Info("Restored configuration backup",
		"imported", len(result.ImportedComponents),
		"skipped", len(result.SkippedComponents))
	return result, nil
}

// restoreConfigurationComponent restores a single configuration component
func (s *ConfigService) restoreConfigurationComponent(
	backup *models.ConfigurationBackup,
	component string,
	result *models.ConfigurationImportResult,
) {
	switch component {
	case "host":
		if backup.HostConfig != nil {
			if err := s.UpdateHostConfig(backup.HostConfig); err != nil {
				result.ValidationErrors["host"] = []string{err.Error()}
				result.Success = false
			} else {
				result.ImportedComponents = append(result.ImportedComponents, "host")
			}
		}
	case "naming":
		if backup.NamingConfig != nil {
			if err := s.UpdateNamingConfig(backup.NamingConfig); err != nil {
				result.ValidationErrors["naming"] = []string{err.Error()}
				result.Success = false
			} else {
				result.ImportedComponents = append(result.ImportedComponents, "naming")
			}
		}
	case "media":
		if backup.MediaManagementConfig != nil {
			if err := s.UpdateMediaManagementConfig(backup.MediaManagementConfig); err != nil {
				result.ValidationErrors["media"] = []string{err.Error()}
				result.Success = false
			} else {
				result.ImportedComponents = append(result.ImportedComponents, "media")
			}
		}
	case "app":
		if backup.AppSettings != nil {
			if err := s.UpdateAppSettings(backup.AppSettings); err != nil {
				result.ValidationErrors["app"] = []string{err.Error()}
				result.Success = false
			} else {
				result.ImportedComponents = append(result.ImportedComponents, "app")
			}
		}
	default:
		result.SkippedComponents = append(result.SkippedComponents, component)
	}
}

// FactoryResetConfiguration resets all configuration to default values
func (s *ConfigService) FactoryResetConfiguration(components []string) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	for _, component := range components {
		switch component {
		case "host":
			defaultConfig := models.GetDefaultHostConfig()
			if err := s.UpdateHostConfig(defaultConfig); err != nil {
				return fmt.Errorf("failed to reset host config: %w", err)
			}
		case "naming":
			defaultConfig := models.GetDefaultNamingConfig()
			if err := s.UpdateNamingConfig(defaultConfig); err != nil {
				return fmt.Errorf("failed to reset naming config: %w", err)
			}
		case "media":
			defaultConfig := models.GetDefaultMediaManagementConfig()
			if err := s.UpdateMediaManagementConfig(defaultConfig); err != nil {
				return fmt.Errorf("failed to reset media management config: %w", err)
			}
		case "app":
			defaultSettings := models.GetDefaultAppSettings()
			if err := s.UpdateAppSettings(defaultSettings); err != nil {
				return fmt.Errorf("failed to reset app settings: %w", err)
			}
		}
	}

	s.logger.Info("Performed factory reset", "components", components)
	return nil
}

// Configuration Validation

// ValidateAllConfigurations validates all configuration components
func (s *ConfigService) ValidateAllConfigurations() (*models.ConfigurationValidationResult, error) {
	result := &models.ConfigurationValidationResult{
		IsValid:          true,
		ValidationErrors: make(map[string][]string),
		Warnings:         make(map[string][]string),
		ComponentStatus:  make(map[string]models.ConfigurationStatus),
		TestResults:      make(map[string]models.ConfigurationTestResult),
	}

	// Validate all components
	s.validateHostConfig(result)
	s.validateNamingConfig(result)
	s.validateMediaManagementConfig(result)
	s.validateAppSettings(result)
	s.validateRootFolders(result)

	return result, nil
}

// validateHostConfig validates host configuration
func (s *ConfigService) validateHostConfig(result *models.ConfigurationValidationResult) {
	if hostConfig, err := s.GetHostConfig(); err == nil {
		if errors := hostConfig.ValidateConfiguration(); len(errors) > 0 {
			result.ValidationErrors["host"] = errors
			result.ComponentStatus["host"] = models.ConfigurationStatusError
			result.IsValid = false
		} else {
			result.ComponentStatus["host"] = models.ConfigurationStatusOK
		}
	}
}

// validateNamingConfig validates naming configuration
func (s *ConfigService) validateNamingConfig(result *models.ConfigurationValidationResult) {
	if namingConfig, err := s.GetNamingConfig(); err == nil {
		if errors := namingConfig.ValidateConfiguration(); len(errors) > 0 {
			result.ValidationErrors["naming"] = errors
			result.ComponentStatus["naming"] = models.ConfigurationStatusError
			result.IsValid = false
		} else {
			result.ComponentStatus["naming"] = models.ConfigurationStatusOK
		}
	}
}

// validateMediaManagementConfig validates media management configuration
func (s *ConfigService) validateMediaManagementConfig(result *models.ConfigurationValidationResult) {
	if mediaConfig, err := s.GetMediaManagementConfig(); err == nil {
		if errors := mediaConfig.ValidateConfiguration(); len(errors) > 0 {
			result.ValidationErrors["media"] = errors
			result.ComponentStatus["media"] = models.ConfigurationStatusError
			result.IsValid = false
		} else {
			result.ComponentStatus["media"] = models.ConfigurationStatusOK
		}
	}
}

// validateAppSettings validates app settings
func (s *ConfigService) validateAppSettings(result *models.ConfigurationValidationResult) {
	if appSettings, err := s.GetAppSettings(); err == nil {
		if errors := appSettings.ValidateSettings(); len(errors) > 0 {
			result.ValidationErrors["app"] = errors
			result.ComponentStatus["app"] = models.ConfigurationStatusError
			result.IsValid = false
		} else {
			result.ComponentStatus["app"] = models.ConfigurationStatusOK
		}
	}
}

// validateRootFolders validates root folders
func (s *ConfigService) validateRootFolders(result *models.ConfigurationValidationResult) {
	if rootFolders, err := s.GetRootFolders(); err == nil {
		for _, folder := range rootFolders {
			if errors := models.ValidateRootFolder(&folder); len(errors) > 0 {
				key := fmt.Sprintf("rootFolder_%d", folder.ID)
				result.ValidationErrors[key] = errors
				result.ComponentStatus[key] = models.ConfigurationStatusError
				result.IsValid = false
			} else if !folder.Accessible {
				key := fmt.Sprintf("rootFolder_%d", folder.ID)
				result.Warnings[key] = []string{"Root folder is not accessible"}
				result.ComponentStatus[key] = models.ConfigurationStatusWarning
			}
		}
	}
}

// Helper methods

// applyConfigToSettings applies configuration map to settings struct
func (s *ConfigService) applyConfigToSettings(configMap map[string]interface{}, settings *models.AppSettings) {
	// Implementation would map specific keys to struct fields
	if version, ok := configMap["app.version"].(string); ok {
		settings.Version = version
	}
	if initialized, ok := configMap["app.initialized"].(bool); ok {
		settings.Initialized = initialized
	}
	if theme, ok := configMap["ui.theme"].(string); ok {
		settings.Theme = theme
	}
	// Add more mappings as needed
}

// convertSettingsToConfigs converts settings struct to config entries
func (s *ConfigService) convertSettingsToConfigs(settings *models.AppSettings) []models.AppConfig {
	configs := []models.AppConfig{
		{Key: "app.version", Value: models.JSON{"value": settings.Version}, Description: "Application version"},
		{Key: "app.initialized", Value: models.JSON{"value": settings.Initialized},
			Description: "Application initialization status"},
		{Key: "ui.theme", Value: models.JSON{"value": settings.Theme}, Description: "UI theme preference"},
		{Key: "ui.language", Value: models.JSON{"value": settings.Language}, Description: "UI language preference"},
		{Key: "security.api_key_required", Value: models.JSON{"value": settings.APIKeyRequired},
			Description: "Whether API key is required"},
		{Key: "performance.max_concurrent_tasks",
			Value:       models.JSON{"value": settings.MaxConcurrentTasks},
			Description: "Maximum concurrent tasks"},
	}
	return configs
}

// countConfigurations counts the number of configuration components in a backup
func (s *ConfigService) countConfigurations(backup *models.ConfigurationBackup) int {
	count := 0
	if backup.HostConfig != nil {
		count++
	}
	if backup.NamingConfig != nil {
		count++
	}
	if backup.MediaManagementConfig != nil {
		count++
	}
	if backup.AppSettings != nil {
		count++
	}
	count += len(backup.QualityProfiles)
	count += len(backup.RootFolders)
	return count
}
