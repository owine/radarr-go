// Package models defines data structures for parsing and renaming functionality.
package models

import (
	"database/sql/driver"
	"encoding/json"
)

// ParsedMovieInfo represents information extracted from a release name
type ParsedMovieInfo struct {
	MovieTitles        []string `json:"movieTitles"`
	PrimaryMovieTitle  string   `json:"primaryMovieTitle"`
	OriginalTitle      string   `json:"originalTitle"`
	ReleaseTitle       string   `json:"releaseTitle"`
	SimpleReleaseTitle string   `json:"simpleReleaseTitle"`
	Quality            Quality  `json:"quality"`
	Languages          []string `json:"languages"`
	ReleaseGroup       string   `json:"releaseGroup"`
	ReleaseHash        string   `json:"releaseHash"`
	Edition            string   `json:"edition"`
	Year               int      `json:"year"`
	ImdbID             string   `json:"imdbId"`
	TmdbID             int      `json:"tmdbId"`
	HardcodedSubs      string   `json:"hardcodedSubs"`
}

// ParseResult represents the result of parsing a release name
type ParseResult struct {
	Title             string           `json:"title"`
	ParsedMovieInfo   *ParsedMovieInfo `json:"parsedMovieInfo"`
	Movie             *Movie           `json:"movie,omitempty"`
	CustomFormats     []CustomFormat   `json:"customFormats,omitempty"`
	CustomFormatScore int              `json:"customFormatScore"`
	MappingResult     string           `json:"mappingResult"`
}

// RenamePreview represents a file rename preview
type RenamePreview struct {
	MovieID      int    `json:"movieId"`
	MovieFileID  int    `json:"movieFileId"`
	ExistingPath string `json:"existingPath"`
	NewPath      string `json:"newPath"`
}

// RenameMovieRequest represents a request to rename movie files
type RenameMovieRequest struct {
	MovieIDs []int `json:"movieIds"`
}

// CustomFormatArray is a custom type for handling arrays of custom formats
type CustomFormatArray []CustomFormat

// Value implements the driver.Valuer interface for CustomFormatArray
func (cfa CustomFormatArray) Value() (driver.Value, error) {
	return json.Marshal(cfa)
}

// Scan implements the sql.Scanner interface for CustomFormatArray
func (cfa *CustomFormatArray) Scan(value interface{}) error {
	if value == nil {
		*cfa = CustomFormatArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, cfa)
}

// ParseRequest represents a request to parse release names
type ParseRequest struct {
	Titles []string `json:"titles"`
}

// ParseResponse represents the response from parsing release names
type ParseResponse struct {
	Results []ParseResult `json:"results"`
}
