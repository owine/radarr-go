package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// FileOrganization represents a file organization operation
type FileOrganization struct {
	ID                int                `json:"id" gorm:"primaryKey;autoIncrement"`
	SourcePath        string             `json:"sourcePath" gorm:"not null"`
	DestinationPath   string             `json:"destinationPath" gorm:"not null"`
	MovieID           *int               `json:"movieId,omitempty" gorm:"index"`
	Status            OrganizationStatus `json:"status" gorm:"default:'pending'"`
	StatusMessage     string             `json:"statusMessage"`
	Operation         FileOperation      `json:"operation" gorm:"default:'move'"`
	Size              int64              `json:"size"`
	Quality           *Quality           `json:"quality,omitempty" gorm:"type:text"`
	Languages         []Language         `json:"languages" gorm:"type:text"`
	ReleaseGroup      string             `json:"releaseGroup"`
	Edition           string             `json:"edition"`
	OriginalFileName  string             `json:"originalFileName"`
	OrganizedFileName string             `json:"organizedFileName"`
	BackupPath        string             `json:"backupPath"`
	ProcessedAt       *time.Time         `json:"processedAt"`
	ErrorMessage      string             `json:"errorMessage"`
	AttemptCount      int                `json:"attemptCount" gorm:"default:0"`
	LastAttemptAt     *time.Time         `json:"lastAttemptAt"`
	CreatedAt         time.Time          `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time          `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for FileOrganization
func (FileOrganization) TableName() string {
	return "file_organizations"
}

// OrganizationStatus represents the status of a file organization operation
type OrganizationStatus string

const (
	OrganizationStatusPending    OrganizationStatus = "pending"
	OrganizationStatusProcessing OrganizationStatus = "processing"
	OrganizationStatusCompleted  OrganizationStatus = "completed"
	OrganizationStatusFailed     OrganizationStatus = "failed"
	OrganizationStatusSkipped    OrganizationStatus = "skipped"
)

// FileOperation represents the type of file operation to perform
type FileOperation string

const (
	FileOperationMove     FileOperation = "move"
	FileOperationCopy     FileOperation = "copy"
	FileOperationHardlink FileOperation = "hardlink"
	FileOperationSymlink  FileOperation = "symlink"
)

// ImportDecision represents a decision about importing a file
type ImportDecision struct {
	LocalMovie   *Movie             `json:"localMovie"`
	RemoteMovie  *Movie             `json:"remoteMovie"`
	Decision     ImportDecisionType `json:"decision"`
	Rejections   []ImportRejection  `json:"rejections"`
	DownloadItem *QueueItem         `json:"downloadItem,omitempty"`
	IsUpgrade    bool               `json:"isUpgrade"`
	Item         ImportableFile     `json:"item"`
}

// ImportDecisionType represents the type of import decision
type ImportDecisionType string

const (
	ImportDecisionApproved ImportDecisionType = "approved"
	ImportDecisionRejected ImportDecisionType = "rejected"
	ImportDecisionUnknown  ImportDecisionType = "unknown"
)

// ImportRejection represents a reason why an import was rejected
type ImportRejection struct {
	Reason ImportRejectionReason `json:"reason"`
	Type   ImportRejectionType   `json:"type"`
}

// ImportRejectionReason represents specific rejection reasons
type ImportRejectionReason string

const (
	ImportRejectionUnknownMovie      ImportRejectionReason = "Unknown Movie"
	ImportRejectionExistingFile      ImportRejectionReason = "Existing File"
	ImportRejectionSameFile          ImportRejectionReason = "Same File"
	ImportRejectionQualityCutoff     ImportRejectionReason = "Quality Cutoff"
	ImportRejectionBetterQuality     ImportRejectionReason = "Better Quality Available"
	ImportRejectionUnwantedLanguage  ImportRejectionReason = "Unwanted Language"
	ImportRejectionUnwantedQuality   ImportRejectionReason = "Unwanted Quality"
	ImportRejectionTorrentNotSeeding ImportRejectionReason = "Torrent Not Seeding"
	ImportRejectionInvalidPath       ImportRejectionReason = "Invalid Path"
	ImportRejectionFileNotFound      ImportRejectionReason = "File Not Found"
	ImportRejectionAlreadyImported   ImportRejectionReason = "Already Imported"
	ImportRejectionSample            ImportRejectionReason = "Sample"
	ImportRejectionWrongMovie        ImportRejectionReason = "Wrong Movie"
	ImportRejectionHardlinkedFile    ImportRejectionReason = "Hardline File"
)

// ImportRejectionType represents the type of rejection
type ImportRejectionType string

const (
	ImportRejectionTypePermanent ImportRejectionType = "permanent"
	ImportRejectionTypeTemporary ImportRejectionType = "temporary"
)

// ImportableFile represents a file that can be imported
type ImportableFile struct {
	ID            int        `json:"id"`
	Path          string     `json:"path"`
	RelativePath  string     `json:"relativePath"`
	FolderName    string     `json:"folderName"`
	Name          string     `json:"name"`
	Size          int64      `json:"size"`
	DateModified  time.Time  `json:"dateModified"`
	Movie         *Movie     `json:"movie"`
	Quality       *Quality   `json:"quality"`
	QualityWeight int        `json:"qualityWeight"`
	Languages     []Language `json:"languages"`
	ReleaseGroup  string     `json:"releaseGroup"`
	Edition       string     `json:"edition"`
	SceneName     string     `json:"sceneName"`
	MediaInfo     *MediaInfo `json:"mediaInfo"`
	DownloadItem  *QueueItem `json:"downloadItem,omitempty"`
}

// FileImportResult represents the result of importing files
type FileImportResult struct {
	ImportDecisions []ImportDecision `json:"importDecisions"`
	ImportedFiles   []MovieFile      `json:"importedFiles"`
	ImportedSize    int64            `json:"importedSize"`
	SkippedFiles    []ImportableFile `json:"skippedFiles"`
	SkippedSize     int64            `json:"skippedSize"`
	RejectedFiles   []ImportableFile `json:"rejectedFiles"`
	RejectedSize    int64            `json:"rejectedSize"`
	ErrorFiles      []ImportableFile `json:"errorFiles"`
	ErrorSize       int64            `json:"errorSize"`
	ProcessingTime  time.Duration    `json:"processingTime"`
}

// FileOrganizationResult represents the result of organizing files
type FileOrganizationResult struct {
	OriginalPath     string        `json:"originalPath"`
	OrganizedPath    string        `json:"organizedPath"`
	Movie            *Movie        `json:"movie"`
	MovieFile        *MovieFile    `json:"movieFile"`
	Success          bool          `json:"success"`
	Error            string        `json:"error"`
	OrganizationType string        `json:"organizationType"`
	ProcessingTime   time.Duration `json:"processingTime"`
}

// ManualImport represents a manual import operation
type ManualImport struct {
	ID           int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Path         string               `json:"path" gorm:"not null"`
	Name         string               `json:"name"`
	Size         int64                `json:"size"`
	Quality      Quality              `json:"quality" gorm:"type:text"`
	Languages    []Language           `json:"languages" gorm:"type:text"`
	MovieID      *int                 `json:"movieId" gorm:"index"`
	DownloadID   string               `json:"downloadId"`
	FolderName   string               `json:"folderName"`
	SceneName    string               `json:"sceneName"`
	ReleaseGroup string               `json:"releaseGroup"`
	Edition      string               `json:"edition"`
	Movie        *Movie               `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	Rejections   ImportRejectionArray `json:"rejections" gorm:"type:text"`
	CreatedAt    time.Time            `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time            `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for ManualImport
func (ManualImport) TableName() string {
	return "manual_imports"
}

// ImportRejectionArray is a custom type for handling ImportRejection slices in GORM
type ImportRejectionArray []ImportRejection

// Value implements the driver.Valuer interface for ImportRejection slice
func (r ImportRejectionArray) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements the sql.Scanner interface for ImportRejection slice
func (r *ImportRejectionArray) Scan(value interface{}) error {
	if value == nil {
		*r = ImportRejectionArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, r)
}

// GetFileName returns the file name from the path
func (f *ImportableFile) GetFileName() string {
	return filepath.Base(f.Path)
}

// GetFileExtension returns the file extension
func (f *ImportableFile) GetFileExtension() string {
	return strings.ToLower(filepath.Ext(f.Path))
}

// IsVideoFile returns true if the file is a video file
func (f *ImportableFile) IsVideoFile() bool {
	videoExts := []string{".mp4", ".mkv", ".avi", ".wmv", ".mov", ".flv", ".m4v", ".mpg", ".mpeg", ".ts", ".webm"}
	ext := f.GetFileExtension()

	for _, validExt := range videoExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// IsSubtitleFile returns true if the file is a subtitle file
func (f *ImportableFile) IsSubtitleFile() bool {
	subtitleExts := []string{".srt", ".ass", ".ssa", ".sub", ".idx", ".sup", ".vtt"}
	ext := f.GetFileExtension()

	for _, validExt := range subtitleExts {
		if ext == validExt {
			return true
		}
	}

	return false
}

// IsSample returns true if the file appears to be a sample
func (f *ImportableFile) IsSample() bool {
	name := strings.ToLower(f.Name)

	// Check for common sample indicators
	sampleIndicators := []string{"sample", "preview", "trailer", "rarbg"}

	for _, indicator := range sampleIndicators {
		if strings.Contains(name, indicator) {
			return true
		}
	}

	// Check file size (samples are typically small)
	const sampleSizeThresholdMB = 150
	if f.Size > 0 && f.Size < sampleSizeThresholdMB*1024*1024 {
		return true
	}

	return false
}

// GetOrganizationQuality returns a displayable quality string
func (fo *FileOrganization) GetOrganizationQuality() string {
	if fo.Quality != nil {
		return fo.Quality.Quality.Name
	}
	return "Unknown"
}

// GetLanguageNames returns a comma-separated list of language names
func (fo *FileOrganization) GetLanguageNames() string {
	if len(fo.Languages) == 0 {
		return "Unknown"
	}

	var names []string
	for _, lang := range fo.Languages {
		names = append(names, lang.Name)
	}

	return strings.Join(names, ", ")
}

// IsProcessing returns true if the organization is currently being processed
func (fo *FileOrganization) IsProcessing() bool {
	return fo.Status == OrganizationStatusProcessing
}

// IsCompleted returns true if the organization completed successfully
func (fo *FileOrganization) IsCompleted() bool {
	return fo.Status == OrganizationStatusCompleted
}

// IsFailed returns true if the organization failed
func (fo *FileOrganization) IsFailed() bool {
	return fo.Status == OrganizationStatusFailed
}

// CanRetry returns true if the operation can be retried
func (fo *FileOrganization) CanRetry() bool {
	return fo.Status == OrganizationStatusFailed && fo.AttemptCount < 3
}

// MarkAsProcessing updates the status to processing
func (fo *FileOrganization) MarkAsProcessing() {
	fo.Status = OrganizationStatusProcessing
	now := time.Now()
	fo.LastAttemptAt = &now
	fo.AttemptCount++
}

// MarkAsCompleted updates the status to completed
func (fo *FileOrganization) MarkAsCompleted(destinationPath string) {
	fo.Status = OrganizationStatusCompleted
	fo.DestinationPath = destinationPath
	now := time.Now()
	fo.ProcessedAt = &now
	fo.ErrorMessage = ""
}

// MarkAsFailed updates the status to failed with an error message
func (fo *FileOrganization) MarkAsFailed(errorMsg string) {
	fo.Status = OrganizationStatusFailed
	fo.ErrorMessage = errorMsg
	fo.StatusMessage = fmt.Sprintf("Failed after %d attempts: %s", fo.AttemptCount, errorMsg)
}

// MarkAsSkipped updates the status to skipped with a reason
func (fo *FileOrganization) MarkAsSkipped(reason string) {
	fo.Status = OrganizationStatusSkipped
	fo.StatusMessage = reason
}

// GetStatusDisplay returns a human-readable status
func (fo *FileOrganization) GetStatusDisplay() string {
	switch fo.Status {
	case OrganizationStatusPending:
		return "Pending"
	case OrganizationStatusProcessing:
		return "Processing"
	case OrganizationStatusCompleted:
		return "Completed"
	case OrganizationStatusFailed:
		return fmt.Sprintf("Failed (%d/%d attempts)", fo.AttemptCount, 3)
	case OrganizationStatusSkipped:
		return "Skipped"
	default:
		return string(fo.Status)
	}
}
