// Package services provides business logic for the Radarr application.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// ParseService handles release name parsing and caching
type ParseService struct {
	db     *database.Database
	logger *logger.Logger

	// Regular expressions for parsing release names
	titleYearRegex    *regexp.Regexp
	qualityRegex      *regexp.Regexp
	releaseGroupRegex *regexp.Regexp
	languageRegex     *regexp.Regexp
	editionRegex      *regexp.Regexp
	sourceRegex       *regexp.Regexp
	codecRegex        *regexp.Regexp
	resolutionRegex   *regexp.Regexp
}

// ParseCacheEntry represents a cached parse result
type ParseCacheEntry struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	ReleaseTitle string    `json:"releaseTitle" gorm:"uniqueIndex;size:1000;not null"`
	ParsedInfo   string    `json:"parsedInfo" gorm:"type:text;not null"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for ParseCacheEntry
func (ParseCacheEntry) TableName() string {
	return "parse_cache"
}

// NewParseService creates a new parse service
func NewParseService(db *database.Database, logger *logger.Logger) *ParseService {
	service := &ParseService{
		db:     db,
		logger: logger,
	}

	// Initialize regular expressions for parsing
	service.initializeRegexes()

	return service
}

// ParseReleaseTitle parses a release title and returns movie information
func (s *ParseService) ParseReleaseTitle(ctx context.Context, releaseTitle string) (*models.ParseResult, error) {
	// Check cache first
	if cached, err := s.getCachedResult(ctx, releaseTitle); err == nil {
		return cached, nil
	}

	// Parse the release title
	parsedInfo := s.parseTitle(releaseTitle)

	result := &models.ParseResult{
		Title:           releaseTitle,
		ParsedMovieInfo: parsedInfo,
		MappingResult:   "parsed",
	}

	// Try to find matching movie
	if movie, err := s.findMatchingMovie(ctx, parsedInfo); err == nil {
		result.Movie = movie
		result.MappingResult = "found"
	}

	// Cache the result
	if err := s.cacheResult(ctx, releaseTitle, result); err != nil {
		s.logger.Warn("Failed to cache parse result", "title", releaseTitle, "error", err)
	}

	s.logger.Debug("Parsed release title", "title", releaseTitle, "movieTitle", parsedInfo.PrimaryMovieTitle, "year", parsedInfo.Year)
	return result, nil
}

// ParseMultipleTitles parses multiple release titles
func (s *ParseService) ParseMultipleTitles(ctx context.Context, titles []string) ([]*models.ParseResult, error) {
	results := make([]*models.ParseResult, 0, len(titles))

	for _, title := range titles {
		if result, err := s.ParseReleaseTitle(ctx, title); err == nil {
			results = append(results, result)
		} else {
			// Still add a result with error information
			results = append(results, &models.ParseResult{
				Title:         title,
				MappingResult: "failed",
			})
		}
	}

	s.logger.Info("Parsed multiple titles", "total", len(titles), "successful", len(results))
	return results, nil
}

// ClearCache clears the parse cache
func (s *ParseService) ClearCache(ctx context.Context) error {
	if err := s.db.GORM.WithContext(ctx).Delete(&ParseCacheEntry{}, "1=1").Error; err != nil {
		s.logger.Error("Failed to clear parse cache", "error", err)
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	s.logger.Info("Cleared parse cache")
	return nil
}

// CleanupOldCacheEntries removes cache entries older than specified duration
func (s *ParseService) CleanupOldCacheEntries(ctx context.Context, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	result := s.db.GORM.WithContext(ctx).
		Where("created_at < ?", cutoff).
		Delete(&ParseCacheEntry{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup old cache entries", "error", result.Error)
		return fmt.Errorf("failed to cleanup cache: %w", result.Error)
	}

	s.logger.Info("Cleaned up old cache entries", "removed", result.RowsAffected, "cutoff", cutoff)
	return nil
}

// initializeRegexes initializes all the regular expressions used for parsing
func (s *ParseService) initializeRegexes() {
	// Title and year pattern (most important)
	s.titleYearRegex = regexp.MustCompile(`(?i)^(.+?)[\.\s]+(?:[\(\[]?)(\d{4})(?:[\)\]]?)?.*`)

	// Quality patterns
	s.qualityRegex = regexp.MustCompile(`(?i)\b(?:720p|1080p|2160p|4k|uhd|hd|sd|480p|576p|hdtv|webrip|web-dl|bluray|bdrip|dvdrip|ts|cam|hdcam)\b`)

	// Release group (usually at the end in brackets or after a dash)
	s.releaseGroupRegex = regexp.MustCompile(`(?i)[\[\(]?([a-z0-9_\-\.]+)[\]\)]?$`)

	// Language patterns
	s.languageRegex = regexp.MustCompile(`(?i)\b(?:english|spanish|french|german|italian|portuguese|russian|chinese|japanese|korean|hindi|arabic|dutch|swedish|norwegian|danish|finnish)\b`)

	// Edition patterns
	s.editionRegex = regexp.MustCompile(`(?i)\b(?:director'?s?\.?cut|extended|unrated|theatrical|remastered|criterion|special\.edition)\b`)

	// Source patterns
	s.sourceRegex = regexp.MustCompile(`(?i)\b(?:bluray|bdrip|web-dl|webrip|hdtv|dvdrip|ts|cam|hdcam|r5|dvdscr|workprint|ppv)\b`)

	// Codec patterns
	s.codecRegex = regexp.MustCompile(`(?i)\b(?:x264|x265|h264|h265|hevc|xvid|divx|vc-1|mpeg2)\b`)

	// Resolution patterns
	s.resolutionRegex = regexp.MustCompile(`(?i)\b(?:480p|576p|720p|1080p|2160p|4k)\b`)
}

// parseTitle parses a release title and extracts movie information
func (s *ParseService) parseTitle(title string) *models.ParsedMovieInfo {
	cleanTitle := s.cleanTitle(title)

	parsed := &models.ParsedMovieInfo{
		ReleaseTitle:       title,
		SimpleReleaseTitle: cleanTitle,
		MovieTitles:        []string{},
		Languages:          []string{"english"}, // default
	}

	// Extract title and year
	if matches := s.titleYearRegex.FindStringSubmatch(cleanTitle); len(matches) >= 3 {
		movieTitle := strings.TrimSpace(matches[1])
		movieTitle = strings.ReplaceAll(movieTitle, ".", " ")
		movieTitle = strings.ReplaceAll(movieTitle, "_", " ")
		movieTitle = s.cleanMovieTitle(movieTitle)

		parsed.MovieTitles = append(parsed.MovieTitles, movieTitle)
		parsed.PrimaryMovieTitle = movieTitle

		if year, err := strconv.Atoi(matches[2]); err == nil {
			parsed.Year = year
		}
	} else {
		// Fallback: try to extract just the movie title without year
		parts := strings.Split(cleanTitle, ".")
		if len(parts) > 0 {
			movieTitle := s.cleanMovieTitle(parts[0])
			parsed.MovieTitles = append(parsed.MovieTitles, movieTitle)
			parsed.PrimaryMovieTitle = movieTitle
		}
	}

	// Extract quality information
	parsed.Quality = s.extractQuality(title)

	// Extract release group
	if matches := s.releaseGroupRegex.FindStringSubmatch(title); len(matches) >= 2 {
		parsed.ReleaseGroup = matches[1]
	}

	// Extract language
	if matches := s.languageRegex.FindStringSubmatch(title); len(matches) >= 1 {
		parsed.Languages = []string{strings.ToLower(matches[0])}
	}

	// Extract edition
	if matches := s.editionRegex.FindStringSubmatch(title); len(matches) >= 1 {
		parsed.Edition = matches[0]
	}

	// Generate original title from parsed info
	if parsed.PrimaryMovieTitle != "" {
		parsed.OriginalTitle = parsed.PrimaryMovieTitle
		if parsed.Year > 0 {
			parsed.OriginalTitle += fmt.Sprintf(" (%d)", parsed.Year)
		}
	}

	return parsed
}

// extractQuality extracts quality information from the release title
func (s *ParseService) extractQuality(title string) models.Quality {
	quality := models.Quality{
		Quality:  models.QualityDefinition{Name: "Unknown"},
		Revision: models.Revision{Version: 1},
	}

	titleLower := strings.ToLower(title)

	// Determine quality based on resolution and source
	switch {
	case strings.Contains(titleLower, "2160p") || strings.Contains(titleLower, "4k"):
		quality.Quality = models.QualityDefinition{ID: 10, Name: "2160p", Resolution: 2160}
	case strings.Contains(titleLower, "1080p"):
		switch {
		case strings.Contains(titleLower, "bluray"):
			quality.Quality = models.QualityDefinition{ID: 7, Name: "1080p Bluray", Resolution: 1080, Source: "bluray"}
		case strings.Contains(titleLower, "web-dl") || strings.Contains(titleLower, "webrip"):
			quality.Quality = models.QualityDefinition{ID: 6, Name: "1080p WEB-DL", Resolution: 1080, Source: "webdl"}
		default:
			quality.Quality = models.QualityDefinition{ID: 5, Name: "1080p HDTV", Resolution: 1080, Source: "hdtv"}
		}
	case strings.Contains(titleLower, "720p"):
		switch {
		case strings.Contains(titleLower, "bluray"):
			quality.Quality = models.QualityDefinition{ID: 4, Name: "720p Bluray", Resolution: 720, Source: "bluray"}
		case strings.Contains(titleLower, "web-dl") || strings.Contains(titleLower, "webrip"):
			quality.Quality = models.QualityDefinition{ID: 3, Name: "720p WEB-DL", Resolution: 720, Source: "webdl"}
		default:
			quality.Quality = models.QualityDefinition{ID: 2, Name: "720p HDTV", Resolution: 720, Source: "hdtv"}
		}
	case strings.Contains(titleLower, "480p") || strings.Contains(titleLower, "dvdrip"):
		quality.Quality = models.QualityDefinition{ID: 1, Name: "DVD-Rip", Resolution: 480, Source: "dvd"}
	default:
		quality.Quality = models.QualityDefinition{ID: 0, Name: "Unknown", Resolution: 0}
	}

	return quality
}

// cleanTitle removes common release name artifacts
func (s *ParseService) cleanTitle(title string) string {
	// Remove file extension
	if dotIndex := strings.LastIndex(title, "."); dotIndex > 0 {
		if ext := title[dotIndex+1:]; len(ext) <= 4 && strings.ToLower(ext) != title[dotIndex+1:] {
			title = title[:dotIndex]
		}
	}

	// Remove common prefixes/suffixes
	prefixesToRemove := []string{"www.", "rarbg.", "yts."}
	for _, prefix := range prefixesToRemove {
		title = strings.TrimPrefix(title, prefix)
	}

	return title
}

// cleanMovieTitle cleans the extracted movie title
func (s *ParseService) cleanMovieTitle(title string) string {
	// Remove extra spaces
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)

	// Remove common artifacts
	title = regexp.MustCompile(`(?i)\b(?:dvdrip|bdrip|webrip|hdtv|web-dl|bluray|1080p|720p|480p)\b`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)

	return title
}

// findMatchingMovie finds a movie that matches the parsed information
func (s *ParseService) findMatchingMovie(ctx context.Context, parsed *models.ParsedMovieInfo) (*models.Movie, error) {
	if parsed.PrimaryMovieTitle == "" {
		return nil, fmt.Errorf("no movie title to search for")
	}

	var movie models.Movie
	query := s.db.GORM.WithContext(ctx)

	// First try: exact title and year match
	if parsed.Year > 0 {
		if err := query.Where("title ILIKE ? AND year = ?", parsed.PrimaryMovieTitle, parsed.Year).First(&movie).Error; err == nil {
			return &movie, nil
		}
	}

	// Second try: fuzzy title match with year
	if parsed.Year > 0 {
		if err := query.Where("title ILIKE ? AND year = ?", fmt.Sprintf("%%%s%%", parsed.PrimaryMovieTitle), parsed.Year).First(&movie).Error; err == nil {
			return &movie, nil
		}
	}

	// Third try: exact title match without year
	if err := query.Where("title ILIKE ?", parsed.PrimaryMovieTitle).First(&movie).Error; err == nil {
		return &movie, nil
	}

	// Fourth try: fuzzy title match without year
	if err := query.Where("title ILIKE ?", fmt.Sprintf("%%%s%%", parsed.PrimaryMovieTitle)).First(&movie).Error; err == nil {
		return &movie, nil
	}

	return nil, fmt.Errorf("no matching movie found")
}

// getCachedResult retrieves a cached parse result
func (s *ParseService) getCachedResult(ctx context.Context, releaseTitle string) (*models.ParseResult, error) {
	var entry ParseCacheEntry
	if err := s.db.GORM.WithContext(ctx).
		Where("release_title = ?", releaseTitle).
		First(&entry).Error; err != nil {
		return nil, err
	}

	var result models.ParseResult
	if err := json.Unmarshal([]byte(entry.ParsedInfo), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached result: %w", err)
	}

	return &result, nil
}

// cacheResult caches a parse result
func (s *ParseService) cacheResult(ctx context.Context, releaseTitle string, result *models.ParseResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	entry := &ParseCacheEntry{
		ReleaseTitle: releaseTitle,
		ParsedInfo:   string(data),
	}

	// Use ON CONFLICT for upsert
	if err := s.db.GORM.WithContext(ctx).
		Create(entry).Error; err != nil {
		// If creation failed due to duplicate, try update
		if err := s.db.GORM.WithContext(ctx).
			Model(&ParseCacheEntry{}).
			Where("release_title = ?", releaseTitle).
			Update("parsed_info", string(data)).Error; err != nil {
			return fmt.Errorf("failed to cache result: %w", err)
		}
	}

	return nil
}
