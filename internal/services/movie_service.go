package services

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// MovieService provides operations for managing movies in the database.
type MovieService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewMovieService creates a new instance of MovieService with the provided database and logger.
func NewMovieService(db *database.Database, logger *logger.Logger) *MovieService {
	return &MovieService{
		db:     db,
		logger: logger,
	}
}

// GetAll retrieves all movies from the database with their movie files preloaded.
func (s *MovieService) GetAll() ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Preload("MovieFile").Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to get all movies", "error", err)
		return nil, fmt.Errorf("failed to get movies: %w", err)
	}

	return movies, nil
}

// GetByID retrieves a single movie by its ID with movie file preloaded.
func (s *MovieService) GetByID(id int) (*models.Movie, error) {
	var movie models.Movie

	err := s.db.GORM.Preload("MovieFile").Where("id = ?", id).First(&movie).Error
	if err != nil {
		s.logger.Error("Failed to get movie by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	return &movie, nil
}

// GetByTmdbID retrieves a movie by its TMDB ID with movie file preloaded.
func (s *MovieService) GetByTmdbID(tmdbID int) (*models.Movie, error) {
	var movie models.Movie

	err := s.db.GORM.Preload("MovieFile").Where("tmdb_id = ?", tmdbID).First(&movie).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get movie by TMDB ID: %w", err)
	}

	return &movie, nil
}

// Create creates a new movie in the database.
func (s *MovieService) Create(movie *models.Movie) error {
	err := s.db.GORM.Create(movie).Error
	if err != nil {
		s.logger.Error("Failed to create movie", "title", movie.Title, "error", err)
		return fmt.Errorf("failed to create movie: %w", err)
	}

	s.logger.Info("Created movie", "id", movie.ID, "title", movie.Title)
	return nil
}

// Update saves changes to an existing movie in the database.
func (s *MovieService) Update(movie *models.Movie) error {
	err := s.db.GORM.Save(movie).Error
	if err != nil {
		s.logger.Error("Failed to update movie", "id", movie.ID, "error", err)
		return fmt.Errorf("failed to update movie: %w", err)
	}

	s.logger.Info("Updated movie", "id", movie.ID, "title", movie.Title)
	return nil
}

// Delete removes a movie from the database by its ID.
func (s *MovieService) Delete(id int) error {
	err := s.db.GORM.Delete(&models.Movie{}, id).Error
	if err != nil {
		s.logger.Error("Failed to delete movie", "id", id, "error", err)
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	s.logger.Info("Deleted movie", "id", id)
	return nil
}

// Search finds movies by searching title, original title, and clean title fields.
func (s *MovieService) Search(query string) ([]models.Movie, error) {
	var movies []models.Movie

	searchQuery := "%" + query + "%"
	err := s.db.GORM.Preload("MovieFile").Where(
		"title LIKE ? OR original_title LIKE ? OR clean_title LIKE ?",
		searchQuery, searchQuery, searchQuery,
	).Find(&movies).Error

	if err != nil {
		s.logger.Error("Failed to search movies", "query", query, "error", err)
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	return movies, nil
}

// GetMonitored retrieves all movies that are currently being monitored.
func (s *MovieService) GetMonitored() ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Preload("MovieFile").Where("monitored = ?", true).Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to get monitored movies", "error", err)
		return nil, fmt.Errorf("failed to get monitored movies: %w", err)
	}

	return movies, nil
}

// GetUnmonitored retrieves all movies that are not being monitored.
func (s *MovieService) GetUnmonitored() ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Preload("MovieFile").Where("monitored = ?", false).Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to get unmonitored movies", "error", err)
		return nil, fmt.Errorf("failed to get unmonitored movies: %w", err)
	}

	return movies, nil
}

// GetMoviesWithoutFiles retrieves all monitored movies that don't have associated files.
func (s *MovieService) GetMoviesWithoutFiles() ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Where("has_file = ? AND monitored = ?", false, true).Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to get movies without files", "error", err)
		return nil, fmt.Errorf("failed to get movies without files: %w", err)
	}

	return movies, nil
}

// GetMoviesWithFiles retrieves all movies that have associated files with movie files preloaded.
func (s *MovieService) GetMoviesWithFiles() ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Preload("MovieFile").Where("has_file = ?", true).Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to get movies with files", "error", err)
		return nil, fmt.Errorf("failed to get movies with files: %w", err)
	}

	return movies, nil
}
