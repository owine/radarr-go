package services

import (
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
)

type Container struct {
	DB     *database.Database
	Config *config.Config
	Logger *logger.Logger

	// Services
	MovieService        *MovieService
	MovieFileService    *MovieFileService
	QualityService      *QualityService
	IndexerService      *IndexerService
	DownloadService     *DownloadService
	NotificationService *NotificationService
}

func NewContainer(db *database.Database, cfg *config.Config, logger *logger.Logger) *Container {
	container := &Container{
		DB:     db,
		Config: cfg,
		Logger: logger,
	}

	// Initialize services
	container.MovieService = NewMovieService(db, logger)
	container.MovieFileService = NewMovieFileService(db, logger)
	container.QualityService = NewQualityService(db, logger)
	container.IndexerService = NewIndexerService(db, logger)
	container.DownloadService = NewDownloadService(db, logger)
	container.NotificationService = NewNotificationService(db, logger)

	return container
}
