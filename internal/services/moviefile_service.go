package services

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

type MovieFileService struct {
	db     *database.Database
	logger *logger.Logger
}

func NewMovieFileService(db *database.Database, logger *logger.Logger) *MovieFileService {
	return &MovieFileService{
		db:     db,
		logger: logger,
	}
}

func (s *MovieFileService) GetAll() ([]models.MovieFile, error) {
	var movieFiles []models.MovieFile

	err := s.db.GORM.Find(&movieFiles).Error
	if err != nil {
		s.logger.Error("Failed to get all movie files", "error", err)
		return nil, fmt.Errorf("failed to get movie files: %w", err)
	}

	return movieFiles, nil
}

func (s *MovieFileService) GetByID(id int) (*models.MovieFile, error) {
	var movieFile models.MovieFile

	err := s.db.GORM.Where("id = ?", id).First(&movieFile).Error
	if err != nil {
		s.logger.Error("Failed to get movie file by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get movie file: %w", err)
	}

	return &movieFile, nil
}

func (s *MovieFileService) GetByMovieID(movieID int) ([]models.MovieFile, error) {
	var movieFiles []models.MovieFile

	err := s.db.GORM.Where("movie_id = ?", movieID).Find(&movieFiles).Error
	if err != nil {
		s.logger.Error("Failed to get movie files by movie ID", "movieID", movieID, "error", err)
		return nil, fmt.Errorf("failed to get movie files: %w", err)
	}

	return movieFiles, nil
}

func (s *MovieFileService) Create(movieFile *models.MovieFile) error {
	err := s.db.GORM.Create(movieFile).Error
	if err != nil {
		s.logger.Error("Failed to create movie file", "path", movieFile.Path, "error", err)
		return fmt.Errorf("failed to create movie file: %w", err)
	}

	s.logger.Info("Created movie file", "id", movieFile.ID, "path", movieFile.Path)
	return nil
}

func (s *MovieFileService) Update(movieFile *models.MovieFile) error {
	err := s.db.GORM.Save(movieFile).Error
	if err != nil {
		s.logger.Error("Failed to update movie file", "id", movieFile.ID, "error", err)
		return fmt.Errorf("failed to update movie file: %w", err)
	}

	s.logger.Info("Updated movie file", "id", movieFile.ID, "path", movieFile.Path)
	return nil
}

func (s *MovieFileService) Delete(id int) error {
	err := s.db.GORM.Delete(&models.MovieFile{}, id).Error
	if err != nil {
		s.logger.Error("Failed to delete movie file", "id", id, "error", err)
		return fmt.Errorf("failed to delete movie file: %w", err)
	}

	s.logger.Info("Deleted movie file", "id", id)
	return nil
}

func (s *MovieFileService) GetByPath(path string) (*models.MovieFile, error) {
	var movieFile models.MovieFile

	err := s.db.GORM.Where("path = ?", path).First(&movieFile).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get movie file by path: %w", err)
	}

	return &movieFile, nil
}
