package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// ImportService provides file import and processing functionality
type ImportService struct {
	db                      *database.Database
	logger                  *logger.Logger
	movieService            *MovieService
	movieFileService        *MovieFileService
	fileOrganizationService *FileOrganizationService
	mediaInfoService        *MediaInfoService
	namingService           *NamingService
}

// NewImportService creates a new instance of ImportService
func NewImportService(
	db *database.Database,
	logger *logger.Logger,
	movieService *MovieService,
	movieFileService *MovieFileService,
	fileOrganizationService *FileOrganizationService,
	mediaInfoService *MediaInfoService,
	namingService *NamingService,
) *ImportService {
	return &ImportService{
		db:                      db,
		logger:                  logger,
		movieService:            movieService,
		movieFileService:        movieFileService,
		fileOrganizationService: fileOrganizationService,
		mediaInfoService:        mediaInfoService,
		namingService:           namingService,
	}
}

// ProcessImport processes imported files and makes decisions about them
func (s *ImportService) ProcessImport(path string, options *ImportOptions) (*models.FileImportResult, error) {
	start := time.Now()
	s.logger.Info("Starting import process", "path", path)

	if options == nil {
		options = &ImportOptions{
			ImportMode: models.ImportDecisionApproved,
		}
	}

	// Scan directory for importable files
	importableFiles, err := s.fileOrganizationService.ScanDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(importableFiles) == 0 {
		s.logger.Info("No importable files found", "path", path)
		return &models.FileImportResult{
			ImportDecisions: []models.ImportDecision{},
			ProcessingTime:  time.Since(start),
		}, nil
	}

	s.logger.Info("Found importable files", "count", len(importableFiles), "path", path)

	// Make import decisions for each file
	importDecisions, err := s.makeImportDecisions(importableFiles, options)
	if err != nil {
		return nil, fmt.Errorf("failed to make import decisions: %w", err)
	}

	result := &models.FileImportResult{
		ImportDecisions: importDecisions,
		ProcessingTime:  time.Since(start),
	}

	// Process approved imports
	for _, decision := range importDecisions {
		switch decision.Decision {
		case models.ImportDecisionApproved:
			s.processApprovedImport(&decision, result)
		case models.ImportDecisionRejected:
			s.processRejectedImport(&decision, result)
		case models.ImportDecisionUnknown:
			// Unknown decisions are treated as skipped
			result.SkippedFiles = append(result.SkippedFiles, decision.Item)
		}
	}

	s.logger.Info("Import process completed",
		"path", path,
		"imported", len(result.ImportedFiles),
		"skipped", len(result.SkippedFiles),
		"rejected", len(result.RejectedFiles),
		"duration", result.ProcessingTime)

	return result, nil
}

// makeImportDecisions analyzes files and makes decisions about whether to import them
func (s *ImportService) makeImportDecisions(files []models.ImportableFile, options *ImportOptions) ([]models.ImportDecision, error) {
	var decisions []models.ImportDecision

	for _, file := range files {
		decision, err := s.makeImportDecision(file, options)
		if err != nil {
			s.logger.Error("Failed to make import decision", "file", file.Path, "error", err)
			decision = models.ImportDecision{
				Decision: models.ImportDecisionRejected,
				Item:     file,
				Rejections: []models.ImportRejection{
					{
						Reason: models.ImportRejectionReason(err.Error()),
						Type:   models.ImportRejectionTypePermanent,
					},
				},
			}
		}

		decisions = append(decisions, decision)
	}

	return decisions, nil
}

// makeImportDecision makes a decision about whether to import a specific file
func (s *ImportService) makeImportDecision(file models.ImportableFile, options *ImportOptions) (models.ImportDecision, error) {
	s.logger.Debug("Making import decision", "file", file.Path)

	decision := models.ImportDecision{
		Item: file,
	}

	// Check if it's a sample file
	if file.IsSample() {
		decision.Decision = models.ImportDecisionRejected
		decision.Rejections = append(decision.Rejections, models.ImportRejection{
			Reason: models.ImportRejectionSample,
			Type:   models.ImportRejectionTypePermanent,
		})
		return decision, nil
	}

	// Check if it's a video file
	if !file.IsVideoFile() {
		decision.Decision = models.ImportDecisionRejected
		decision.Rejections = append(decision.Rejections, models.ImportRejection{
			Reason: models.ImportRejectionReason("Not a video file"),
			Type:   models.ImportRejectionTypePermanent,
		})
		return decision, nil
	}

	// Attempt to identify the movie from the filename
	movie, err := s.identifyMovieFromFilename(file)
	if err != nil {
		decision.Decision = models.ImportDecisionRejected
		decision.Rejections = append(decision.Rejections, models.ImportRejection{
			Reason: models.ImportRejectionUnknownMovie,
			Type:   models.ImportRejectionTypeTemporary,
		})
		return decision, nil
	}

	decision.LocalMovie = movie
	decision.RemoteMovie = movie

	// Check if movie already has a file
	if existingFiles, err := s.movieFileService.GetByMovieID(movie.ID); err == nil && len(existingFiles) > 0 {
		// Check quality comparison
		isUpgrade, upgradeReason := s.checkQualityUpgrade(file, existingFiles[0])
		if !isUpgrade {
			decision.Decision = models.ImportDecisionRejected
			decision.Rejections = append(decision.Rejections, models.ImportRejection{
				Reason: models.ImportRejectionReason(upgradeReason),
				Type:   models.ImportRejectionTypeTemporary,
			})
			return decision, nil
		}
		decision.IsUpgrade = true
	}

	// Check if file already exists at the destination
	if exists, err := s.checkExistingFile(file, movie); err == nil && exists {
		decision.Decision = models.ImportDecisionRejected
		decision.Rejections = append(decision.Rejections, models.ImportRejection{
			Reason: models.ImportRejectionExistingFile,
			Type:   models.ImportRejectionTypePermanent,
		})
		return decision, nil
	}

	// All checks passed
	decision.Decision = models.ImportDecisionApproved

	s.logger.Debug("Import decision made",
		"file", file.Name,
		"decision", decision.Decision,
		"movie", movie.Title,
		"isUpgrade", decision.IsUpgrade)

	return decision, nil
}

// identifyMovieFromFilename attempts to identify a movie from its filename
func (s *ImportService) identifyMovieFromFilename(file models.ImportableFile) (*models.Movie, error) {
	// This is a simplified implementation
	// In reality, this would use sophisticated parsing to extract movie title and year

	fileName := file.GetFileName()
	cleanName := s.cleanMovieName(fileName)

	// Extract year from filename
	year := s.extractYearFromFilename(fileName)

	// Search for movies by title
	movies, err := s.movieService.Search(cleanName)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	if len(movies) == 0 {
		return nil, fmt.Errorf("no movies found for: %s", cleanName)
	}

	// Find best match based on title and year
	var bestMatch *models.Movie
	bestScore := 0

	for _, movie := range movies {
		score := s.calculateMovieMatchScore(movie, cleanName, year)
		if score > bestScore {
			bestScore = score
			bestMatch = &movie
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no good match found for: %s", cleanName)
	}

	return bestMatch, nil
}

// cleanMovieName cleans a filename to extract the likely movie title
func (s *ImportService) cleanMovieName(filename string) string {
	// Remove file extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Remove common patterns
	patterns := []string{
		`\b(19|20)\d{2}\b`,                             // Year
		`\b(720p|1080p|2160p|4k)\b`,                    // Resolution
		`\b(x264|x265|h264|h265|hevc|xvid|divx)\b`,     // Codecs
		`\b(bluray|bdrip|dvdrip|hdtv|webrip|web-dl)\b`, // Source
		`\b(aac|ac3|dts|mp3|flac)\b`,                   // Audio
		`\[\w+\]`,                                      // Release group in brackets
		`\(\w+\)`,                                      // Release group in parentheses
		`[-_.]`,                                        // Separators
	}

	cleaned := name
	for _, pattern := range patterns {
		cleaned = strings.ReplaceAll(strings.ToLower(cleaned), pattern, " ")
	}

	// Clean up spaces
	cleaned = strings.TrimSpace(cleaned)
	parts := strings.Fields(cleaned)

	// Take first reasonable number of words as title
	if len(parts) > 6 {
		parts = parts[:6]
	}

	return strings.Join(parts, " ")
}

// extractYearFromFilename extracts year from filename
func (s *ImportService) extractYearFromFilename(filename string) int {
	// Look for 4-digit year pattern
	yearPattern := `\b(19|20)(\d{2})\b`

	// This would be implemented with regex
	// For simplicity, returning 0 for now
	_ = yearPattern
	return 0
}

// calculateMovieMatchScore calculates how well a movie matches the search criteria
func (s *ImportService) calculateMovieMatchScore(movie models.Movie, searchTitle string, year int) int {
	score := 0

	// Title similarity (simplified)
	if strings.Contains(strings.ToLower(movie.Title), strings.ToLower(searchTitle)) {
		score += 10
	}

	// Year match
	if year > 0 && movie.Year == year {
		score += 5
	}

	return score
}

// checkQualityUpgrade determines if a new file is an upgrade over existing file
func (s *ImportService) checkQualityUpgrade(newFile models.ImportableFile, existingFile models.MovieFile) (bool, string) {
	// This would implement quality comparison logic
	// For simplicity, assume new file is always better if it's larger

	if newFile.Size > existingFile.Size {
		return true, ""
	}

	return false, "Existing file has better or equal quality"
}

// checkExistingFile checks if a file already exists at the intended destination
func (s *ImportService) checkExistingFile(file models.ImportableFile, movie *models.Movie) (bool, error) {
	// Get naming configuration
	namingConfig, err := s.namingService.GetNamingConfig()
	if err != nil {
		return false, err
	}

	// Build expected destination path
	expectedPath, err := s.namingService.BuildMovieFilePath(movie, nil, nil, namingConfig)
	if err != nil {
		return false, err
	}

	// Check if file exists
	_, err = os.Stat(expectedPath)
	return !os.IsNotExist(err), nil
}

// processApprovedImport processes an approved import decision
func (s *ImportService) processApprovedImport(decision *models.ImportDecision, result *models.FileImportResult) {
	s.logger.Info("Processing approved import", "file", decision.Item.Path, "movie", decision.LocalMovie.Title)

	// Get naming configuration
	namingConfig, err := s.namingService.GetNamingConfig()
	if err != nil {
		s.logger.Error("Failed to get naming config", "error", err)
		s.moveToErrored(decision, result, err.Error())
		return
	}

	// Extract media info
	mediaInfo, err := s.mediaInfoService.ExtractMediaInfo(decision.Item.Path)
	if err != nil {
		s.logger.Warn("Failed to extract media info", "file", decision.Item.Path, "error", err)
		// Continue without media info
	}

	// Organize the file
	orgResult, err := s.fileOrganizationService.OrganizeFile(
		decision.Item.Path,
		decision.LocalMovie,
		namingConfig,
		models.FileOperationMove,
	)

	if err != nil {
		s.logger.Error("Failed to organize file", "file", decision.Item.Path, "error", err)
		s.moveToErrored(decision, result, err.Error())
		return
	}

	if !orgResult.Success {
		s.logger.Error("File organization failed", "file", decision.Item.Path, "error", orgResult.Error)
		s.moveToErrored(decision, result, orgResult.Error)
		return
	}

	// Create movie file record
	movieFile := &models.MovieFile{
		MovieID:          decision.LocalMovie.ID,
		Path:             orgResult.OrganizedPath,
		RelativePath:     strings.TrimPrefix(orgResult.OrganizedPath, decision.LocalMovie.Path),
		Size:             decision.Item.Size,
		DateAdded:        time.Now(),
		OriginalFilePath: decision.Item.Path,
	}

	if mediaInfo != nil {
		movieFile.MediaInfo = *mediaInfo
	}

	// Save movie file
	if err := s.movieFileService.Create(movieFile); err != nil {
		s.logger.Error("Failed to create movie file record", "error", err)
		s.moveToErrored(decision, result, err.Error())
		return
	}

	// Update movie to mark it as having a file
	decision.LocalMovie.HasFile = true
	decision.LocalMovie.MovieFileID = movieFile.ID
	if err := s.movieService.Update(decision.LocalMovie); err != nil {
		s.logger.Error("Failed to update movie", "error", err)
		// Continue - file is imported, just movie record update failed
	}

	// Add to results
	result.ImportedFiles = append(result.ImportedFiles, *movieFile)
	result.ImportedSize += decision.Item.Size

	s.logger.Info("Successfully imported file",
		"originalPath", decision.Item.Path,
		"organizedPath", orgResult.OrganizedPath,
		"movie", decision.LocalMovie.Title)
}

// processRejectedImport processes a rejected import decision
func (s *ImportService) processRejectedImport(decision *models.ImportDecision, result *models.FileImportResult) {
	s.logger.Debug("Processing rejected import", "file", decision.Item.Path)

	result.RejectedFiles = append(result.RejectedFiles, decision.Item)
	result.RejectedSize += decision.Item.Size
}

// moveToErrored moves an import to the error list
func (s *ImportService) moveToErrored(decision *models.ImportDecision, result *models.FileImportResult, errorMsg string) {
	s.logger.Error("Moving import to errors", "file", decision.Item.Path, "error", errorMsg)

	result.ErrorFiles = append(result.ErrorFiles, decision.Item)
	result.ErrorSize += decision.Item.Size
}

// GetManualImports retrieves manual import records for interactive import
func (s *ImportService) GetManualImports(path string) ([]models.ManualImport, error) {
	// Scan the path for files
	importableFiles, err := s.fileOrganizationService.ScanDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	var manualImports []models.ManualImport

	for _, file := range importableFiles {
		manualImport := models.ManualImport{
			Path:       file.Path,
			Name:       file.Name,
			Size:       file.Size,
			FolderName: file.FolderName,
		}

		// Try to identify the movie
		if movie, err := s.identifyMovieFromFilename(file); err == nil {
			manualImport.MovieID = &movie.ID
			manualImport.Movie = movie
		}

		// Extract quality information
		// This would be more sophisticated in reality
		manualImport.Quality = models.Quality{
			Quality: models.QualityDefinition{
				ID:   1,
				Name: "Unknown",
			},
		}

		manualImports = append(manualImports, manualImport)
	}

	return manualImports, nil
}

// ProcessManualImport processes a manual import with user-provided details
func (s *ImportService) ProcessManualImport(manualImport *models.ManualImport) error {
	s.logger.Info("Processing manual import", "file", manualImport.Path)

	if manualImport.Movie == nil {
		return fmt.Errorf("movie must be specified for manual import")
	}

	// Create import decision
	file := models.ImportableFile{
		Path: manualImport.Path,
		Name: manualImport.Name,
		Size: manualImport.Size,
	}

	decision := models.ImportDecision{
		LocalMovie:  manualImport.Movie,
		RemoteMovie: manualImport.Movie,
		Decision:    models.ImportDecisionApproved,
		Item:        file,
	}

	// Create result structure
	result := &models.FileImportResult{}

	// Process the import
	s.processApprovedImport(&decision, result)

	return nil
}

// ImportOptions configures import behavior
type ImportOptions struct {
	ImportMode           models.ImportDecisionType `json:"importMode"`
	ReplaceExistingFiles bool                      `json:"replaceExistingFiles"`
	SkipFreeSpaceCheck   bool                      `json:"skipFreeSpaceCheck"`
}
