package testhelpers

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// SeedData provides methods for creating comprehensive test datasets
type SeedData struct {
	db      *gorm.DB
	factory *TestDataFactory
}

// NewSeedData creates a new seed data generator
func NewSeedData(db *gorm.DB) *SeedData {
	return &SeedData{
		db:      db,
		factory: NewTestDataFactory(db),
	}
}

// SeedBasicDataset creates a basic dataset for integration tests
func (s *SeedData) SeedBasicDataset() (*BasicDataset, error) {
	dataset := &BasicDataset{}

	// Create quality profiles
	dataset.QualityProfile = s.factory.CreateQualityProfile()
	dataset.QualityProfileHD = s.factory.CreateQualityProfile(func(qp *models.QualityProfile) {
		qp.Name = "HD Quality Profile"
		qp.Cutoff = 5 // WEBDL-720p
	})

	// Create indexers
	dataset.PrimaryIndexer = s.factory.CreateIndexer()
	dataset.SecondaryIndexer = s.factory.CreateIndexer(func(i *models.Indexer) {
		i.Name = "Secondary Indexer"
		i.BaseURL = "https://api2.example.com"
		i.Priority = 15
	})

	// Create download clients
	dataset.DownloadClient = s.factory.CreateDownloadClient()

	// Create notifications
	dataset.EmailNotification = s.factory.CreateNotification()
	dataset.SlackNotification = s.factory.CreateNotification(func(n *models.Notification) {
		n.Name = "Slack Notification"
		n.Implementation = "Slack"
		n.ConfigContract = "SlackSettings"
		n.Settings = map[string]interface{}{
			"webhook": "https://hooks.slack.com/services/test",
			"channel": "#radarr-test",
		}
	})

	// Create movies with different statuses
	dataset.ReleasedMovie = s.factory.CreateMovie(func(m *models.Movie) {
		m.Title = "Released Test Movie"
		m.TitleSlug = "released-test-movie"
		m.Status = models.MovieStatusReleased
		m.QualityProfileID = dataset.QualityProfile.ID
		m.HasFile = true
		m.TmdbID = 100001
	})

	dataset.AnnouncedMovie = s.factory.CreateMovie(func(m *models.Movie) {
		m.Title = "Announced Test Movie"
		m.TitleSlug = "announced-test-movie"
		m.Status = models.MovieStatusAnnounced
		m.QualityProfileID = dataset.QualityProfile.ID
		m.HasFile = false
		m.TmdbID = 100002
		futureDate := time.Now().Add(30 * 24 * time.Hour)
		m.InCinemas = &futureDate
	})

	dataset.MonitoredMovie = s.factory.CreateMovie(func(m *models.Movie) {
		m.Title = "Monitored Test Movie"
		m.TitleSlug = "monitored-test-movie"
		m.Status = models.MovieStatusReleased
		m.QualityProfileID = dataset.QualityProfile.ID
		m.Monitored = true
		m.HasFile = false
		m.TmdbID = 100003
	})

	dataset.UnmonitoredMovie = s.factory.CreateMovie(func(m *models.Movie) {
		m.Title = "Unmonitored Test Movie"
		m.TitleSlug = "unmonitored-test-movie"
		m.Status = models.MovieStatusReleased
		m.QualityProfileID = dataset.QualityProfile.ID
		m.Monitored = false
		m.HasFile = false
		m.TmdbID = 100004
	})

	// Create movie file for released movie
	dataset.MovieFile = s.factory.CreateMovieFile(dataset.ReleasedMovie.ID, func(mf *models.MovieFile) {
		mf.Path = fmt.Sprintf("/movies/%s/%s.mkv", dataset.ReleasedMovie.FolderName, dataset.ReleasedMovie.TitleSlug)
		mf.RelativePath = fmt.Sprintf("%s.mkv", dataset.ReleasedMovie.TitleSlug)
	})

	// Update released movie to reference the file
	dataset.ReleasedMovie.MovieFile = dataset.MovieFile
	dataset.ReleasedMovie.SizeOnDisk = dataset.MovieFile.Size
	if err := s.db.Save(dataset.ReleasedMovie).Error; err != nil {
		return nil, fmt.Errorf("failed to update movie with file: %w", err)
	}

	// Create tasks
	dataset.CompletedTask = s.factory.CreateTask(func(t *models.Task) {
		t.Name = "Refresh Movie Metadata"
		t.CommandName = "refresh-movie"
		t.Status = models.TaskStatusCompleted
		endedAt := time.Now().Add(-time.Minute)
		t.EndedAt = &endedAt
	})

	dataset.RunningTask = s.factory.CreateTask(func(t *models.Task) {
		t.Name = "Search for Movies"
		t.CommandName = "search-movies"
		t.Status = models.TaskStatusStarted
		startedAt := time.Now().Add(-10 * time.Minute)
		t.StartedAt = &startedAt
		t.EndedAt = nil
	})

	return dataset, nil
}

// SeedPerformanceDataset creates a large dataset for performance testing
func (s *SeedData) SeedPerformanceDataset(movieCount int) (*PerformanceDataset, error) {
	dataset := &PerformanceDataset{
		Movies:         make([]*models.Movie, 0, movieCount),
		QualityProfile: s.factory.CreateQualityProfile(),
	}

	// Create indexers
	for i := 0; i < 5; i++ {
		indexer := s.factory.CreateIndexer(func(idx *models.Indexer) {
			idx.Name = fmt.Sprintf("Performance Indexer %d", i+1)
			idx.BaseURL = fmt.Sprintf("https://api%d.example.com", i+1)
			idx.Priority = 25 - i*5
		})
		dataset.Indexers = append(dataset.Indexers, indexer)
	}

	// Create movies in batches to avoid memory issues
	batchSize := 100
	for batch := 0; batch < movieCount; batch += batchSize {
		var batchMovies []*models.Movie

		endIdx := batch + batchSize
		if endIdx > movieCount {
			endIdx = movieCount
		}

		for i := batch; i < endIdx; i++ {
			movie := &models.Movie{
				TmdbID:              200000 + i,
				Title:               fmt.Sprintf("Performance Movie %d", i+1),
				TitleSlug:           fmt.Sprintf("performance-movie-%d", i+1),
				Year:                2020 + (i % 4),
				Overview:            fmt.Sprintf("This is performance test movie number %d for testing scalability", i+1),
				Runtime:             90 + (i % 60), // 90-150 minutes
				Status:              models.MovieStatusReleased,
				Monitored:           i%2 == 0, // Every other movie monitored
				MinimumAvailability: models.AvailabilityReleased,
				IsAvailable:         true,
				HasFile:             i%3 == 0, // Every third movie has a file
				QualityProfileID:    dataset.QualityProfile.ID,
				Path:                fmt.Sprintf("/movies/performance-movie-%d", i+1),
				RootFolderPath:      "/movies",
				FolderName:          fmt.Sprintf("performance-movie-%d", i+1),
				SizeOnDisk:          0,
				Genres:              getRandomGenres(i),
				Tags:                models.IntArray{},
			}

			// Add some variety to release dates
			if i%5 != 0 { // 80% have release dates
				inCinemas := time.Date(2020+(i%4), time.Month((i%12)+1), (i%28)+1, 0, 0, 0, 0, time.UTC)
				movie.InCinemas = &inCinemas

				digitalRelease := inCinemas.Add(90 * 24 * time.Hour)
				movie.DigitalRelease = &digitalRelease

				physicalRelease := inCinemas.Add(120 * 24 * time.Hour)
				movie.PhysicalRelease = &physicalRelease
			}

			batchMovies = append(batchMovies, movie)
		}

		// Insert batch
		if err := s.db.CreateInBatches(batchMovies, batchSize).Error; err != nil {
			return nil, fmt.Errorf("failed to create movie batch: %w", err)
		}

		dataset.Movies = append(dataset.Movies, batchMovies...)
	}

	// Create movie files for movies that have files
	var movieFiles []*models.MovieFile
	for _, movie := range dataset.Movies {
		if movie.HasFile {
			movieFile := &models.MovieFile{
				MovieID:      movie.ID,
				RelativePath: fmt.Sprintf("%s.mkv", movie.TitleSlug),
				Path:         fmt.Sprintf("%s/%s.mkv", movie.Path, movie.TitleSlug),
				Size:         1024 * 1024 * (1000 + int64(movie.ID%2000)), // 1-3GB files
				DateAdded:    time.Now().Add(-time.Duration(movie.ID) * time.Hour),
				Quality: models.Quality{
					Quality: models.QualityDefinition{
						ID:   6,
						Name: "Bluray-1080p",
					},
				},
			}
			movieFiles = append(movieFiles, movieFile)
		}
	}

	// Create movie files in batches
	if len(movieFiles) > 0 {
		if err := s.db.CreateInBatches(movieFiles, batchSize).Error; err != nil {
			return nil, fmt.Errorf("failed to create movie files: %w", err)
		}
		dataset.MovieFiles = movieFiles
	}

	return dataset, nil
}

// SeedStressTestDataset creates a comprehensive dataset for stress testing
func (s *SeedData) SeedStressTestDataset() (*StressTestDataset, error) {
	dataset := &StressTestDataset{}

	// Create multiple quality profiles
	for i := 0; i < 10; i++ {
		profile := s.factory.CreateQualityProfile(func(qp *models.QualityProfile) {
			qp.Name = fmt.Sprintf("Stress Quality Profile %d", i+1)
			qp.Cutoff = 4 + (i % 3) // Vary cutoff quality
		})
		dataset.QualityProfiles = append(dataset.QualityProfiles, profile)
	}

	// Create many indexers
	for i := 0; i < 25; i++ {
		indexer := s.factory.CreateIndexer(func(idx *models.Indexer) {
			idx.Name = fmt.Sprintf("Stress Indexer %d", i+1)
			idx.BaseURL = fmt.Sprintf("https://stress%d.example.com", i+1)
			idx.Priority = 50 - i
			idx.EnableRSS = i%2 == 0
			idx.EnableAutomaticSearch = i%3 == 0
			idx.EnableInteractiveSearch = i%4 == 0
		})
		dataset.Indexers = append(dataset.Indexers, indexer)
	}

	// Create download clients
	for i := 0; i < 10; i++ {
		client := s.factory.CreateDownloadClient(func(dc *models.DownloadClient) {
			dc.Name = fmt.Sprintf("Stress Download Client %d", i+1)
			dc.Priority = 10 - i
			dc.Enable = i%2 == 0
		})
		dataset.DownloadClients = append(dataset.DownloadClients, client)
	}

	// Create notifications
	notificationTypes := []models.NotificationType{
		models.NotificationTypeEmail,
		models.NotificationTypeSlack,
		models.NotificationTypeDiscord,
		models.NotificationTypeWebhook,
		models.NotificationTypePushover,
	}
	for i := 0; i < 15; i++ {
		notification := s.factory.CreateNotification(func(n *models.Notification) {
			n.Name = fmt.Sprintf("Stress Notification %d", i+1)
			n.Implementation = notificationTypes[i%len(notificationTypes)]
			n.OnGrab = i%2 == 0
			n.OnDownload = i%3 == 0
			n.OnUpgrade = i%4 == 0
		})
		dataset.Notifications = append(dataset.Notifications, notification)
	}

	// Create many tasks with different statuses
	taskStatuses := []models.TaskStatus{
		models.TaskStatusQueued,
		models.TaskStatusStarted,
		models.TaskStatusCompleted,
		models.TaskStatusFailed,
		models.TaskStatusAborted,
	}

	for i := 0; i < 100; i++ {
		task := s.factory.CreateTask(func(t *models.Task) {
			t.Name = fmt.Sprintf("Stress Task %d", i+1)
			t.CommandName = fmt.Sprintf("stress-command-%d", i%10)
			t.Status = taskStatuses[i%len(taskStatuses)]

			baseTime := time.Now().Add(-time.Duration(i) * time.Hour)
			t.QueuedAt = baseTime

			if t.Status != models.TaskStatusQueued {
				startedAt := baseTime.Add(5 * time.Minute)
				t.StartedAt = &startedAt

				if t.Status == models.TaskStatusCompleted || t.Status == models.TaskStatusFailed {
					endedAt := startedAt.Add(time.Duration(10+i%50) * time.Minute)
					t.EndedAt = &endedAt
					duration := endedAt.Sub(startedAt)
					t.Duration = &duration
				}
			}
		})
		dataset.Tasks = append(dataset.Tasks, task)
	}

	return dataset, nil
}

// getRandomGenres returns a random set of genres for variety
func getRandomGenres(seed int) models.StringArray {
	genres := []string{"Action", "Adventure", "Comedy", "Drama", "Horror", "Sci-Fi", "Thriller", "Romance", "Fantasy", "Mystery"}

	// Use seed to determine which genres to include
	var selected []string
	for i, genre := range genres {
		if (seed+i)%3 == 0 { // Include every third genre based on seed
			selected = append(selected, genre)
		}
	}

	if len(selected) == 0 {
		selected = []string{"Drama"} // Fallback
	}

	return models.StringArray(selected)
}

// BasicDataset represents a basic set of test data
type BasicDataset struct {
	QualityProfile    *models.QualityProfile
	QualityProfileHD  *models.QualityProfile
	PrimaryIndexer    *models.Indexer
	SecondaryIndexer  *models.Indexer
	DownloadClient    *models.DownloadClient
	EmailNotification *models.Notification
	SlackNotification *models.Notification
	ReleasedMovie     *models.Movie
	AnnouncedMovie    *models.Movie
	MonitoredMovie    *models.Movie
	UnmonitoredMovie  *models.Movie
	MovieFile         *models.MovieFile
	CompletedTask     *models.Task
	RunningTask       *models.Task
}

// PerformanceDataset represents data for performance testing
type PerformanceDataset struct {
	Movies         []*models.Movie
	MovieFiles     []*models.MovieFile
	QualityProfile *models.QualityProfile
	Indexers       []*models.Indexer
}

// StressTestDataset represents data for stress testing
type StressTestDataset struct {
	QualityProfiles []*models.QualityProfile
	Indexers        []*models.Indexer
	DownloadClients []*models.DownloadClient
	Notifications   []*models.Notification
	Tasks           []*models.Task
}

// Cleanup removes all data created by the seed data generator
func (s *SeedData) Cleanup() {
	if s.factory != nil {
		s.factory.Cleanup()
	}
}
