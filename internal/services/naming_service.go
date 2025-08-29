package services

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// NamingService provides file and folder naming operations based on templates
type NamingService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewNamingService creates a new instance of NamingService
func NewNamingService(db *database.Database, logger *logger.Logger) *NamingService {
	return &NamingService{
		db:     db,
		logger: logger,
	}
}

// BuildMovieFilePath generates the complete file path for a movie using naming configuration
func (s *NamingService) BuildMovieFilePath(
	movie *models.Movie,
	quality *models.Quality,
	mediaInfo *models.MediaInfo,
	namingConfig *models.NamingConfig,
) (string, error) {
	if movie == nil {
		return "", fmt.Errorf("movie cannot be nil")
	}

	if namingConfig == nil {
		return "", fmt.Errorf("naming config cannot be nil")
	}

	// Get the root folder (this would typically come from movie.RootFolderPath)
	rootFolderPath := "/movies" // Default fallback

	// Build folder name using movie folder format
	folderName, err := s.BuildFolderName(movie, namingConfig)
	if err != nil {
		return "", fmt.Errorf("failed to build folder name: %w", err)
	}

	// Build file name using standard movie format
	fileName, err := s.BuildFileName(movie, quality, mediaInfo, namingConfig)
	if err != nil {
		return "", fmt.Errorf("failed to build file name: %w", err)
	}

	// Combine parts into full path
	fullPath := filepath.Join(rootFolderPath, folderName, fileName)

	// Apply character replacement rules
	fullPath = s.applyCharacterReplacement(fullPath, namingConfig)

	s.logger.Debug("Built movie file path",
		"movie", movie.Title,
		"folder", folderName,
		"filename", fileName,
		"fullPath", fullPath)

	return fullPath, nil
}

// BuildFolderName generates a folder name for a movie based on the folder format template
func (s *NamingService) BuildFolderName(movie *models.Movie, namingConfig *models.NamingConfig) (string, error) {
	folderFormat := namingConfig.MovieFolderFormat
	if folderFormat == "" {
		folderFormat = "{Movie Title} ({Release Year})"
	}

	// Create token replacement map
	tokens := s.buildTokenMap(movie, nil, nil)

	// Replace tokens in the format string
	folderName := s.replaceTokens(folderFormat, tokens)

	// Apply character replacement rules
	folderName = s.applyCharacterReplacement(folderName, namingConfig)

	return folderName, nil
}

// BuildFileName generates a file name for a movie based on the standard movie format template
func (s *NamingService) BuildFileName(
	movie *models.Movie,
	quality *models.Quality,
	mediaInfo *models.MediaInfo,
	namingConfig *models.NamingConfig,
) (string, error) {
	fileFormat := namingConfig.StandardMovieFormat
	if fileFormat == "" {
		fileFormat = "{Movie Title} ({Release Year}) {Quality Full}"
	}

	// Create token replacement map
	tokens := s.buildTokenMap(movie, quality, mediaInfo)

	// Replace tokens in the format string
	fileName := s.replaceTokens(fileFormat, tokens)

	// Apply character replacement rules
	fileName = s.applyCharacterReplacement(fileName, namingConfig)

	// Add file extension (assume .mkv for now, would be determined from source file)
	if !strings.Contains(fileName, ".") {
		fileName += ".mkv"
	}

	return fileName, nil
}

// buildTokenMap creates a map of tokens and their replacement values
func (s *NamingService) buildTokenMap(movie *models.Movie, quality *models.Quality, mediaInfo *models.MediaInfo) map[string]string {
	tokens := make(map[string]string)

	s.addMovieTokens(tokens, movie)
	s.addQualityTokens(tokens, quality)
	s.addMediaInfoTokens(tokens, mediaInfo)
	s.setOptionalTokenDefaults(tokens)

	return tokens
}

// addMovieTokens adds movie-related tokens to the token map
func (s *NamingService) addMovieTokens(tokens map[string]string, movie *models.Movie) {
	if movie == nil {
		return
	}

	// Movie tokens
	tokens["{Movie Title}"] = movie.Title
	tokens["{Movie CleanTitle}"] = s.cleanTitle(movie.Title)
	tokens["{Movie TitleThe}"] = s.moveArticleToEnd(movie.Title)
	tokens["{Movie OriginalTitle}"] = movie.OriginalTitle
	tokens["{Movie TitleFirstCharacter}"] = s.getFirstCharacter(movie.Title)

	// Release year tokens
	if movie.Year > 0 {
		tokens["{Release Year}"] = strconv.Itoa(movie.Year)
		tokens["{Release YearFirst}"] = strconv.Itoa(movie.Year)
	} else {
		tokens["{Release Year}"] = "Unknown"
		tokens["{Release YearFirst}"] = "Unknown"
	}

	// IMDB/TMDB tokens
	tokens["{ImdbId}"] = movie.ImdbID
	tokens["{Tmdb Id}"] = strconv.Itoa(movie.TmdbID)
}

// addQualityTokens adds quality-related tokens to the token map
func (s *NamingService) addQualityTokens(tokens map[string]string, quality *models.Quality) {
	if quality == nil {
		return
	}

	// Quality tokens
	tokens["{Quality Full}"] = s.buildQualityString(quality)
	tokens["{Quality Title}"] = quality.Quality.Name

	// Proper/Repack tokens
	switch {
	case quality.Revision.IsRepack:
		tokens["{Quality Proper}"] = "Repack"
		tokens["{Quality Real}"] = "REPACK"
	case quality.Revision.Real > 0:
		tokens["{Quality Proper}"] = "Proper"
		tokens["{Quality Real}"] = "REAL"
	default:
		tokens["{Quality Proper}"] = ""
		tokens["{Quality Real}"] = ""
	}
}

// addMediaInfoTokens adds media info-related tokens to the token map
func (s *NamingService) addMediaInfoTokens(tokens map[string]string, mediaInfo *models.MediaInfo) {
	if mediaInfo == nil {
		return
	}

	// Media info tokens
	tokens["{MediaInfo Simple}"] = s.buildSimpleMediaInfo(mediaInfo)
	tokens["{MediaInfo Full}"] = s.buildFullMediaInfo(mediaInfo)
	tokens["{MediaInfo VideoCodec}"] = mediaInfo.VideoCodec
	tokens["{MediaInfo VideoBitDepth}"] = s.formatBitDepth(mediaInfo.VideoBitDepth)
	tokens["{MediaInfo VideoResolution}"] = mediaInfo.Resolution
	tokens["{MediaInfo AudioCodec}"] = mediaInfo.AudioCodec
	tokens["{MediaInfo AudioChannels}"] = s.formatAudioChannels(mediaInfo.AudioChannels)
	tokens["{MediaInfo AudioLanguages}"] = mediaInfo.AudioLanguages
	tokens["{MediaInfo SubtitleLanguages}"] = mediaInfo.Subtitles
}

// setOptionalTokenDefaults sets empty values for missing optional tokens
func (s *NamingService) setOptionalTokenDefaults(tokens map[string]string) {
	optionalTokens := []string{
		"{Movie OriginalTitle}", "{Movie Collection}", "{Edition Tags}",
		"{Custom Formats}", "{Release Group}", "{ImdbId}",
		"{Quality Proper}", "{Quality Real}",
		"{MediaInfo Simple}", "{MediaInfo Full}", "{MediaInfo VideoCodec}",
		"{MediaInfo VideoBitDepth}", "{MediaInfo VideoResolution}",
		"{MediaInfo AudioCodec}", "{MediaInfo AudioChannels}",
		"{MediaInfo AudioLanguages}", "{MediaInfo SubtitleLanguages}",
	}

	for _, token := range optionalTokens {
		if _, exists := tokens[token]; !exists {
			tokens[token] = ""
		}
	}
}

// replaceTokens replaces all tokens in the format string with their values
func (s *NamingService) replaceTokens(format string, tokens map[string]string) string {
	result := format

	for token, value := range tokens {
		result = strings.ReplaceAll(result, token, value)
	}

	// Clean up multiple spaces and trim
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	// Remove empty parentheses and brackets
	result = regexp.MustCompile(`\(\s*\)`).ReplaceAllString(result, "")
	result = regexp.MustCompile(`\[\s*\]`).ReplaceAllString(result, "")
	result = regexp.MustCompile(`\{\s*\}`).ReplaceAllString(result, "")

	// Clean up multiple spaces again after removing empty brackets
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	return result
}

// applyCharacterReplacement applies character replacement rules based on naming config
func (s *NamingService) applyCharacterReplacement(input string, config *models.NamingConfig) string {
	if !config.ReplaceIllegalChars {
		return input
	}

	// Apply colon replacement first
	result := config.ApplyColonReplacement(input)

	// Apply illegal character replacement
	result = config.ReplaceIllegalCharacters(result)

	return result
}

// cleanTitle removes special characters and formats the title for clean usage
func (s *NamingService) cleanTitle(title string) string {
	// Remove special characters commonly found in titles
	cleaned := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(title, "")

	// Normalize spaces
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}

// moveArticleToEnd moves articles (The, A, An) from the beginning to the end
func (s *NamingService) moveArticleToEnd(title string) string {
	articles := []string{"The ", "A ", "An "}

	for _, article := range articles {
		if strings.HasPrefix(title, article) {
			remaining := strings.TrimPrefix(title, article)
			return fmt.Sprintf("%s, %s", remaining, strings.TrimSpace(article))
		}
	}

	return title
}

// getFirstCharacter returns the first character of the title, handling articles
func (s *NamingService) getFirstCharacter(title string) string {
	if title == "" {
		return ""
	}

	// Skip common articles
	cleaned := title
	articles := []string{"The ", "A ", "An "}

	for _, article := range articles {
		if strings.HasPrefix(cleaned, article) {
			cleaned = strings.TrimPrefix(cleaned, article)
			break
		}
	}

	if cleaned == "" {
		return ""
	}

	return strings.ToUpper(string(cleaned[0]))
}

// buildQualityString creates a full quality string including proper/repack info
func (s *NamingService) buildQualityString(quality *models.Quality) string {
	qualityStr := quality.Quality.Name

	if quality.Revision.IsRepack {
		qualityStr += " Repack"
	} else if quality.Revision.Real > 0 {
		qualityStr += " Proper"
	}

	return qualityStr
}

// buildSimpleMediaInfo creates a simple media info string
func (s *NamingService) buildSimpleMediaInfo(mediaInfo *models.MediaInfo) string {
	parts := []string{}

	if mediaInfo.VideoCodec != "" {
		parts = append(parts, mediaInfo.VideoCodec)
	}

	if mediaInfo.AudioCodec != "" {
		parts = append(parts, mediaInfo.AudioCodec)
	}

	return strings.Join(parts, " ")
}

// buildFullMediaInfo creates a comprehensive media info string
func (s *NamingService) buildFullMediaInfo(mediaInfo *models.MediaInfo) string {
	parts := []string{}

	if mediaInfo.VideoCodec != "" {
		codec := mediaInfo.VideoCodec
		if mediaInfo.VideoBitDepth > 0 && mediaInfo.VideoBitDepth != 8 {
			codec += fmt.Sprintf(" %dbit", mediaInfo.VideoBitDepth)
		}
		parts = append(parts, codec)
	}

	if mediaInfo.AudioCodec != "" {
		codec := mediaInfo.AudioCodec
		if mediaInfo.AudioChannels > 0 {
			codec += fmt.Sprintf(" %.1f", mediaInfo.AudioChannels)
		}
		parts = append(parts, codec)
	}

	if mediaInfo.AudioLanguages != "" {
		parts = append(parts, mediaInfo.AudioLanguages)
	}

	return strings.Join(parts, " ")
}

// formatBitDepth formats video bit depth
func (s *NamingService) formatBitDepth(bitDepth int) string {
	if bitDepth <= 0 || bitDepth == 8 {
		return ""
	}
	return fmt.Sprintf("%dbit", bitDepth)
}

// formatAudioChannels formats audio channel count
func (s *NamingService) formatAudioChannels(channels float64) string {
	if channels <= 0 {
		return ""
	}

	// Convert common channel counts to standard notation
	switch channels {
	case 1:
		return "Mono"
	case 2:
		return "Stereo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%.1f", channels)
	}
}

// ValidateNamingFormat validates that a naming format is valid
func (s *NamingService) ValidateNamingFormat(format string) []string {
	return models.ValidateNamingFormat(format)
}

// GetAvailableTokens returns all available naming tokens
func (s *NamingService) GetAvailableTokens() []models.NamingToken {
	return models.GetAvailableTokens()
}

// PreviewNaming generates a preview of how files would be named
func (s *NamingService) PreviewNaming(
	movie *models.Movie,
	namingConfig *models.NamingConfig,
) (*NamingPreview, error) {
	// Create sample quality and media info for preview
	sampleQuality := &models.Quality{
		Quality: models.QualityDefinition{
			ID:         6,
			Name:       "Bluray-1080p",
			Source:     "bluray",
			Resolution: 1080,
		},
		Revision: models.Revision{Version: 1},
	}

	sampleMediaInfo := &models.MediaInfo{
		VideoCodec:     "x264",
		AudioCodec:     "DTS",
		AudioChannels:  5.1,
		Resolution:     "1920x1080",
		VideoBitDepth:  8,
		AudioLanguages: "[EN]",
		Subtitles:      "[EN+ES]",
	}

	folderName, err := s.BuildFolderName(movie, namingConfig)
	if err != nil {
		return nil, err
	}

	fileName, err := s.BuildFileName(movie, sampleQuality, sampleMediaInfo, namingConfig)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(folderName, fileName)

	return &NamingPreview{
		Movie:      movie,
		FolderName: folderName,
		FileName:   fileName,
		FullPath:   fullPath,
		Quality:    sampleQuality,
		MediaInfo:  sampleMediaInfo,
	}, nil
}

// NamingPreview represents a preview of naming output
type NamingPreview struct {
	Movie      *models.Movie     `json:"movie"`
	FolderName string            `json:"folderName"`
	FileName   string            `json:"fileName"`
	FullPath   string            `json:"fullPath"`
	Quality    *models.Quality   `json:"quality"`
	MediaInfo  *models.MediaInfo `json:"mediaInfo"`
}

// GetNamingConfig retrieves the current naming configuration
func (s *NamingService) GetNamingConfig() (*models.NamingConfig, error) {
	var config models.NamingConfig

	err := s.db.GORM.First(&config).Error
	if err != nil {
		// If no config exists, return default
		defaultConfig := models.GetDefaultNamingConfig()
		return defaultConfig, nil
	}

	return &config, nil
}

// UpdateNamingConfig updates the naming configuration
func (s *NamingService) UpdateNamingConfig(config *models.NamingConfig) error {
	// Validate configuration
	if errors := config.ValidateConfiguration(); len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	// Check if config exists
	var existingConfig models.NamingConfig
	err := s.db.GORM.First(&existingConfig).Error

	if err != nil {
		// Create new config
		if createErr := s.db.GORM.Create(config).Error; createErr != nil {
			return fmt.Errorf("failed to create naming config: %w", createErr)
		}
	} else {
		// Update existing config
		config.ID = existingConfig.ID
		if updateErr := s.db.GORM.Save(config).Error; updateErr != nil {
			return fmt.Errorf("failed to update naming config: %w", updateErr)
		}
	}

	s.logger.Info("Updated naming configuration", "id", config.ID)
	return nil
}
