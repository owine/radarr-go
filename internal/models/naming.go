package models

import (
	"fmt"
	"strings"
	"time"
)

// NamingConfig represents the file and folder naming configuration for Radarr
type NamingConfig struct {
	ID                     int              `json:"id" gorm:"primaryKey;autoIncrement"`
	RenameMovies           bool             `json:"renameMovies" gorm:"default:false"`
	ReplaceIllegalChars    bool             `json:"replaceIllegalCharacters" gorm:"default:true"`
	ColonReplacementFormat ColonReplacement `json:"colonReplacementFormat" gorm:"default:'delete'"`
	StandardMovieFormat    string           `json:"standardMovieFormat"`
	MovieFolderFormat      string           `json:"movieFolderFormat"`
	CreateEmptyFolders     bool             `json:"createEmptyMovieFolders" gorm:"default:false"`
	DeleteEmptyFolders     bool             `json:"deleteEmptyFolders" gorm:"default:false"`
	SkipFreeSpaceCheck     bool             `json:"skipFreeSpaceCheckWhenImporting" gorm:"default:false"`
	MinimumFreeSpace       int64            `json:"minimumFreeSpaceWhenImporting" gorm:"default:100"`
	UseHardlinks           bool             `json:"copyUsingHardlinks" gorm:"default:true"`
	ImportExtraFiles       bool             `json:"importExtraFiles" gorm:"default:false"`
	ExtraFileExtensions    StringArray      `json:"extraFileExtensions" gorm:"type:text"`
	EnableMediaInfo        bool             `json:"enableMediaInfo" gorm:"default:true"`
	CreatedAt              time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt              time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the NamingConfig model
func (NamingConfig) TableName() string {
	return "naming_config"
}

// ColonReplacement represents how colons should be handled in file names
type ColonReplacement string

const (
	// ColonReplacementDelete removes colons entirely
	ColonReplacementDelete ColonReplacement = "delete"
	// ColonReplacementDash replaces colons with dashes
	ColonReplacementDash ColonReplacement = "dash"
	// ColonReplacementSpaceDash replaces colons with space-dash
	ColonReplacementSpaceDash ColonReplacement = "spaceDash"
	// ColonReplacementSpaceDashSpace replaces colons with space-dash-space
	ColonReplacementSpaceDashSpace ColonReplacement = "spaceDashSpace"
)

// NamingToken represents a token that can be used in naming patterns
type NamingToken struct {
	Token       string `json:"token"`
	Example     string `json:"example"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

// GetDefaultNamingConfig returns the default naming configuration
func GetDefaultNamingConfig() *NamingConfig {
	return &NamingConfig{
		RenameMovies:           false,
		ReplaceIllegalChars:    true,
		ColonReplacementFormat: ColonReplacementDelete,
		StandardMovieFormat:    "{Movie Title} ({Release Year}) {Quality Full}",
		MovieFolderFormat:      "{Movie Title} ({Release Year})",
		CreateEmptyFolders:     false,
		DeleteEmptyFolders:     false,
		SkipFreeSpaceCheck:     false,
		MinimumFreeSpace:       100,
		UseHardlinks:           true,
		ImportExtraFiles:       false,
		ExtraFileExtensions:    StringArray{"srt", "nfo"},
		EnableMediaInfo:        true,
	}
}

// GetAvailableTokens returns all available naming tokens
func GetAvailableTokens() []NamingToken {
	return []NamingToken{
		// Movie Tokens
		{Token: "{Movie Title}", Example: "The Dark Knight", Description: "Movie Title", Optional: false},
		{Token: "{Movie CleanTitle}", Example: "The Dark Knight",
			Description: "Movie title without special characters", Optional: false},
		{Token: "{Movie TitleThe}", Example: "Dark Knight, The",
			Description: "Movie title with 'The' moved to the end", Optional: false},
		{Token: "{Movie OriginalTitle}", Example: "The Dark Knight", Description: "Original movie title", Optional: true},
		{Token: "{Movie TitleFirstCharacter}", Example: "D", Description: "First character of movie title", Optional: false},
		{Token: "{Movie Collection}", Example: "The Dark Knight Collection",
			Description: "Movie collection name", Optional: true},

		// Year Tokens
		{Token: "{Release Year}", Example: "2008", Description: "Year the movie was released", Optional: false},
		{Token: "{Release YearFirst}", Example: "2008", Description: "Year from first release date", Optional: false},

		// Quality Tokens
		{Token: "{Quality Full}", Example: "HDTV-720p Proper",
			Description: "Full quality name including proper/repack", Optional: false},
		{Token: "{Quality Title}", Example: "HDTV-720p", Description: "Quality name", Optional: false},
		{Token: "{Quality Proper}", Example: "Proper", Description: "Quality Proper", Optional: true},
		{Token: "{Quality Real}", Example: "REAL", Description: "Quality Real", Optional: true},

		// Media Info Tokens
		{Token: "{MediaInfo Simple}", Example: "x264 DTS", Description: "Simple media info", Optional: true},
		{Token: "{MediaInfo Full}", Example: "x264 DTS [EN+DE+ES]", Description: "Full media info", Optional: true},
		{Token: "{MediaInfo VideoCodec}", Example: "x264", Description: "Video codec", Optional: true},
		{Token: "{MediaInfo VideoBitDepth}", Example: "10bit", Description: "Video bit depth", Optional: true},
		{Token: "{MediaInfo VideoResolution}", Example: "1080p", Description: "Video resolution", Optional: true},
		{Token: "{MediaInfo AudioCodec}", Example: "DTS", Description: "Audio codec", Optional: true},
		{Token: "{MediaInfo AudioChannels}", Example: "5.1", Description: "Audio channels", Optional: true},
		{Token: "{MediaInfo AudioLanguages}", Example: "[EN+DE+ES]", Description: "Audio languages", Optional: true},
		{Token: "{MediaInfo SubtitleLanguages}", Example: "[EN+DE+ES]", Description: "Subtitle languages", Optional: true},

		// Source Tokens
		{Token: "{Edition Tags}", Example: "Director's Cut", Description: "Edition information", Optional: true},
		{Token: "{Custom Formats}", Example: "iNTERNAL", Description: "Custom format tags", Optional: true},

		// Release Group Tokens
		{Token: "{Release Group}", Example: "EVOLVE", Description: "Release group name", Optional: true},

		// IMDB Tokens
		{Token: "{ImdbId}", Example: "tt0468569", Description: "IMDB ID", Optional: true},
		{Token: "{Tmdb Id}", Example: "155", Description: "TMDB ID", Optional: false},
	}
}

// ValidateNamingFormat validates a naming format string
func ValidateNamingFormat(format string) []string {
	var errors []string

	if format == "" {
		errors = append(errors, "Naming format cannot be empty")
		return errors
	}

	// Check for required tokens
	requiredTokens := []string{"{Movie Title}", "{Movie CleanTitle}"}
	hasRequiredToken := false

	for _, token := range requiredTokens {
		if strings.Contains(format, token) {
			hasRequiredToken = true
			break
		}
	}

	if !hasRequiredToken {
		errors = append(errors, "Naming format must contain at least one movie title token")
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(format, char) {
			errors = append(errors, fmt.Sprintf("Naming format contains invalid character: %s", char))
		}
	}

	return errors
}

// ApplyColonReplacement applies colon replacement rules to a string
func (nc *NamingConfig) ApplyColonReplacement(input string) string {
	if !strings.Contains(input, ":") {
		return input
	}

	switch nc.ColonReplacementFormat {
	case ColonReplacementDelete:
		return strings.ReplaceAll(input, ":", "")
	case ColonReplacementDash:
		return strings.ReplaceAll(input, ":", "-")
	case ColonReplacementSpaceDash:
		return strings.ReplaceAll(input, ":", " -")
	case ColonReplacementSpaceDashSpace:
		return strings.ReplaceAll(input, ":", " - ")
	default:
		return strings.ReplaceAll(input, ":", "")
	}
}

// ReplaceIllegalCharacters removes or replaces illegal file system characters
func (nc *NamingConfig) ReplaceIllegalCharacters(input string) string {
	if !nc.ReplaceIllegalChars {
		return input
	}

	// Windows illegal characters
	illegalChars := map[string]string{
		"<":  "",
		">":  "",
		":":  "",
		"\"": "",
		"|":  "",
		"?":  "",
		"*":  "",
		"/":  "",
		"\\": "",
	}

	result := input
	for illegal, replacement := range illegalChars {
		result = strings.ReplaceAll(result, illegal, replacement)
	}

	// Apply colon replacement after other illegal characters
	result = nc.ApplyColonReplacement(result)

	return result
}

// ValidateConfiguration validates the naming configuration
func (nc *NamingConfig) ValidateConfiguration() []string {
	var errors []string

	// Validate standard movie format
	if formatErrors := ValidateNamingFormat(nc.StandardMovieFormat); len(formatErrors) > 0 {
		for _, err := range formatErrors {
			errors = append(errors, fmt.Sprintf("Standard Movie Format: %s", err))
		}
	}

	// Validate movie folder format
	if formatErrors := ValidateNamingFormat(nc.MovieFolderFormat); len(formatErrors) > 0 {
		for _, err := range formatErrors {
			errors = append(errors, fmt.Sprintf("Movie Folder Format: %s", err))
		}
	}

	// Validate minimum free space
	if nc.MinimumFreeSpace < 0 {
		errors = append(errors, "Minimum free space cannot be negative")
	}

	// Validate extra file extensions
	for _, ext := range nc.ExtraFileExtensions {
		if ext == "" {
			errors = append(errors, "Extra file extensions cannot be empty")
		}
		if strings.Contains(ext, ".") && !strings.HasPrefix(ext, ".") {
			errors = append(errors, fmt.Sprintf("Extra file extension '%s' should not contain dots unless it's a prefix", ext))
		}
	}

	return errors
}
