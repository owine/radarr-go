package services

import (
	"errors"
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/hints"
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
	err := s.db.GORM.
		Clauses(hints.UseIndex("idx_movie_search")).
		Preload("MovieFile").
		Where("title LIKE ? OR original_title LIKE ? OR clean_title LIKE ?",
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

// CreateWithFile creates a new movie and its associated file in a transaction
func (s *MovieService) CreateWithFile(movie *models.Movie, file *models.MovieFile) error {
	return s.db.GORM.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(movie).Error; err != nil {
			s.logger.Error("Failed to create movie in transaction", "title", movie.Title, "error", err)
			return fmt.Errorf("failed to create movie: %w", err)
		}

		if err := s.handleMovieFileInTransaction(tx, movie, file, "create"); err != nil {
			return err
		}

		s.logger.Info("Created movie with file in transaction", "id", movie.ID, "title", movie.Title)
		return nil
	})
}

// handleMovieFileInTransaction handles movie file operations within a transaction
func (s *MovieService) handleMovieFileInTransaction(
	tx *gorm.DB, movie *models.Movie, file *models.MovieFile, operation string,
) error {
	if file == nil {
		return nil
	}

	file.MovieID = movie.ID
	var err error

	switch operation {
	case "create":
		err = tx.Create(file).Error
		if err != nil {
			s.logger.Error("Failed to create movie file in transaction", "movieId", movie.ID, "error", err)
			return fmt.Errorf("failed to create movie file: %w", err)
		}
	case "save":
		err = tx.Save(file).Error
		if err != nil {
			s.logger.Error("Failed to save movie file in transaction", "movieId", movie.ID, "error", err)
			return fmt.Errorf("failed to save movie file: %w", err)
		}
	}

	movie.MovieFileID = file.ID
	movie.HasFile = true
	if err := tx.Save(movie).Error; err != nil {
		s.logger.Error("Failed to update movie with file ID", "movieId", movie.ID, "error", err)
		return fmt.Errorf("failed to update movie with file: %w", err)
	}

	return nil
}

// UpdateWithFile updates a movie and creates/updates its associated file in a transaction
func (s *MovieService) UpdateWithFile(movie *models.Movie, file *models.MovieFile) error {
	return s.db.GORM.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(movie).Error; err != nil {
			s.logger.Error("Failed to update movie in transaction", "id", movie.ID, "error", err)
			return fmt.Errorf("failed to update movie: %w", err)
		}

		if err := s.handleMovieFileInTransaction(tx, movie, file, "save"); err != nil {
			return err
		}

		s.logger.Info("Updated movie with file in transaction", "id", movie.ID, "title", movie.Title)
		return nil
	})
}

// DeleteWithFile removes a movie and its associated file in a transaction
func (s *MovieService) DeleteWithFile(id int) error {
	return s.db.GORM.Transaction(func(tx *gorm.DB) error {
		var movie models.Movie
		if err := tx.First(&movie, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("movie not found: %w", err)
			}
			return fmt.Errorf("failed to find movie: %w", err)
		}

		// Delete associated movie file if exists
		if movie.MovieFileID > 0 {
			if err := tx.Delete(&models.MovieFile{}, movie.MovieFileID).Error; err != nil {
				s.logger.Error("Failed to delete movie file in transaction", "fileId", movie.MovieFileID, "error", err)
				return fmt.Errorf("failed to delete movie file: %w", err)
			}
		}

		if err := tx.Delete(&movie).Error; err != nil {
			s.logger.Error("Failed to delete movie in transaction", "id", id, "error", err)
			return fmt.Errorf("failed to delete movie: %w", err)
		}

		s.logger.Info("Deleted movie with file in transaction", "id", id, "title", movie.Title)
		return nil
	})
}

// GetTotalCount returns the total number of movies in the database
func (s *MovieService) GetTotalCount() (int64, error) {
	var count int64
	err := s.db.GORM.Model(&models.Movie{}).Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get movie count", "error", err)
		return 0, fmt.Errorf("failed to get movie count: %w", err)
	}
	return count, nil
}

// GetMonitoredCount returns the count of monitored movies
func (s *MovieService) GetMonitoredCount() (int64, error) {
	var count int64
	err := s.db.GORM.Model(&models.Movie{}).Where("monitored = ?", true).Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get monitored movie count", "error", err)
		return 0, fmt.Errorf("failed to get monitored movie count: %w", err)
	}
	return count, nil
}

// GetMoviesWithFilesCount returns the count of movies with files
func (s *MovieService) GetMoviesWithFilesCount() (int64, error) {
	var count int64
	err := s.db.GORM.Model(&models.Movie{}).Where("has_file = ?", true).Count(&count).Error
	if err != nil {
		s.logger.Error("Failed to get movies with files count", "error", err)
		return 0, fmt.Errorf("failed to get movies with files count: %w", err)
	}
	return count, nil
}

// AssignToCollection assigns a movie to a collection
func (s *MovieService) AssignToCollection(movieID int, collectionTmdbID int) error {
	err := s.db.GORM.Model(&models.Movie{}).
		Where("id = ?", movieID).
		Update("collection_tmdb_id", collectionTmdbID).Error

	if err != nil {
		s.logger.Error("Failed to assign movie to collection", "movieId", movieID, "collectionTmdbId", collectionTmdbID, "error", err)
		return fmt.Errorf("failed to assign movie to collection: %w", err)
	}

	s.logger.Info("Assigned movie to collection", "movieId", movieID, "collectionTmdbId", collectionTmdbID)
	return nil
}

// RemoveFromCollection removes a movie from its collection
func (s *MovieService) RemoveFromCollection(movieID int) error {
	err := s.db.GORM.Model(&models.Movie{}).
		Where("id = ?", movieID).
		Update("collection_tmdb_id", nil).Error

	if err != nil {
		s.logger.Error("Failed to remove movie from collection", "movieId", movieID, "error", err)
		return fmt.Errorf("failed to remove movie from collection: %w", err)
	}

	s.logger.Info("Removed movie from collection", "movieId", movieID)
	return nil
}

// GetMoviesByCollection retrieves all movies in a collection
func (s *MovieService) GetMoviesByCollection(collectionTmdbID int) ([]models.Movie, error) {
	var movies []models.Movie

	err := s.db.GORM.Preload("MovieFile").
		Where("collection_tmdb_id = ?", collectionTmdbID).
		Find(&movies).Error

	if err != nil {
		s.logger.Error("Failed to get movies by collection", "collectionTmdbId", collectionTmdbID, "error", err)
		return nil, fmt.Errorf("failed to get movies by collection: %w", err)
	}

	return movies, nil
}

// AutoAssignToCollections automatically assigns movies to collections based on TMDB metadata
func (s *MovieService) AutoAssignToCollections() error {
	// This would be called periodically to automatically assign movies to collections
	// based on their TMDB collection information
	var movies []models.Movie

	// Get movies that have collection info but aren't assigned yet
	err := s.db.GORM.Where("collection IS NOT NULL AND collection_tmdb_id IS NULL").Find(&movies).Error
	if err != nil {
		s.logger.Error("Failed to fetch movies for auto-assignment", "error", err)
		return fmt.Errorf("failed to fetch movies: %w", err)
	}

	assigned := 0
	for _, movie := range movies {
		if movie.Collection != nil && movie.Collection.TmdbID > 0 {
			if err := s.AssignToCollection(movie.ID, movie.Collection.TmdbID); err != nil {
				s.logger.Warn("Failed to auto-assign movie to collection", "movieId", movie.ID, "collectionTmdbId", movie.Collection.TmdbID, "error", err)
			} else {
				assigned++
			}
		}
	}

	s.logger.Info("Auto-assigned movies to collections", "assigned", assigned, "total", len(movies))
	return nil
}
