// Package services provides business logic for the Radarr application.
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// RenameService handles file renaming operations
type RenameService struct {
	db            *database.Database
	logger        *logger.Logger
	namingService *NamingService
}

// NewRenameService creates a new rename service
func NewRenameService(db *database.Database, logger *logger.Logger, namingService *NamingService) *RenameService {
	return &RenameService{
		db:            db,
		logger:        logger,
		namingService: namingService,
	}
}

// PreviewRename generates a preview of file renames for given movies
func (s *RenameService) PreviewRename(ctx context.Context, movieIDs []int) ([]*models.RenamePreview, error) {
	previews := make([]*models.RenamePreview, 0)

	for _, movieID := range movieIDs {
		moviePreviews, err := s.previewMovieRename(ctx, movieID)
		if err != nil {
			s.logger.Warn("Failed to preview rename for movie", "movieId", movieID, "error", err)
			continue
		}
		previews = append(previews, moviePreviews...)
	}

	s.logger.Info("Generated rename previews", "totalFiles", len(previews))
	return previews, nil
}

// RenameMovies performs the actual file renaming for given movies
func (s *RenameService) RenameMovies(ctx context.Context, movieIDs []int) error {
	successCount := 0
	errorCount := 0

	for _, movieID := range movieIDs {
		if err := s.renameMovie(ctx, movieID); err != nil {
			s.logger.Error("Failed to rename movie files", "movieId", movieID, "error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("Completed movie renaming", "success", successCount, "errors", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to rename %d out of %d movies", errorCount, len(movieIDs))
	}

	return nil
}

// PreviewMovieFolderRename generates a preview of folder renames for movies
func (s *RenameService) PreviewMovieFolderRename(ctx context.Context, movieIDs []int) ([]*models.RenamePreview, error) {
	previews := make([]*models.RenamePreview, 0)

	for _, movieID := range movieIDs {
		preview, err := s.previewMovieFolderRename(ctx, movieID)
		if err != nil {
			s.logger.Warn("Failed to preview folder rename for movie", "movieId", movieID, "error", err)
			continue
		}
		if preview != nil {
			previews = append(previews, preview)
		}
	}

	s.logger.Info("Generated folder rename previews", "count", len(previews))
	return previews, nil
}

// RenameMovieFolders performs the actual folder renaming for given movies
func (s *RenameService) RenameMovieFolders(ctx context.Context, movieIDs []int) error {
	successCount := 0
	errorCount := 0

	for _, movieID := range movieIDs {
		if err := s.renameMovieFolder(ctx, movieID); err != nil {
			s.logger.Error("Failed to rename movie folder", "movieId", movieID, "error", err)
			errorCount++
		} else {
			successCount++
		}
	}

	s.logger.Info("Completed movie folder renaming", "success", successCount, "errors", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to rename %d out of %d movie folders", errorCount, len(movieIDs))
	}

	return nil
}

// previewMovieRename generates rename previews for a single movie's files
func (s *RenameService) previewMovieRename(ctx context.Context, movieID int) ([]*models.RenamePreview, error) {
	var movie models.Movie
	if err := s.db.GORM.WithContext(ctx).
		Preload("MovieFile").
		First(&movie, movieID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %w", err)
	}

	previews := make([]*models.RenamePreview, 0)

	// Get movie file if exists
	if movie.MovieFile != nil && movie.MovieFile.RelativePath != "" {
		newPath, err := s.generateFileName(ctx, &movie, movie.MovieFile)
		if err != nil {
			return nil, fmt.Errorf("failed to generate new file name: %w", err)
		}

		existingPath := filepath.Join(movie.Path, movie.MovieFile.RelativePath)

		// Only add to preview if the name would actually change
		if newPath != existingPath {
			previews = append(previews, &models.RenamePreview{
				MovieID:      movieID,
				MovieFileID:  movie.MovieFileID,
				ExistingPath: existingPath,
				NewPath:      newPath,
			})
		}
	}

	return previews, nil
}

// previewMovieFolderRename generates a folder rename preview for a single movie
func (s *RenameService) previewMovieFolderRename(ctx context.Context, movieID int) (*models.RenamePreview, error) {
	var movie models.Movie
	if err := s.db.GORM.WithContext(ctx).First(&movie, movieID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %w", err)
	}

	newFolderName, err := s.generateFolderName(ctx, &movie)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new folder name: %w", err)
	}

	newPath := filepath.Join(movie.RootFolderPath, newFolderName)

	// Only return preview if the folder name would actually change
	if newPath != movie.Path {
		return &models.RenamePreview{
			MovieID:      movieID,
			ExistingPath: movie.Path,
			NewPath:      newPath,
		}, nil
	}

	return nil, nil
}

// renameMovie performs the actual file renaming for a single movie
func (s *RenameService) renameMovie(ctx context.Context, movieID int) error {
	var movie models.Movie
	if err := s.db.GORM.WithContext(ctx).
		Preload("MovieFile").
		First(&movie, movieID).Error; err != nil {
		return fmt.Errorf("failed to fetch movie: %w", err)
	}

	if movie.MovieFile == nil || movie.MovieFile.RelativePath == "" {
		s.logger.Debug("No movie file to rename", "movieId", movieID)
		return nil
	}

	newPath, err := s.generateFileName(ctx, &movie, movie.MovieFile)
	if err != nil {
		return fmt.Errorf("failed to generate new file name: %w", err)
	}

	existingPath := filepath.Join(movie.Path, movie.MovieFile.RelativePath)

	// Skip if no change needed
	if newPath == existingPath {
		s.logger.Debug("File name unchanged", "movieId", movieID)
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(newPath), 0750); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Perform the rename
	if err := os.Rename(existingPath, newPath); err != nil {
		return fmt.Errorf("failed to rename file from %s to %s: %w", existingPath, newPath, err)
	}

	// Update database with new path
	newRelativePath, err := filepath.Rel(movie.Path, newPath)
	if err != nil {
		newRelativePath = filepath.Base(newPath)
	}

	if err := s.db.GORM.WithContext(ctx).
		Model(movie.MovieFile).
		Update("relative_path", newRelativePath).Error; err != nil {
		s.logger.Error("Failed to update movie file path in database", "movieId", movieID, "error", err)
		// Don't return error as the physical rename was successful
	}

	s.logger.Info("Renamed movie file", "movieId", movieID, "from", existingPath, "to", newPath)
	return nil
}

// renameMovieFolder performs the actual folder renaming for a single movie
func (s *RenameService) renameMovieFolder(ctx context.Context, movieID int) error {
	var movie models.Movie
	if err := s.db.GORM.WithContext(ctx).
		Preload("MovieFile").
		First(&movie, movieID).Error; err != nil {
		return fmt.Errorf("failed to fetch movie: %w", err)
	}

	newFolderName, err := s.generateFolderName(ctx, &movie)
	if err != nil {
		return fmt.Errorf("failed to generate new folder name: %w", err)
	}

	newPath := filepath.Join(movie.RootFolderPath, newFolderName)

	// Skip if no change needed
	if newPath == movie.Path {
		s.logger.Debug("Folder name unchanged", "movieId", movieID)
		return nil
	}

	// Perform the rename
	if err := os.Rename(movie.Path, newPath); err != nil {
		return fmt.Errorf("failed to rename folder from %s to %s: %w", movie.Path, newPath, err)
	}

	// Update database with new path
	if err := s.db.GORM.WithContext(ctx).
		Model(&movie).
		Updates(map[string]interface{}{
			"path":        newPath,
			"folder_name": newFolderName,
		}).Error; err != nil {
		s.logger.Error("Failed to update movie path in database", "movieId", movieID, "error", err)
		// Don't return error as the physical rename was successful
	}

	s.logger.Info("Renamed movie folder", "movieId", movieID, "from", movie.Path, "to", newPath)
	return nil
}

// generateFileName generates a new file name based on naming configuration
func (s *RenameService) generateFileName(ctx context.Context, movie *models.Movie, movieFile *models.MovieFile) (string, error) {
	namingConfig, err := s.namingService.GetNamingConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get naming config: %w", err)
	}

	if !namingConfig.RenameMovies || namingConfig.StandardMovieFormat == "" {
		// Return existing path if renaming is disabled
		return filepath.Join(movie.Path, movieFile.RelativePath), nil
	}

	// Parse the naming template
	template := namingConfig.StandardMovieFormat
	fileName, err := s.parseNamingTemplate(template, movie, movieFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse naming template: %w", err)
	}

	// Clean the filename
	fileName = s.cleanFileName(fileName)

	// Add file extension
	ext := filepath.Ext(movieFile.RelativePath)
	if !strings.HasSuffix(fileName, ext) {
		fileName += ext
	}

	return filepath.Join(movie.Path, fileName), nil
}

// generateFolderName generates a new folder name based on naming configuration
func (s *RenameService) generateFolderName(ctx context.Context, movie *models.Movie) (string, error) {
	namingConfig, err := s.namingService.GetNamingConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get naming config: %w", err)
	}

	if namingConfig.MovieFolderFormat == "" {
		// Use default format if none specified
		template := "{Movie Title} ({Release Year})"
		folderName, err := s.parseNamingTemplate(template, movie, nil)
		if err != nil {
			return "", fmt.Errorf("failed to parse default naming template: %w", err)
		}
		return s.cleanFileName(folderName), nil
	}

	// Parse the naming template
	folderName, err := s.parseNamingTemplate(namingConfig.MovieFolderFormat, movie, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse naming template: %w", err)
	}

	return s.cleanFileName(folderName), nil
}

// parseNamingTemplate parses a naming template and replaces tokens with actual values
func (s *RenameService) parseNamingTemplate(template string, movie *models.Movie, movieFile *models.MovieFile) (string, error) {
	result := template

	// Movie-based replacements
	result = strings.ReplaceAll(result, "{Movie Title}", movie.Title)
	result = strings.ReplaceAll(result, "{Movie CleanTitle}", movie.CleanTitle)
	result = strings.ReplaceAll(result, "{Movie TitleThe}", s.moveArticleToEnd(movie.Title))
	result = strings.ReplaceAll(result, "{Release Year}", strconv.Itoa(movie.Year))
	result = strings.ReplaceAll(result, "{ImDb Id}", movie.ImdbID)
	result = strings.ReplaceAll(result, "{Tmdb Id}", strconv.Itoa(movie.TmdbID))

	// Quality and format replacements (if movieFile provided)
	if movieFile != nil {
		result = strings.ReplaceAll(result, "{Quality Title}", movieFile.Quality.Quality.Name)
		result = strings.ReplaceAll(result, "{Quality Proper}", s.getQualityProperString(movieFile.Quality))
		result = strings.ReplaceAll(result, "{MediaInfo Simple}", s.getMediaInfoSimple(movieFile))

		// File-specific replacements
		result = strings.ReplaceAll(result, "{Original Title}", movieFile.SceneName)
		result = strings.ReplaceAll(result, "{Original Filename}", s.getOriginalFilename(movieFile))
	}

	// Collection replacements
	if movie.Collection != nil {
		result = strings.ReplaceAll(result, "{Movie Collection}", movie.Collection.Name)
	} else {
		result = strings.ReplaceAll(result, "{Movie Collection}", "")
	}

	// Date replacements
	now := time.Now()
	result = strings.ReplaceAll(result, "{Release Date}", s.formatDate(movie.InCinemas))
	result = strings.ReplaceAll(result, "{Air Date}", s.formatDate(movie.PhysicalRelease))
	result = strings.ReplaceAll(result, "{Today}", now.Format("2006-01-02"))

	// Clean up any double spaces or empty brackets
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = regexp.MustCompile(`\(\s*\)`).ReplaceAllString(result, "")
	result = regexp.MustCompile(`\[\s*\]`).ReplaceAllString(result, "")
	result = strings.TrimSpace(result)

	return result, nil
}

// cleanFileName cleans a filename by removing/replacing illegal characters
func (s *RenameService) cleanFileName(filename string) string {
	// Replace illegal characters based on OS
	illegalChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range illegalChars {
		filename = strings.ReplaceAll(filename, char, "")
	}

	// Replace forward and back slashes (except as path separators)
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Trim periods and spaces from the end
	filename = strings.TrimRight(filename, ". ")

	return filename
}

// moveArticleToEnd moves articles (The, A, An) from the beginning to the end
func (s *RenameService) moveArticleToEnd(title string) string {
	articles := []string{"The ", "A ", "An "}
	for _, article := range articles {
		if strings.HasPrefix(title, article) {
			return strings.TrimSpace(strings.TrimPrefix(title, article)) + ", " + strings.TrimSpace(article)
		}
	}
	return title
}

// getQualityProperString returns proper/repack information
func (s *RenameService) getQualityProperString(quality models.Quality) string {
	if quality.Revision.Version > 1 {
		return fmt.Sprintf("PROPER%d", quality.Revision.Version-1)
	}
	return ""
}

// getMediaInfoSimple returns simplified media info
func (s *RenameService) getMediaInfoSimple(movieFile *models.MovieFile) string {
	// This would extract resolution, codec, etc. from MediaInfo
	// For now, return empty string as a placeholder
	return ""
}

// getOriginalFilename returns the original filename without extension
func (s *RenameService) getOriginalFilename(movieFile *models.MovieFile) string {
	filename := filepath.Base(movieFile.RelativePath)
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// formatDate formats a date pointer to string
func (s *RenameService) formatDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format("2006-01-02")
}
