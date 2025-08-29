package services

import (
	"fmt"
	"io/fs"
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

	// Create file organization record
	fileOrg := &models.FileOrganization{
		SourcePath:       sourcePath,
		MovieID:          &movie.ID,
		Operation:        operation,
		OriginalFileName: filepath.Base(sourcePath),
	}

	// Get file info
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		s.logger.Error("Failed to get file info", "path", sourcePath, "error", err)
		return &models.FileOrganizationResult{
			OriginalPath: sourcePath,
			Success:      false,
			Error:        fmt.Sprintf("Failed to get file info: %v", err),
		}, err
	}

	fileOrg.Size = fileInfo.Size()

	// Parse quality from filename
	quality, err := s.parseQualityFromFilename(sourcePath)
	if err == nil && quality != nil {
		fileOrg.Quality = quality
	}

	// Extract media info if enabled
	var mediaInfo *models.MediaInfo
	if namingConfig.EnableMediaInfo {
		mediaInfo, _ = s.mediaInfoService.ExtractMediaInfo(sourcePath)
	}

	// Generate destination path using naming service
	destinationPath, err := s.namingService.BuildMovieFilePath(movie, fileOrg.Quality, mediaInfo, namingConfig)
	if err != nil {
		s.logger.Error("Failed to build destination path", "error", err)
		fileOrg.MarkAsFailed(fmt.Sprintf("Failed to build destination path: %v", err))
		if saveErr := s.saveFileOrganization(fileOrg); saveErr != nil {
			s.logger.Error("Failed to save file organization failure record", "error", saveErr)
		}
		return &models.FileOrganizationResult{
			OriginalPath: sourcePath,
			Success:      false,
			Error:        err.Error(),
		}, err
	}

	fileOrg.DestinationPath = destinationPath
	fileOrg.OrganizedFileName = filepath.Base(destinationPath)

	// Save initial record
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save file organization record", "error", err)
	}

	// Mark as processing
	fileOrg.MarkAsProcessing()
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save file organization processing record", "error", err)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		s.logger.Error("Failed to create destination directory", "dir", destDir, "error", err)
		fileOrg.MarkAsFailed(fmt.Sprintf("Failed to create destination directory: %v", err))
		if saveErr := s.saveFileOrganization(fileOrg); saveErr != nil {
			s.logger.Error("Failed to save file organization failure record", "error", saveErr)
		}
		return &models.FileOrganizationResult{
			OriginalPath: sourcePath,
			Success:      false,
			Error:        err.Error(),
		}, err
	}

	// Perform the file operation
	var movieFile *models.MovieFile
	switch operation {
	case models.FileOperationMove:
		movieFile, err = s.moveFile(sourcePath, destinationPath, namingConfig)
	case models.FileOperationCopy:
		movieFile, err = s.copyFile(sourcePath, destinationPath, namingConfig)
	case models.FileOperationHardlink:
		movieFile, err = s.hardlinkFile(sourcePath, destinationPath, namingConfig)
	default:
		err = fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		s.logger.Error("File operation failed", "operation", operation, "error", err)
		fileOrg.MarkAsFailed(fmt.Sprintf("File operation failed: %v", err))
		if saveErr := s.saveFileOrganization(fileOrg); saveErr != nil {
			s.logger.Error("Failed to save file organization failure record", "error", saveErr)
		}
		return &models.FileOrganizationResult{
			OriginalPath: sourcePath,
			Success:      false,
			Error:        err.Error(),
		}, err
	}

	// Update movie file record
	if movieFile != nil {
		movieFile.MovieID = movie.ID
		movieFile.Quality = *fileOrg.Quality
		if mediaInfo != nil {
			movieFile.MediaInfo = *mediaInfo
		}
	}

	// Mark as completed
	fileOrg.MarkAsCompleted(destinationPath)
	if err := s.saveFileOrganization(fileOrg); err != nil {
		s.logger.Error("Failed to save completed file organization record", "error", err)
	}

	processingTime := time.Since(start)
	s.logger.Info("File organization completed",
		"source", sourcePath,
		"destination", destinationPath,
		"duration", processingTime)

	return &models.FileOrganizationResult{
		OriginalPath:     sourcePath,
		OrganizedPath:    destinationPath,
		Movie:            movie,
		MovieFile:        movieFile,
		Success:          true,
		OrganizationType: string(operation),
		ProcessingTime:   processingTime,
	}, nil
}

// moveFile moves a file from source to destination
func (s *FileOrganizationService) moveFile(sourcePath, destPath string, config *models.NamingConfig) (*models.MovieFile, error) {
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
		movieFile.Size = fileInfo.Size()
	}

	return movieFile, nil
}

// copyFile copies a file from source to destination
func (s *FileOrganizationService) copyFile(sourcePath, destPath string, config *models.NamingConfig) (*models.MovieFile, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			// Note: This is in a defer, so we can't return the error
			// Log it if we had a logger available in this context
		}
	}()

	destFile, err := os.Create(destPath)
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

	// Calculate available space
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)
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

	if strings.Contains(name, "2160p") || strings.Contains(name, "4k") {
		qualityDef = models.QualityDefinition{ID: 7, Name: "Bluray-2160p", Source: "bluray", Resolution: 2160}
	} else if strings.Contains(name, "1080p") {
		qualityDef = models.QualityDefinition{ID: 6, Name: "Bluray-1080p", Source: "bluray", Resolution: 1080}
	} else if strings.Contains(name, "720p") {
		qualityDef = models.QualityDefinition{ID: 4, Name: "HDTV-720p", Source: "tv", Resolution: 720}
	} else {
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
