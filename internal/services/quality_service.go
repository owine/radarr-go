package services

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// QualityService provides operations for managing quality profiles and settings.
type QualityService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewQualityService creates a new instance of QualityService with the provided database and logger.
func NewQualityService(db *database.Database, logger *logger.Logger) *QualityService {
	return &QualityService{
		db:     db,
		logger: logger,
	}
}

// GetQualityProfiles retrieves all quality profiles from the system.
func (s *QualityService) GetQualityProfiles() ([]*models.QualityProfile, error) {
	var profiles []*models.QualityProfile

	if err := s.db.GORM.Find(&profiles).Error; err != nil {
		s.logger.Error("Failed to fetch quality profiles", "error", err)
		return nil, fmt.Errorf("failed to fetch quality profiles: %w", err)
	}

	return profiles, nil
}

// GetQualityProfileByID retrieves a specific quality profile by its ID.
func (s *QualityService) GetQualityProfileByID(id int) (*models.QualityProfile, error) {
	var profile models.QualityProfile

	if err := s.db.GORM.Where("id = ?", id).First(&profile).Error; err != nil {
		s.logger.Error("Failed to fetch quality profile", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch quality profile with id %d: %w", id, err)
	}

	return &profile, nil
}

// CreateQualityProfile creates a new quality profile.
func (s *QualityService) CreateQualityProfile(profile *models.QualityProfile) error {
	if err := s.db.GORM.Create(profile).Error; err != nil {
		s.logger.Error("Failed to create quality profile", "name", profile.Name, "error", err)
		return fmt.Errorf("failed to create quality profile: %w", err)
	}

	s.logger.Info("Created quality profile", "id", profile.ID, "name", profile.Name)
	return nil
}

// UpdateQualityProfile updates an existing quality profile.
func (s *QualityService) UpdateQualityProfile(profile *models.QualityProfile) error {
	if err := s.db.GORM.Save(profile).Error; err != nil {
		s.logger.Error("Failed to update quality profile", "id", profile.ID, "error", err)
		return fmt.Errorf("failed to update quality profile: %w", err)
	}

	s.logger.Info("Updated quality profile", "id", profile.ID, "name", profile.Name)
	return nil
}

// DeleteQualityProfile removes a quality profile.
func (s *QualityService) DeleteQualityProfile(id int) error {
	// Check if any movies are using this profile
	var count int64
	if err := s.db.GORM.Model(&models.Movie{}).Where("quality_profile_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check profile usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete quality profile: %d movies are using this profile", count)
	}

	result := s.db.GORM.Delete(&models.QualityProfile{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete quality profile", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete quality profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("quality profile with id %d not found", id)
	}

	s.logger.Info("Deleted quality profile", "id", id)
	return nil
}

// GetQualityDefinitions retrieves all quality definitions.
func (s *QualityService) GetQualityDefinitions() ([]*models.QualityLevel, error) {
	var definitions []*models.QualityLevel

	if err := s.db.GORM.Order("weight ASC").Find(&definitions).Error; err != nil {
		s.logger.Error("Failed to fetch quality definitions", "error", err)
		return nil, fmt.Errorf("failed to fetch quality definitions: %w", err)
	}

	return definitions, nil
}

// GetQualityDefinitionByID retrieves a specific quality definition by its ID.
func (s *QualityService) GetQualityDefinitionByID(id int) (*models.QualityLevel, error) {
	var definition models.QualityLevel

	if err := s.db.GORM.Where("id = ?", id).First(&definition).Error; err != nil {
		s.logger.Error("Failed to fetch quality definition", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch quality definition with id %d: %w", id, err)
	}

	return &definition, nil
}

// UpdateQualityDefinition updates an existing quality definition.
func (s *QualityService) UpdateQualityDefinition(definition *models.QualityLevel) error {
	if err := s.db.GORM.Save(definition).Error; err != nil {
		s.logger.Error("Failed to update quality definition", "id", definition.ID, "error", err)
		return fmt.Errorf("failed to update quality definition: %w", err)
	}

	s.logger.Info("Updated quality definition", "id", definition.ID, "title", definition.Title)
	return nil
}

// GetCustomFormats retrieves all custom formats.
func (s *QualityService) GetCustomFormats() ([]*models.CustomFormat, error) {
	var formats []*models.CustomFormat

	if err := s.db.GORM.Find(&formats).Error; err != nil {
		s.logger.Error("Failed to fetch custom formats", "error", err)
		return nil, fmt.Errorf("failed to fetch custom formats: %w", err)
	}

	return formats, nil
}

// GetCustomFormatByID retrieves a specific custom format by its ID.
func (s *QualityService) GetCustomFormatByID(id int) (*models.CustomFormat, error) {
	var format models.CustomFormat

	if err := s.db.GORM.Where("id = ?", id).First(&format).Error; err != nil {
		s.logger.Error("Failed to fetch custom format", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch custom format with id %d: %w", id, err)
	}

	return &format, nil
}

// CreateCustomFormat creates a new custom format.
func (s *QualityService) CreateCustomFormat(format *models.CustomFormat) error {
	if err := s.db.GORM.Create(format).Error; err != nil {
		s.logger.Error("Failed to create custom format", "name", format.Name, "error", err)
		return fmt.Errorf("failed to create custom format: %w", err)
	}

	s.logger.Info("Created custom format", "id", format.ID, "name", format.Name)
	return nil
}

// UpdateCustomFormat updates an existing custom format.
func (s *QualityService) UpdateCustomFormat(format *models.CustomFormat) error {
	if err := s.db.GORM.Save(format).Error; err != nil {
		s.logger.Error("Failed to update custom format", "id", format.ID, "error", err)
		return fmt.Errorf("failed to update custom format: %w", err)
	}

	s.logger.Info("Updated custom format", "id", format.ID, "name", format.Name)
	return nil
}

// DeleteCustomFormat removes a custom format.
func (s *QualityService) DeleteCustomFormat(id int) error {
	result := s.db.GORM.Delete(&models.CustomFormat{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete custom format", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete custom format: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("custom format with id %d not found", id)
	}

	s.logger.Info("Deleted custom format", "id", id)
	return nil
}

// InitializeQualityDefinitions ensures default quality definitions exist.
func (s *QualityService) InitializeQualityDefinitions() error {
	// Check if quality definitions already exist
	var count int64
	if err := s.db.GORM.Model(&models.QualityLevel{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check quality definitions: %w", err)
	}

	if count > 0 {
		s.logger.Debug("Quality definitions already exist, skipping initialization")
		return nil
	}

	// Insert default quality definitions
	defaults := models.DefaultQualityDefinitions()
	for _, def := range defaults {
		if err := s.db.GORM.Create(def).Error; err != nil {
			s.logger.Error("Failed to create default quality definition", "id", def.ID, "title", def.Title, "error", err)
			return fmt.Errorf("failed to create default quality definition: %w", err)
		}
	}

	s.logger.Info("Initialized default quality definitions", "count", len(defaults))
	return nil
}
