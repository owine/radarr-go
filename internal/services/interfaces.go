// Package services defines interfaces for dependency injection and testing
package services

import "github.com/radarr/radarr-go/internal/models"

// MovieServiceInterface defines the interface for movie operations
type MovieServiceInterface interface {
	GetByID(id int) (*models.Movie, error)
	GetAll() ([]models.Movie, error)
	Update(movie *models.Movie) error
}

// MetadataServiceInterface defines the interface for metadata operations
type MetadataServiceInterface interface {
	RefreshMovieMetadata(movieID int) error
}

// ImportListServiceInterface defines the interface for import list operations
type ImportListServiceInterface interface {
	GetImportListByID(id int) (*models.ImportList, error)
	GetEnabledImportLists() ([]models.ImportList, error)
	SyncImportList(id int) (*models.ImportListSyncResult, error)
}
