package services

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// FileOrganizationService provides operations for organizing and managing movie files
type FileOrganizationService struct {
	db               *database.Database
	logger           *logger.Logger
	namingService    *NamingService
	mediaInfoService *MediaInfoService
}

// NewFileOrganizationService creates a new instance of FileOrganizationService
func NewFileOrganizationService(
	db *database.Database,
	logger *logger.Logger,
	namingService *NamingService,
	mediaInfoService *MediaInfoService,
) *FileOrganizationService {
	return &FileOrganizationService{
		db:               db,
		logger:           logger,
		namingService:    namingService,
		mediaInfoService: mediaInfoService,
	}
}

// OrganizeFile organizes a single file according to naming configuration
func (s *FileOrganizationService) OrganizeFile(
	sourcePath string,
	movie *models.Movie,
	namingConfig *models.NamingConfig,
	operation models.FileOperation,
) (*models.FileOrganizationResult, error) {
	start := time.Now()
	s.logger.Info("Starting file organization", "source", sourcePath, "movie", movie.Title)

	fileOrg, err := s.initializeFileOrganization(sourcePath, movie, operation)
	if err != nil {
		return s.buildFailureResult(sourcePath, err.Error()), err
	}

	mediaInfo, destinationPath, err := s.prepareOrganization(fileOrg, movie, namingConfig, sourcePath)
	if err != nil {
		s.handleOrganizationFailure(fileOrg, err.Error())
		return s.buildFailureResult(sourcePath, err.Error()), err
	}

	movieFile, err := s.executeFileOperation(sourcePath, destinationPath, operation, namingConfig)
	if err != nil {
		s.handleOrganizationFailure(fileOrg, fmt.Sprintf("File operation failed: %v", err))
		return s.buildFailureResult(sourcePath, err.Error()), err
	}

	s.finalizeOrganization(fileOrg, movieFile, movie, mediaInfo, destinationPath)

	processingTime := time.Since(start)
	s.logOrganizationCompletion(sourcePath, destinationPath, processingTime)

	return s.buildSuccessResult(sourcePath, destinationPath, movie, movieFile, operation, processingTime), nil
}

// initializeFileOrganization creates and initializes file organization record
func (s *FileOrganizationService) initializeFileOrganization(
	sourcePath string, movie *models.Movie, operation models.FileOperation,
) (*models.FileOrganization, error) {
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		s.logger.Error("Failed to get file info", "path", sourcePath, "error", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileOrg := &models.FileOrganization{
		SourcePath:       sourcePath,
		MovieID:          &movie.ID,
		Operation:        operation,
		OriginalFileName: filepath.Base(sourcePath),
		Size:             fileInfo.Size(),
	}

	// Parse quality from filename
	quality, err := s.parseQualityFromFilename(sourcePath)
	if err == nil && quality != nil {
		fileOrg.Quality = quality
	}

	return fileOrg, nil
}

// prepareOrganization handles preparation steps for file organization
func (s *FileOrganizationService) prepareOrganization(
	fileOrg *models.FileOrganization, movie *models.Movie,
	namingConfig *models.NamingConfig, sourcePath string,
) (*models.MediaInfo, string, error) {
	// Extract media info if enabled
	var mediaInfo *models.MediaInfo
	if namingConfig.EnableMediaInfo {
		mediaInfo, _ = s.mediaInfoService.ExtractMediaInfo(sourcePath)
	}

	// Generate destination path
	destinationPath, err := s.namingService.BuildMovieFilePath(movie, fileOrg.Quality, mediaInfo, namingConfig)
	if err != nil {
		s.logger.Error("Failed to build destination path", "error", err)
		return nil, "", fmt.Errorf("failed to build destination path: %w", err)
	}

	fileOrg.DestinationPath = destinationPath
	fileOrg.OrganizedFileName = filepath.Base(destinationPath)

	// Save and mark as processing
	if err := s.saveAndMarkProcessing(fileOrg); err != nil {
		return nil, "", err
	}

	// Create destination directory
	if err := s.createDestinationDirectory(destinationPath); err != nil {
		return nil, "", err
	}

	return mediaInfo, destinationPath, nil
}

// saveAndMarkProcessing saves initial record and marks as processing
func (s *FileOrganizationService) saveAndMarkProcessing(
	fileOrg *models.FileOrganization,
) error {
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save file organization record", "error", err)
		return err
	}

	fileOrg.MarkAsProcessing()
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save file organization processing record", "error", err)
		return err
	}

	return nil
}

// createDestinationDirectory creates the destination directory if needed
func (s *FileOrganizationService) createDestinationDirectory(destinationPath string) error {
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, 0750); err != nil {
		s.logger.Error("Failed to create destination directory", "dir", destDir, "error", err)
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	return nil
}

// executeFileOperation performs the specified file operation
func (s *FileOrganizationService) executeFileOperation(
	sourcePath, destinationPath string, operation models.FileOperation,
	namingConfig *models.NamingConfig,
) (*models.MovieFile, error) {
	switch operation {
	case models.FileOperationMove:
		return s.moveFile(sourcePath, destinationPath, namingConfig)
	case models.FileOperationCopy:
		return s.copyFile(sourcePath, destinationPath, namingConfig)
	case models.FileOperationHardlink:
		return s.hardlinkFile(sourcePath, destinationPath, namingConfig)
	case models.FileOperationSymlink:
		return s.symlinkFile(sourcePath, destinationPath, namingConfig)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// finalizeOrganization completes the organization process
func (s *FileOrganizationService) finalizeOrganization(
	fileOrg *models.FileOrganization, movieFile *models.MovieFile,
	movie *models.Movie, mediaInfo *models.MediaInfo, destinationPath string,
) {
	// Update movie file record
	if movieFile != nil {
		movieFile.MovieID = movie.ID
		if fileOrg.Quality != nil {
			movieFile.Quality = *fileOrg.Quality
		}
		if mediaInfo != nil {
			movieFile.MediaInfo = *mediaInfo
		}
	}

	// Mark as completed
	fileOrg.MarkAsCompleted(destinationPath)
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save completed file organization record", "error", err)
	}
}

// handleOrganizationFailure handles failure scenarios
func (s *FileOrganizationService) handleOrganizationFailure(fileOrg *models.FileOrganization, errorMsg string) {
	fileOrg.MarkAsFailed(errorMsg)
	if saveErr := s.saveFileOrganization(fileOrg); saveErr != nil {
		s.logger.Error("Failed to save file organization failure record", "error", saveErr)
	}
}

// buildFailureResult creates a failure result
func (s *FileOrganizationService) buildFailureResult(sourcePath, errorMsg string) *models.FileOrganizationResult {
	return &models.FileOrganizationResult{
		OriginalPath: sourcePath,
		Success:      false,
		Error:        errorMsg,
	}
}

// buildSuccessResult creates a success result
func (s *FileOrganizationService) buildSuccessResult(
	sourcePath, destinationPath string, movie *models.Movie,
	movieFile *models.MovieFile, operation models.FileOperation,
	processingTime time.Duration,
) *models.FileOrganizationResult {
	return &models.FileOrganizationResult{
		OriginalPath:     sourcePath,
		OrganizedPath:    destinationPath,
		Movie:            movie,
		MovieFile:        movieFile,
		Success:          true,
		OrganizationType: string(operation),
		ProcessingTime:   processingTime,
	}
}

// logOrganizationCompletion logs successful completion
func (s *FileOrganizationService) logOrganizationCompletion(
	sourcePath, destinationPath string, processingTime time.Duration,
) {
	s.logger.Info("File organization completed",
		"source", sourcePath,
		"destination", destinationPath,
		"duration", processingTime)
}

// moveFile moves a file from source to destination
func (s *FileOrganizationService) moveFile(sourcePath, destPath string, config *models.NamingConfig) (*models.MovieFile, error) {
	// Validate file paths for security
	if err := s.validateFilePath(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}
	if err := s.validateFilePath(destPath); err != nil {
		return nil, fmt.Errorf("invalid destination path: %w", err)
	}

	// Create backup if configured
	if config != nil {
		// Note: backup logic would be implemented based on config
		_ = config // Avoid unused variable warning
	}

	// Move the file
	if err := os.Rename(sourcePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	// Set permissions if configured
	if config != nil {
		// Note: permission setting would be implemented based on MediaManagementConfig
	}

	// Create movie file record
	movieFile := &models.MovieFile{
		Path:             destPath,
		RelativePath:     destPath, // Would be calculated relative to root folder
		DateAdded:        time.Now(),
		OriginalFilePath: sourcePath,
	}

	// Get file info
	if fileInfo, err := os.Stat(destPath); err == nil {
		// Safe assignment - both are int64
		movieFile.Size = fileInfo.Size()
	}

	return movieFile, nil
}

// validateFilePath validates that a file path is safe and doesn't contain directory traversal attacks
func (s *FileOrganizationService) validateFilePath(filePath string) error {
	// Early validation for empty paths
	if filePath == "" {
		return fmt.Errorf("empty file path not allowed")
	}

	// Check for obvious directory traversal attempts before cleaning
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path contains directory traversal attempt: %s", filePath)
	}

	// Clean the path to resolve any ".." or "." components - safe after validation
	cleanPath := filepath.Clean(filePath)

	// Double-check for directory traversal attempts after cleaning
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal attempt after cleaning: %s", filePath)
	}

	// Ensure the path is absolute or within expected bounds
	if filepath.IsAbs(cleanPath) {
		return nil
	}

	// For relative paths, ensure they don't escape the working directory
	if strings.HasPrefix(cleanPath, "../") || cleanPath == ".." {
		return fmt.Errorf("relative path escapes working directory: %s", filePath)
	}

	return nil
}

// copyFile copies a file from source to destination
func (s *FileOrganizationService) copyFile(sourcePath, destPath string, config *models.NamingConfig) (*models.MovieFile, error) {
	// Validate file paths for security
	if err := s.validateFilePath(sourcePath); err != nil {
		return nil, fmt.Errorf("invalid source path: %w", err)
	}
	if err := s.validateFilePath(destPath); err != nil {
		return nil, fmt.Errorf("invalid destination path: %w", err)
	}

	sourceFile, err := os.Open(sourcePath) // #nosec G304 - path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Note: This is in a defer, so we can't return the error
			// Log it if we had a logger available in this context
		}
	}()

	destFile, err := os.Create(destPath) // #nosec G304 - path validated above
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			// Note: This is in a defer, so we can't return the error
			// Log it if we had a logger available in this context
		}
	}()

	// Copy file contents
	if _, err := sourceFile.Seek(0, 0); err != nil {
		return nil, err
	}

	buffer := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := sourceFile.Read(buffer)
		if n > 0 {
			if _, writeErr := destFile.Write(buffer[:n]); writeErr != nil {
				return nil, fmt.Errorf("failed to write to destination: %w", writeErr)
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read source file: %w", err)
		}
	}

	// Copy file permissions
	if sourceInfo, err := sourceFile.Stat(); err == nil {
		if err := os.Chmod(destPath, sourceInfo.Mode()); err != nil {
			// Log but don't fail - permissions are not critical for functionality
			s.logger.Warn("Failed to set file permissions", "path", destPath, "error", err)
		}
	}

	// Create movie file record
	movieFile := &models.MovieFile{
		Path:             destPath,
		RelativePath:     destPath,
		DateAdded:        time.Now(),
		OriginalFilePath: sourcePath,
	}

	if fileInfo, err := os.Stat(destPath); err == nil {
		// Safe assignment - both are int64
		movieFile.Size = fileInfo.Size()
	}

	return movieFile, nil
}

// hardlinkFile creates a hard link from source to destination
func (s *FileOrganizationService) hardlinkFile(sourcePath, destPath string, config *models.NamingConfig) (*models.MovieFile, error) {
	// Create hard link
	if err := os.Link(sourcePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to create hard link: %w", err)
	}

	// Create movie file record
	movieFile := &models.MovieFile{
		Path:             destPath,
		RelativePath:     destPath,
		DateAdded:        time.Now(),
		OriginalFilePath: sourcePath,
	}

	if fileInfo, err := os.Stat(destPath); err == nil {
		// Safe assignment - both are int64
		movieFile.Size = fileInfo.Size()
	}

	return movieFile, nil
}

func (s *FileOrganizationService) symlinkFile(
	sourcePath, destPath string, _ *models.NamingConfig,
) (*models.MovieFile, error) {
	// Create symbolic link
	if err := os.Symlink(sourcePath, destPath); err != nil {
		return nil, fmt.Errorf("failed to create symbolic link: %w", err)
	}

	// Create movie file record
	movieFile := &models.MovieFile{
		Path:             destPath,
		RelativePath:     destPath,
		DateAdded:        time.Now(),
		OriginalFilePath: sourcePath,
	}

	if fileInfo, err := os.Stat(sourcePath); err == nil {
		movieFile.Size = fileInfo.Size()
	}

	return movieFile, nil
}

// CheckFreeSpace checks if there's enough free space for the operation
func (s *FileOrganizationService) CheckFreeSpace(destPath string, fileSize int64, config *models.NamingConfig) error {
	if config.SkipFreeSpaceCheck {
		return nil
	}

	// Get disk usage for destination path
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Dir(destPath), &stat); err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	// Calculate available space with overflow protection
	bavail := stat.Bavail // uint64
	bsize := stat.Bsize   // uint32

	// Check for overflow in multiplication
	if bavail > 0 && bsize > 0 && bavail > uint64(math.MaxInt64)/uint64(bsize) {
		// Handle overflow by using MaxInt64 as available space
		availableBytes := int64(math.MaxInt64)
		requiredBytes := fileSize + (config.MinimumFreeSpace * 1024 * 1024) // Convert MB to bytes

		if availableBytes < requiredBytes {
			return fmt.Errorf("insufficient disk space: available %d bytes, required %d bytes",
				availableBytes, requiredBytes)
		}
		return nil
	}

	// Safe conversion after overflow check
	product := bavail * uint64(bsize)
	if product > uint64(math.MaxInt64) {
		availableBytes := int64(math.MaxInt64)
		requiredBytes := fileSize + (config.MinimumFreeSpace * 1024 * 1024) // Convert MB to bytes

		if availableBytes < requiredBytes {
			return fmt.Errorf("insufficient disk space: available %d bytes, required %d bytes",
				availableBytes, requiredBytes)
		}
		return nil
	}

	availableBytes := int64(product)
	requiredBytes := fileSize + (config.MinimumFreeSpace * 1024 * 1024) // Convert MB to bytes

	if availableBytes < requiredBytes {
		return fmt.Errorf("insufficient disk space: available %d bytes, required %d bytes",
			availableBytes, requiredBytes)
	}

	return nil
}

// GetFileOrganizations retrieves file organization records
func (s *FileOrganizationService) GetFileOrganizations(limit, offset int) ([]models.FileOrganization, error) {
	var organizations []models.FileOrganization

	query := s.db.GORM.Order("created_at desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&organizations).Error; err != nil {
		return nil, fmt.Errorf("failed to get file organizations: %w", err)
	}

	return organizations, nil
}

// GetFileOrganizationByID retrieves a specific file organization record
func (s *FileOrganizationService) GetFileOrganizationByID(id int) (*models.FileOrganization, error) {
	var organization models.FileOrganization

	if err := s.db.GORM.Where("id = ?", id).First(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to get file organization: %w", err)
	}

	return &organization, nil
}

// RetryFailedOrganizations retries failed file organization operations
func (s *FileOrganizationService) RetryFailedOrganizations() error {
	var failedOrganizations []models.FileOrganization

	// Get failed organizations that can be retried
	if err := s.db.GORM.Where("status = ? AND attempt_count < ?",
		models.OrganizationStatusFailed, 3).Find(&failedOrganizations).Error; err != nil {
		return fmt.Errorf("failed to get failed organizations: %w", err)
	}

	s.logger.Info("Retrying failed file organizations", "count", len(failedOrganizations))

	for _, org := range failedOrganizations {
		if !org.CanRetry() {
			continue
		}

		s.logger.Info("Retrying file organization", "id", org.ID, "path", org.SourcePath)

		// Check if source file still exists
		if _, err := os.Stat(org.SourcePath); os.IsNotExist(err) {
			org.MarkAsFailed("Source file no longer exists")
			if err := s.saveFileOrganization(&org); err != nil {
				s.logger.Error("Failed to save file organization record", "error", err)
			}
			continue
		}

		// Reset status and retry
		org.Status = models.OrganizationStatusPending
		org.ErrorMessage = ""
		if err := s.saveFileOrganization(&org); err != nil {
			s.logger.Error("Failed to update organization for retry", "id", org.ID, "error", err)
		}
	}

	return nil
}

// CleanupOldOrganizations removes old completed file organization records
func (s *FileOrganizationService) CleanupOldOrganizations(olderThanDays int) error {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	result := s.db.GORM.Where("status = ? AND created_at < ?",
		models.OrganizationStatusCompleted, cutoff).Delete(&models.FileOrganization{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old organizations: %w", result.Error)
	}

	s.logger.Info("Cleaned up old file organizations", "deleted", result.RowsAffected)
	return nil
}

// ScanDirectory scans a directory for importable files
func (s *FileOrganizationService) ScanDirectory(path string) ([]models.ImportableFile, error) {
	var importableFiles []models.ImportableFile

	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if it's a video file
		ext := strings.ToLower(filepath.Ext(filePath))
		if !s.isVideoFile(ext) {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			s.logger.Warn("Failed to get file info", "path", filePath, "error", err)
			return nil
		}

		// Create importable file
		importableFile := models.ImportableFile{
			Path:         filePath,
			RelativePath: strings.TrimPrefix(filePath, path),
			Name:         d.Name(),
			Size:         fileInfo.Size(),
			DateModified: fileInfo.ModTime(),
			FolderName:   filepath.Base(filepath.Dir(filePath)),
		}

		// Skip samples
		if importableFile.IsSample() {
			s.logger.Debug("Skipping sample file", "path", filePath)
			return nil
		}

		importableFiles = append(importableFiles, importableFile)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	s.logger.Info("Scanned directory for importable files",
		"path", path,
		"found", len(importableFiles))

	return importableFiles, nil
}

// parseQualityFromFilename attempts to parse quality information from filename
func (s *FileOrganizationService) parseQualityFromFilename(filename string) (*models.Quality, error) {
	// This would implement quality parsing logic similar to Radarr's parser
	// For now, return a basic implementation
	name := strings.ToLower(filepath.Base(filename))

	// Basic quality detection
	var qualityDef models.QualityDefinition

	switch {
	case strings.Contains(name, "2160p") || strings.Contains(name, "4k"):
		qualityDef = models.QualityDefinition{ID: 7, Name: "Bluray-2160p", Source: "bluray", Resolution: 2160}
	case strings.Contains(name, "1080p"):
		qualityDef = models.QualityDefinition{ID: 6, Name: "Bluray-1080p", Source: "bluray", Resolution: 1080}
	case strings.Contains(name, "720p"):
		qualityDef = models.QualityDefinition{ID: 4, Name: "HDTV-720p", Source: "tv", Resolution: 720}
	default:
		qualityDef = models.QualityDefinition{ID: 1, Name: "Unknown", Source: "unknown", Resolution: 0}
	}

	return &models.Quality{
		Quality:  qualityDef,
		Revision: models.Revision{Version: 1, Real: 0, IsRepack: false},
	}, nil
}

// isVideoFile checks if a file extension indicates a video file
func (s *FileOrganizationService) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".mkv", ".avi", ".wmv", ".mov", ".flv", ".m4v", ".mpg", ".mpeg", ".ts", ".webm"}

	for _, validExt := range videoExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// saveFileOrganization saves a file organization record to the database
func (s *FileOrganizationService) saveFileOrganization(org *models.FileOrganization) error {
	if org.ID == 0 {
		return s.db.GORM.Create(org).Error
	}
	return s.db.GORM.Save(org).Error
}
