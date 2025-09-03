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

	// Create all the basic components
	s.seedBasicQualityProfiles(dataset)
	s.seedBasicIndexers(dataset)
	s.seedBasicDownloadClient(dataset)
	s.seedBasicNotifications(dataset)
	s.seedBasicMovies(dataset)

	// Create movie file and update movie
	if err := s.seedBasicMovieFile(dataset); err != nil {
		return nil, err
	}

	// Create tasks
	s.seedBasicTasks(dataset)

	return dataset, nil
}

// seedBasicQualityProfiles creates quality profiles for basic dataset
func (s *SeedData) seedBasicQualityProfiles(dataset *BasicDataset) {
	dataset.QualityProfile = s.factory.CreateQualityProfile()
	dataset.QualityProfileHD = s.factory.CreateQualityProfile(func(qp *models.QualityProfile) {
		qp.Name = "HD Quality Profile"
		qp.Cutoff = 5 // WEBDL-720p
	})
}

// seedBasicIndexers creates indexers for basic dataset
func (s *SeedData) seedBasicIndexers(dataset *BasicDataset) {
	dataset.PrimaryIndexer = s.factory.CreateIndexer()
	dataset.SecondaryIndexer = s.factory.CreateIndexer(func(i *models.Indexer) {
		i.Name = "Secondary Indexer"
		i.BaseURL = "https://api2.example.com"
		i.Priority = 15
	})
}

// seedBasicDownloadClient creates download client for basic dataset
func (s *SeedData) seedBasicDownloadClient(dataset *BasicDataset) {
	dataset.DownloadClient = s.factory.CreateDownloadClient()
}

// seedBasicNotifications creates notifications for basic dataset
func (s *SeedData) seedBasicNotifications(dataset *BasicDataset) {
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
}

// seedBasicMovies creates movies with different statuses for basic dataset
func (s *SeedData) seedBasicMovies(dataset *BasicDataset) {
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
}

// seedBasicMovieFile creates movie file and links it to released movie
func (s *SeedData) seedBasicMovieFile(dataset *BasicDataset) error {
	dataset.MovieFile = s.factory.CreateMovieFile(dataset.ReleasedMovie.ID, func(mf *models.MovieFile) {
		mf.Path = fmt.Sprintf("/movies/%s/%s.mkv", dataset.ReleasedMovie.FolderName, dataset.ReleasedMovie.TitleSlug)
		mf.RelativePath = fmt.Sprintf("%s.mkv", dataset.ReleasedMovie.TitleSlug)
	})

	// Update released movie to reference the file
	dataset.ReleasedMovie.MovieFile = dataset.MovieFile
	dataset.ReleasedMovie.SizeOnDisk = dataset.MovieFile.Size
	if err := s.db.Save(dataset.ReleasedMovie).Error; err != nil {
		return fmt.Errorf("failed to update movie with file: %w", err)
	}

	return nil
}

// seedBasicTasks creates tasks for basic dataset
func (s *SeedData) seedBasicTasks(dataset *BasicDataset) {
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
}

// SeedPerformanceDataset creates a large dataset for performance testing
func (s *SeedData) SeedPerformanceDataset(movieCount int) (*PerformanceDataset, error) {
	dataset := &PerformanceDataset{
		Movies:         make([]*models.Movie, 0, movieCount),
		QualityProfile: s.factory.CreateQualityProfile(),
	}

	// Create indexers
	s.seedPerformanceIndexers(dataset)

	// Create movies in batches
	if err := s.seedPerformanceMovies(dataset, movieCount); err != nil {
		return nil, err
	}

	// Create movie files for movies that have files
	if err := s.seedPerformanceMovieFiles(dataset); err != nil {
		return nil, err
	}

	return dataset, nil
}

// seedPerformanceIndexers creates indexers for performance dataset
func (s *SeedData) seedPerformanceIndexers(dataset *PerformanceDataset) {
	for i := 0; i < 5; i++ {
		indexer := s.factory.CreateIndexer(func(idx *models.Indexer) {
			idx.Name = fmt.Sprintf("Performance Indexer %d", i+1)
			idx.BaseURL = fmt.Sprintf("https://api%d.example.com", i+1)
			idx.Priority = 25 - i*5
		})
		dataset.Indexers = append(dataset.Indexers, indexer)
	}
}

// seedPerformanceMovies creates movies in batches for performance dataset
func (s *SeedData) seedPerformanceMovies(dataset *PerformanceDataset, movieCount int) error {
	batchSize := 100
	for batch := 0; batch < movieCount; batch += batchSize {
		batchMovies := s.createPerformanceMovieBatch(dataset, batch, movieCount, batchSize)

		// Insert batch
		if err := s.db.CreateInBatches(batchMovies, batchSize).Error; err != nil {
			return fmt.Errorf("failed to create movie batch: %w", err)
		}

		dataset.Movies = append(dataset.Movies, batchMovies...)
	}
	return nil
}

// createPerformanceMovieBatch creates a batch of movies for performance testing
func (s *SeedData) createPerformanceMovieBatch(dataset *PerformanceDataset,
	batch, movieCount, batchSize int) []*models.Movie {
	endIdx := batch + batchSize
	if endIdx > movieCount {
		endIdx = movieCount
	}

	var batchMovies []*models.Movie
	for i := batch; i < endIdx; i++ {
		movie := s.createPerformanceMovie(i, dataset.QualityProfile.ID)
		batchMovies = append(batchMovies, movie)
	}

	return batchMovies
}

// createPerformanceMovie creates a single performance test movie
func (s *SeedData) createPerformanceMovie(index int, qualityProfileID int) *models.Movie {
	movie := &models.Movie{
		TmdbID:              200000 + index,
		Title:               fmt.Sprintf("Performance Movie %d", index+1),
		TitleSlug:           fmt.Sprintf("performance-movie-%d", index+1),
		Year:                2020 + (index % 4),
		Overview:            fmt.Sprintf("This is performance test movie number %d for testing scalability", index+1),
		Runtime:             90 + (index % 60), // 90-150 minutes
		Status:              models.MovieStatusReleased,
		Monitored:           index%2 == 0, // Every other movie monitored
		MinimumAvailability: models.AvailabilityReleased,
		IsAvailable:         true,
		HasFile:             index%3 == 0, // Every third movie has a file
		QualityProfileID:    qualityProfileID,
		Path:                fmt.Sprintf("/movies/performance-movie-%d", index+1),
		RootFolderPath:      "/movies",
		FolderName:          fmt.Sprintf("performance-movie-%d", index+1),
		SizeOnDisk:          0,
		Genres:              getRandomGenres(index),
		Tags:                models.IntArray{},
	}

	// Add some variety to release dates
	if index%5 != 0 { // 80% have release dates
		s.addReleaseDates(movie, index)
	}

	return movie
}

// addReleaseDates adds release date variety to performance movies
func (s *SeedData) addReleaseDates(movie *models.Movie, index int) {
	inCinemas := time.Date(2020+(index%4), time.Month((index%12)+1), (index%28)+1, 0, 0, 0, 0, time.UTC)
	movie.InCinemas = &inCinemas

	digitalRelease := inCinemas.Add(90 * 24 * time.Hour)
	movie.DigitalRelease = &digitalRelease

	physicalRelease := inCinemas.Add(120 * 24 * time.Hour)
	movie.PhysicalRelease = &physicalRelease
}

// seedPerformanceMovieFiles creates movie files for performance movies
func (s *SeedData) seedPerformanceMovieFiles(dataset *PerformanceDataset) error {
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
		batchSize := 100
		if err := s.db.CreateInBatches(movieFiles, batchSize).Error; err != nil {
			return fmt.Errorf("failed to create movie files: %w", err)
		}
		dataset.MovieFiles = movieFiles
	}

	return nil
}

// SeedStressTestDataset creates a comprehensive dataset for stress testing
func (s *SeedData) SeedStressTestDataset() (*StressTestDataset, error) {
	dataset := &StressTestDataset{}

	// Create all stress test components
	s.seedStressQualityProfiles(dataset)
	s.seedStressIndexers(dataset)
	s.seedStressDownloadClients(dataset)
	s.seedStressNotifications(dataset)
	s.seedStressTasks(dataset)

	return dataset, nil
}

// seedStressQualityProfiles creates multiple quality profiles for stress testing
func (s *SeedData) seedStressQualityProfiles(dataset *StressTestDataset) {
	for i := 0; i < 10; i++ {
		profile := s.factory.CreateQualityProfile(func(qp *models.QualityProfile) {
			qp.Name = fmt.Sprintf("Stress Quality Profile %d", i+1)
			qp.Cutoff = 4 + (i % 3) // Vary cutoff quality
		})
		dataset.QualityProfiles = append(dataset.QualityProfiles, profile)
	}
}

// seedStressIndexers creates many indexers for stress testing
func (s *SeedData) seedStressIndexers(dataset *StressTestDataset) {
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
}

// seedStressDownloadClients creates download clients for stress testing
func (s *SeedData) seedStressDownloadClients(dataset *StressTestDataset) {
	for i := 0; i < 10; i++ {
		client := s.factory.CreateDownloadClient(func(dc *models.DownloadClient) {
			dc.Name = fmt.Sprintf("Stress Download Client %d", i+1)
			dc.Priority = 10 - i
			dc.Enable = i%2 == 0
		})
		dataset.DownloadClients = append(dataset.DownloadClients, client)
	}
}

// seedStressNotifications creates various notifications for stress testing
func (s *SeedData) seedStressNotifications(dataset *StressTestDataset) {
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
}

// seedStressTasks creates many tasks with different statuses for stress testing
func (s *SeedData) seedStressTasks(dataset *StressTestDataset) {
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

			s.setTaskTimestamps(t, i)
		})
		dataset.Tasks = append(dataset.Tasks, task)
	}
}

// setTaskTimestamps sets appropriate timestamps for stress test tasks
func (s *SeedData) setTaskTimestamps(task *models.Task, index int) {
	baseTime := time.Now().Add(-time.Duration(index) * time.Hour)
	task.QueuedAt = baseTime

	if task.Status != models.TaskStatusQueued {
		startedAt := baseTime.Add(5 * time.Minute)
		task.StartedAt = &startedAt

		if task.Status == models.TaskStatusCompleted || task.Status == models.TaskStatusFailed {
			endedAt := startedAt.Add(time.Duration(10+index%50) * time.Minute)
			task.EndedAt = &endedAt
			duration := endedAt.Sub(startedAt)
			task.Duration = &duration
		}
	}
}

// getRandomGenres returns a random set of genres for variety
func getRandomGenres(seed int) models.StringArray {
	genres := []string{"Action", "Adventure", "Comedy", "Drama", "Horror",
		"Sci-Fi", "Thriller", "Romance", "Fantasy", "Mystery"}

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
