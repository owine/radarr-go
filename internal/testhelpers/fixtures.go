package testhelpers

import (
	"time"

	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// TestDataFactory provides methods to create test data
type TestDataFactory struct {
	db *gorm.DB
}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory(db *gorm.DB) *TestDataFactory {
	return &TestDataFactory{db: db}
}

// CreateMovie creates a test movie with default values
func (f *TestDataFactory) CreateMovie(overrides ...func(*models.Movie)) *models.Movie {
	inCinemas := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
	digitalRelease := time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC)
	physicalRelease := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)

	movie := &models.Movie{
		TmdbID:              12345,
		Title:               "Test Movie",
		TitleSlug:           "test-movie",
		Year:                2023,
		Overview:            "A test movie for testing purposes",
		Runtime:             120,
		Status:              models.MovieStatusReleased,
		InCinemas:           &inCinemas,
		DigitalRelease:      &digitalRelease,
		PhysicalRelease:     &physicalRelease,
		Monitored:           true,
		MinimumAvailability: models.AvailabilityReleased,
		IsAvailable:         true,
		HasFile:             false,
		QualityProfileID:    1,
		Path:                "/movies/test-movie-2023",
		RootFolderPath:      "/movies",
		FolderName:          "test-movie-2023",
		SizeOnDisk:          0,
		Genres:              models.StringArray{"Action", "Adventure"},
		Tags:                models.IntArray{},
		Images: models.MediaCover{
			{
				CoverType: "poster",
				URL:       "https://example.com/poster.jpg",
			},
		},
		Website:          "https://example.com/movie",
		YouTubeTrailerID: "abc123",
		Studio:           "Test Studios",
		ImdbID:           "tt1234567",
		Certification:    "PG-13",
		Collection: &models.Collection{
			Name:   "Test Collection",
			TmdbID: 54321,
		},
	}

	// Apply any overrides
	for _, override := range overrides {
		override(movie)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(movie).Error; err != nil {
			panic(err)
		}
	}

	return movie
}

// CreateMovieFile creates a test movie file with default values
func (f *TestDataFactory) CreateMovieFile(movieID int, overrides ...func(*models.MovieFile)) *models.MovieFile {
	movieFile := &models.MovieFile{
		MovieID:      movieID,
		RelativePath: "test-movie-2023.mkv",
		Path:         "/movies/test-movie-2023/test-movie-2023.mkv",
		Size:         1024 * 1024 * 1024, // 1GB
		DateAdded:    time.Now(),
		SceneName:    "Test.Movie.2023.1080p.BluRay.x264-GROUP",
		ReleaseGroup: "GROUP",
		Quality: models.Quality{
			Quality: models.QualityDefinition{
				ID:   6,
				Name: "Bluray-1080p",
			},
		},
		IndexerFlags: 0,
		Edition:      "Director's Cut",
		Languages: []models.Language{
			{
				ID:   1,
				Name: "English",
			},
		},
		MediaInfo: models.MediaInfo{
			AudioBitrate:                 1509,
			AudioChannels:                6,
			AudioCodec:                   "DTS",
			AudioLanguages:               "English",
			AudioStreamCount:             1,
			VideoBitDepth:                8,
			VideoBitrate:                 8000,
			VideoCodec:                   "AVC",
			VideoFps:                     23.976,
			Resolution:                   "1920x1080",
			RunTime:                      "2:00:00",
			ScanType:                     "Progressive",
			Subtitles:                    "English",
			VideoMultiViewCount:          1,
			VideoColourPrimaries:         "BT.709",
			VideoTransferCharacteristics: "BT.709",
			SchemaRevision:               5,
		},
	}

	// Apply any overrides
	for _, override := range overrides {
		override(movieFile)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(movieFile).Error; err != nil {
			panic(err)
		}
	}

	return movieFile
}

// CreateQualityProfile creates a test quality profile with default values
func (f *TestDataFactory) CreateQualityProfile(overrides ...func(*models.QualityProfile)) *models.QualityProfile {
	profile := &models.QualityProfile{
		Name:   "Test Quality Profile",
		Cutoff: 6, // Bluray-1080p quality ID
		Items: models.QualityProfileItems{
			{
				Quality: &models.QualityLevel{
					ID:    1,
					Title: "SDTV",
				},
				Allowed: false,
			},
			{
				Quality: &models.QualityLevel{
					ID:    2,
					Title: "DVD",
				},
				Allowed: true,
			},
			{
				Quality: &models.QualityLevel{
					ID:    4,
					Title: "HDTV-720p",
				},
				Allowed: true,
			},
			{
				Quality: &models.QualityLevel{
					ID:    5,
					Title: "WEBDL-720p",
				},
				Allowed: true,
			},
			{
				Quality: &models.QualityLevel{
					ID:    6,
					Title: "Bluray-1080p",
				},
				Allowed: true,
			},
		},
		MinFormatScore:    0,
		CutoffFormatScore: 0,
		FormatItems:       models.CustomFormatItems{},
		Language:          "english",
		UpgradeAllowed:    true,
	}

	// Apply any overrides
	for _, override := range overrides {
		override(profile)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(profile).Error; err != nil {
			panic(err)
		}
	}

	return profile
}

// CreateIndexer creates a test indexer with default values
func (f *TestDataFactory) CreateIndexer(overrides ...func(*models.Indexer)) *models.Indexer {
	indexer := &models.Indexer{
		Name:                    "Test Indexer",
		Type:                    models.IndexerTypeNewznab,
		BaseURL:                 "https://api.example.com",
		APIKey:                  "test-api-key",
		Categories:              "2000,2010,2020",
		EnableRSS:               true,
		EnableAutomaticSearch:   true,
		EnableInteractiveSearch: true,
		SupportsRSS:             true,
		SupportsSearch:          true,
		Priority:                25,
		Status:                  models.IndexerStatusEnabled,
		Settings: models.IndexerSettings{
			"baseUrl":    "https://api.example.com",
			"apiKey":     "test-api-key",
			"categories": []int{2000, 2010, 2020},
		},
	}

	// Apply any overrides
	for _, override := range overrides {
		override(indexer)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(indexer).Error; err != nil {
			panic(err)
		}
	}

	return indexer
}

// CreateDownloadClient creates a test download client with default values
func (f *TestDataFactory) CreateDownloadClient(overrides ...func(*models.DownloadClient)) *models.DownloadClient {
	client := &models.DownloadClient{
		Name:     "Test Download Client",
		Type:     "sabnzbd",
		Enable:   true,
		Priority: 1,
	}

	// Apply any overrides
	for _, override := range overrides {
		override(client)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(client).Error; err != nil {
			panic(err)
		}
	}

	return client
}

// CreateNotification creates a test notification with default values
func (f *TestDataFactory) CreateNotification(overrides ...func(*models.Notification)) *models.Notification {
	notification := &models.Notification{
		Name:                  "Test Notification",
		Implementation:        "Email",
		ConfigContract:        "EmailSettings",
		OnGrab:                true,
		OnDownload:            true,
		OnUpgrade:             true,
		OnRename:              false,
		OnMovieDelete:         true,
		OnMovieFileDelete:     true,
		OnHealthIssue:         true,
		IncludeHealthWarnings: false,
		Tags:                  []int{},
		Settings: map[string]interface{}{
			"server":   "smtp.example.com",
			"port":     587,
			"username": "test@example.com",
			"password": "password",
			"to":       []string{"user@example.com"},
			"subject":  "Radarr Notification",
		},
	}

	// Apply any overrides
	for _, override := range overrides {
		override(notification)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(notification).Error; err != nil {
			panic(err)
		}
	}

	return notification
}

// CreateTask creates a test task with default values
func (f *TestDataFactory) CreateTask(overrides ...func(*models.Task)) *models.Task {
	startedAt := time.Now().Add(-time.Hour)
	endedAt := time.Now().Add(-55 * time.Minute)
	duration := 5 * time.Minute

	task := &models.Task{
		Name:        "Test Task",
		CommandName: "test-command",
		Message:     "Test task message",
		Priority:    models.TaskPriorityNormal,
		Status:      models.TaskStatusCompleted,
		QueuedAt:    time.Now().Add(-time.Hour),
		StartedAt:   &startedAt,
		EndedAt:     &endedAt,
		Duration:    &duration,
		Trigger:     models.TaskTriggerManual,
	}

	// Apply any overrides
	for _, override := range overrides {
		override(task)
	}

	// Save to database if db is available
	if f.db != nil {
		if err := f.db.Create(task).Error; err != nil {
			panic(err)
		}
	}

	return task
}

// Cleanup removes all test data from the database
func (f *TestDataFactory) Cleanup() {
	if f.db == nil {
		return
	}

	// Clean up in reverse order of dependencies
	_ = f.db.Exec("DELETE FROM movie_files")
	_ = f.db.Exec("DELETE FROM movies")
	_ = f.db.Exec("DELETE FROM quality_profiles")
	_ = f.db.Exec("DELETE FROM indexers")
	_ = f.db.Exec("DELETE FROM download_clients")
	_ = f.db.Exec("DELETE FROM notifications")
	_ = f.db.Exec("DELETE FROM tasks")
	_ = f.db.Exec("DELETE FROM health_checks")
	_ = f.db.Exec("DELETE FROM health_issues")
	_ = f.db.Exec("DELETE FROM collections")
}
