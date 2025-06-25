package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/tmdb"
)

// MetadataService handles movie metadata operations
type MetadataService struct {
	db     *database.Database
	tmdb   *tmdb.Client
	logger *logger.Logger
}

// NewMetadataService creates a new metadata service
func NewMetadataService(db *database.Database, cfg *config.Config, logger *logger.Logger) *MetadataService {
	tmdbClient := tmdb.NewClient(cfg, logger)

	return &MetadataService{
		db:     db,
		tmdb:   tmdbClient,
		logger: logger,
	}
}

// SearchMovies searches for movies using TMDB
func (s *MetadataService) SearchMovies(query string, page int) (*tmdb.SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if page <= 0 {
		page = 1
	}

	s.logger.Debug("Searching for movies", "query", query, "page", page)

	response, err := s.tmdb.SearchMovies(query, page)
	if err != nil {
		s.logger.Error("Failed to search movies", "query", query, "error", err)
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	s.logger.Debug("Movie search completed", "query", query, "results", len(response.Results))
	return response, nil
}

// LookupMovieByTMDBID retrieves detailed movie information from TMDB
func (s *MetadataService) LookupMovieByTMDBID(tmdbID int) (*models.Movie, error) {
	if tmdbID <= 0 {
		return nil, fmt.Errorf("invalid TMDB ID: %d", tmdbID)
	}

	s.logger.Debug("Looking up movie by TMDB ID", "tmdbId", tmdbID)

	// Get movie details from TMDB
	tmdbMovie, err := s.tmdb.GetMovie(tmdbID)
	if err != nil {
		s.logger.Error("Failed to get movie from TMDB", "tmdbId", tmdbID, "error", err)
		return nil, fmt.Errorf("failed to get movie from TMDB: %w", err)
	}

	// Get credits for cast/crew information
	credits, err := s.tmdb.GetCredits(tmdbID)
	if err != nil {
		s.logger.Warn("Failed to get movie credits", "tmdbId", tmdbID, "error", err)
		// Continue without credits - not critical
	}

	// Convert TMDB movie to internal movie model
	movie := s.convertTMDBToMovie(tmdbMovie, credits)

	s.logger.Debug("Movie lookup completed", "tmdbId", tmdbID, "title", movie.Title)
	return movie, nil
}

// RefreshMovieMetadata updates movie metadata from TMDB
func (s *MetadataService) RefreshMovieMetadata(movieID int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Get existing movie from database
	var existingMovie models.Movie
	if err := s.db.GORM.First(&existingMovie, movieID).Error; err != nil {
		return fmt.Errorf("movie not found: %w", err)
	}

	if existingMovie.TmdbID == 0 {
		return fmt.Errorf("movie has no TMDB ID for metadata refresh")
	}

	s.logger.Info("Refreshing movie metadata", "movieId", movieID, "tmdbId", existingMovie.TmdbID)

	// Get updated metadata from TMDB
	updatedMovie, err := s.LookupMovieByTMDBID(existingMovie.TmdbID)
	if err != nil {
		return fmt.Errorf("failed to lookup updated metadata: %w", err)
	}

	// Preserve important fields that shouldn't be overwritten
	updatedMovie.ID = existingMovie.ID
	updatedMovie.MovieFileID = existingMovie.MovieFileID
	updatedMovie.HasFile = existingMovie.HasFile
	updatedMovie.Monitored = existingMovie.Monitored
	updatedMovie.Path = existingMovie.Path
	updatedMovie.RootFolderPath = existingMovie.RootFolderPath
	updatedMovie.QualityProfileID = existingMovie.QualityProfileID
	updatedMovie.Added = existingMovie.Added

	// Update the movie in database
	if err := s.db.GORM.Save(updatedMovie).Error; err != nil {
		return fmt.Errorf("failed to save updated movie: %w", err)
	}

	s.logger.Info("Movie metadata refreshed successfully", "movieId", movieID, "title", updatedMovie.Title)
	return nil
}

// GetPopularMovies retrieves popular movies from TMDB
func (s *MetadataService) GetPopularMovies(page int) (*tmdb.SearchResponse, error) {
	if page <= 0 {
		page = 1
	}

	s.logger.Debug("Getting popular movies", "page", page)

	response, err := s.tmdb.GetPopular(page)
	if err != nil {
		s.logger.Error("Failed to get popular movies", "error", err)
		return nil, fmt.Errorf("failed to get popular movies: %w", err)
	}

	s.logger.Debug("Popular movies retrieved", "count", len(response.Results))
	return response, nil
}

// GetTrendingMovies retrieves trending movies from TMDB
func (s *MetadataService) GetTrendingMovies(timeWindow string, page int) (*tmdb.SearchResponse, error) {
	if timeWindow == "" {
		timeWindow = "week"
	}
	if page <= 0 {
		page = 1
	}

	s.logger.Debug("Getting trending movies", "timeWindow", timeWindow, "page", page)

	response, err := s.tmdb.GetTrending(timeWindow, page)
	if err != nil {
		s.logger.Error("Failed to get trending movies", "error", err)
		return nil, fmt.Errorf("failed to get trending movies: %w", err)
	}

	s.logger.Debug("Trending movies retrieved", "count", len(response.Results))
	return response, nil
}

// convertTMDBToMovie converts a TMDB movie to internal movie model
func (s *MetadataService) convertTMDBToMovie(tmdbMovie *tmdb.Movie, _ *tmdb.Credits) *models.Movie {
	releaseDate := s.parseReleaseDate(tmdbMovie.ReleaseDate)
	ratings := s.buildRatings(tmdbMovie)
	collection := s.buildCollection(tmdbMovie.BelongsToCollection)

	year := s.extractYear(releaseDate)
	cleanTitle := s.buildCleanTitle(tmdbMovie.Title)
	genresArray := s.buildGenresArray(tmdbMovie.Genres)

	movie := &models.Movie{
		TmdbID:              tmdbMovie.ID,
		ImdbID:              tmdbMovie.IMDbID,
		Title:               tmdbMovie.Title,
		OriginalTitle:       tmdbMovie.OriginalTitle,
		OriginalLanguage:    models.Language{Name: tmdbMovie.OriginalLanguage},
		Overview:            tmdbMovie.Overview,
		Website:             tmdbMovie.Homepage,
		Year:                year,
		Runtime:             tmdbMovie.Runtime,
		CleanTitle:          cleanTitle,
		Status:              models.MovieStatus(tmdbMovie.Status),
		Genres:              genresArray,
		Ratings:             ratings,
		Collection:          collection,
		Popularity:          tmdbMovie.Popularity,
		Studio:              s.extractStudio(tmdbMovie.ProductionCompanies),
		Certification:       "",
		Monitored:           true,
		MinimumAvailability: models.AvailabilityTBA,
		IsAvailable:         false,
	}

	s.setReleaseDates(movie, releaseDate)
	s.setMovieImages(movie, tmdbMovie)
	movie.TitleSlug = s.generateTitleSlug(movie.Title, movie.Year)

	return movie
}

// parseReleaseDate parses TMDB release date string to time.Time
func (s *MetadataService) parseReleaseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
		return &parsed
	}
	return nil
}

// buildRatings creates a ratings object from TMDB movie data
func (s *MetadataService) buildRatings(tmdbMovie *tmdb.Movie) models.Ratings {
	return models.Ratings{
		Tmdb: models.Rating{
			Value: tmdbMovie.VoteAverage,
			Votes: tmdbMovie.VoteCount,
		},
	}
}

// buildCollection creates a collection object from TMDB collection data
func (s *MetadataService) buildCollection(tmdbCollection *tmdb.Collection) *models.Collection {
	if tmdbCollection == nil {
		return nil
	}
	return &models.Collection{
		TmdbID: tmdbCollection.ID,
		Name:   tmdbCollection.Name,
	}
}

// extractYear extracts year from release date
func (s *MetadataService) extractYear(releaseDate *time.Time) int {
	if releaseDate == nil {
		return 0
	}
	return releaseDate.Year()
}

// buildCleanTitle creates a clean title for searching
func (s *MetadataService) buildCleanTitle(title string) string {
	cleanTitle := strings.ToLower(strings.TrimSpace(title))
	return strings.ReplaceAll(cleanTitle, " ", "")
}

// buildGenresArray converts TMDB genres to StringArray
func (s *MetadataService) buildGenresArray(genres []tmdb.Genre) models.StringArray {
	var genresArray models.StringArray
	for _, genre := range genres {
		genresArray = append(genresArray, genre.Name)
	}
	return genresArray
}

// setReleaseDates sets release dates on the movie
func (s *MetadataService) setReleaseDates(movie *models.Movie, releaseDate *time.Time) {
	if releaseDate != nil {
		movie.InCinemas = releaseDate
		movie.PhysicalRelease = releaseDate
		movie.DigitalRelease = releaseDate
	}
}

// setMovieImages sets movie images from TMDB data
func (s *MetadataService) setMovieImages(movie *models.Movie, tmdbMovie *tmdb.Movie) {
	var images models.MediaCover
	if tmdbMovie.PosterPath != "" {
		images = append(images, models.MediaCoverImage{
			CoverType: "poster",
			URL:       "https://image.tmdb.org/t/p/original" + tmdbMovie.PosterPath,
			RemoteURL: "https://image.tmdb.org/t/p/original" + tmdbMovie.PosterPath,
		})
	}
	if tmdbMovie.BackdropPath != "" {
		images = append(images, models.MediaCoverImage{
			CoverType: "fanart",
			URL:       "https://image.tmdb.org/t/p/original" + tmdbMovie.BackdropPath,
			RemoteURL: "https://image.tmdb.org/t/p/original" + tmdbMovie.BackdropPath,
		})
	}
	movie.Images = images
}

// extractStudio extracts the primary studio from production companies
func (s *MetadataService) extractStudio(companies []tmdb.ProductionCompany) string {
	if len(companies) == 0 {
		return ""
	}
	// Return the first production company as the primary studio
	return companies[0].Name
}

// generateTitleSlug generates a URL-friendly slug from title and year
func (s *MetadataService) generateTitleSlug(title string, year int) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters except hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")

	// Add year if available
	if year > 0 {
		if slug == "" {
			slug = strconv.Itoa(year)
		} else {
			slug += "-" + strconv.Itoa(year)
		}
	}

	return slug
}
