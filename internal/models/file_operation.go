package models

import (
	"time"
)

// FileOperationRecord represents a file system operation being tracked
type FileOperationRecord struct {
	ID              int                 `json:"id" gorm:"primaryKey;autoIncrement"`
	OperationType   FileOperationType   `json:"operationType" gorm:"not null"`
	SourcePath      string              `json:"sourcePath" gorm:"not null"`
	DestinationPath string              `json:"destinationPath"`
	MovieID         *int                `json:"movieId,omitempty" gorm:"index"`
	Movie           *Movie              `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	Status          FileOperationStatus `json:"status" gorm:"default:'pending'"`
	Progress        float64             `json:"progress" gorm:"default:0"`
	Size            int64               `json:"size" gorm:"default:0"`
	BytesProcessed  int64               `json:"bytesProcessed" gorm:"default:0"`
	ErrorMessage    string              `json:"errorMessage"`
	StartedAt       *time.Time          `json:"startedAt"`
	CompletedAt     *time.Time          `json:"completedAt"`
	CreatedAt       time.Time           `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time           `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name
func (FileOperationRecord) TableName() string {
	return "file_operations"
}

// FileOperationType represents the type of file operation
type FileOperationType string

// File operation type constants
const (
	FileOperationTypeImport    FileOperationType = "import"    // Import operation
	FileOperationTypeOrganize  FileOperationType = "organize"  // Organization operation
	FileOperationTypeMove      FileOperationType = "move"      // Move operation
	FileOperationTypeCopy      FileOperationType = "copy"      // Copy operation
	FileOperationTypeHardlink  FileOperationType = "hardlink"  // Hard link operation
	FileOperationTypeDelete    FileOperationType = "delete"    // Delete operation
	FileOperationTypeRename    FileOperationType = "rename"    // Rename operation
	FileOperationTypeMediaInfo FileOperationType = "mediainfo" // Media info operation
)

// FileOperationStatus represents the status of a file operation
type FileOperationStatus string

// File operation status constants
const (
	FileOperationStatusPending    FileOperationStatus = "pending"    // Operation is pending
	FileOperationStatusProcessing FileOperationStatus = "processing" // Operation is processing
	FileOperationStatusCompleted  FileOperationStatus = "completed"  // Operation completed successfully
	FileOperationStatusFailed     FileOperationStatus = "failed"     // Operation failed
	FileOperationStatusCanceled   FileOperationStatus = "canceled"   // Operation was canceled
)

// IsCompleted returns true if the operation has completed successfully
func (f *FileOperationRecord) IsCompleted() bool {
	return f.Status == FileOperationStatusCompleted
}

// IsFailed returns true if the operation has failed
func (f *FileOperationRecord) IsFailed() bool {
	return f.Status == FileOperationStatusFailed
}

// IsProcessing returns true if the operation is currently being processed
func (f *FileOperationRecord) IsProcessing() bool {
	return f.Status == FileOperationStatusProcessing
}

// CanCancel returns true if the operation can be canceled
func (f *FileOperationRecord) CanCancel() bool {
	return f.Status == FileOperationStatusPending || f.Status == FileOperationStatusProcessing
}

// MarkAsStarted marks the operation as started
func (f *FileOperationRecord) MarkAsStarted() {
	f.Status = FileOperationStatusProcessing
	now := time.Now()
	f.StartedAt = &now
}

// MarkAsCompleted marks the operation as completed successfully
func (f *FileOperationRecord) MarkAsCompleted() {
	f.Status = FileOperationStatusCompleted
	f.Progress = 100.0
	now := time.Now()
	f.CompletedAt = &now
}

// MarkAsFailed marks the operation as failed with an error message
func (f *FileOperationRecord) MarkAsFailed(errorMsg string) {
	f.Status = FileOperationStatusFailed
	f.ErrorMessage = errorMsg
	now := time.Now()
	f.CompletedAt = &now
}

// MarkAsCanceled marks the operation as canceled
func (f *FileOperationRecord) MarkAsCanceled() {
	f.Status = FileOperationStatusCanceled
	now := time.Now()
	f.CompletedAt = &now
}

// UpdateProgress updates the progress of the operation
func (f *FileOperationRecord) UpdateProgress(bytesProcessed int64) {
	f.BytesProcessed = bytesProcessed
	if f.Size > 0 {
		f.Progress = float64(bytesProcessed) / float64(f.Size) * 100.0
		if f.Progress > 100.0 {
			f.Progress = 100.0
		}
	}
}

// GetDuration returns the duration of the operation
func (f *FileOperationRecord) GetDuration() *time.Duration {
	if f.StartedAt == nil {
		return nil
	}

	endTime := time.Now()
	if f.CompletedAt != nil {
		endTime = *f.CompletedAt
	}

	duration := endTime.Sub(*f.StartedAt)
	return &duration
}

// GetStatusDisplay returns a human-readable status
func (f *FileOperationRecord) GetStatusDisplay() string {
	switch f.Status {
	case FileOperationStatusPending:
		return "Pending"
	case FileOperationStatusProcessing:
		return "Processing"
	case FileOperationStatusCompleted:
		return "Completed"
	case FileOperationStatusFailed:
		return "Failed"
	case FileOperationStatusCanceled:
		return "Canceled"
	default:
		return string(f.Status)
	}
}

// FileOperationSummary provides a summary of file operations
type FileOperationSummary struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Canceled   int `json:"canceled"`
}
