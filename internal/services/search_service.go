package services

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

const (
	// defaultSortOrder is the default sort order for search results
	defaultSortOrder = "desc"
)

// NewznabResponse represents a Newznab/Torznab XML response structure
type NewznabResponse struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Items []NewznabItem `xml:"item"`
	} `xml:"channel"`
}

// NewznabItem represents a single item in a Newznab response
type NewznabItem struct {
	Title       string             `xml:"title"`
	Link        string             `xml:"link"`
	GUID        string             `xml:"guid"`
	PubDate     string             `xml:"pubDate"`
	Description string             `xml:"description"`
	Enclosure   NewznabEnclosure   `xml:"enclosure"`
	Attributes  []NewznabAttribute `xml:"attr"`
}

// NewznabEnclosure represents the enclosure element
type NewznabEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

// NewznabAttribute represents a Newznab attribute
type NewznabAttribute struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// SearchService handles movie release searches and management
type SearchService struct {
	db                  *database.Database
	logger              *logger.Logger
	indexerService      *IndexerService
	qualityService      *QualityService
	movieService        *MovieService
	downloadService     *DownloadService
	notificationService *NotificationService
	httpClient          *http.Client
}

// NewSearchService creates a new search service
func NewSearchService(
	db *database.Database,
	logger *logger.Logger,
	indexerService *IndexerService,
	qualityService *QualityService,
	movieService *MovieService,
	downloadService *DownloadService,
	notificationService *NotificationService,
) *SearchService {
	return &SearchService{
		db:                  db,
		logger:              logger,
		indexerService:      indexerService,
		qualityService:      qualityService,
		movieService:        movieService,
		downloadService:     downloadService,
		notificationService: notificationService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchMovieReleases searches for releases for a specific movie
func (s *SearchService) SearchMovieReleases(movieID int, forceSearch bool) (*models.SearchResponse, error) {
	movie, err := s.movieService.GetByID(movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	searchRequest := &models.SearchRequest{
		MovieID: &movieID,
		Title:   movie.Title,
		Year:    &movie.Year,
		ImdbID:  movie.ImdbID,
		TmdbID:  &movie.TmdbID,
		Source:  models.ReleaseSourceSearch,
		Limit:   100,
	}

	return s.SearchReleases(searchRequest, forceSearch)
}

// SearchReleases performs a search across enabled indexers
func (s *SearchService) SearchReleases(request *models.SearchRequest, forceSearch bool) (
	*models.SearchResponse, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	indexers, err := s.indexerService.GetEnabledIndexers()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled indexers: %w", err)
	}

	if len(indexers) == 0 {
		s.logger.Warn("No enabled indexers found for search")
		return &models.SearchResponse{Releases: []models.Release{}, Total: 0}, nil
	}

	allReleases, totalSearchTime := s.searchAllIndexers(indexers, request, forceSearch)
	allReleases = s.processSearchResults(allReleases, request)

	return &models.SearchResponse{
		Releases:   allReleases,
		Total:      len(allReleases),
		Limit:      request.Limit,
		Offset:     request.Offset,
		SearchTime: totalSearchTime,
	}, nil
}

// searchAllIndexers searches across all enabled indexers
func (s *SearchService) searchAllIndexers(indexers []*models.Indexer, request *models.SearchRequest,
	forceSearch bool) ([]models.Release, float64) {
	allReleases := make([]models.Release, 0)
	totalSearchTime := 0.0

	for _, indexer := range indexers {
		if !s.shouldSearchIndexer(indexer, request) {
			continue
		}

		releases, searchTime := s.performIndexerSearch(indexer, request, forceSearch)
		totalSearchTime += searchTime
		allReleases = append(allReleases, releases...)
	}

	return allReleases, totalSearchTime
}

// shouldSearchIndexer determines if an indexer should be searched
func (s *SearchService) shouldSearchIndexer(indexer *models.Indexer, request *models.SearchRequest) bool {
	if !indexer.CanSearch() {
		return false
	}

	if len(request.IndexerIDs) > 0 {
		for _, id := range request.IndexerIDs {
			if id == indexer.ID {
				return true
			}
		}
		return false
	}

	return true
}

// performIndexerSearch performs search on a single indexer
func (s *SearchService) performIndexerSearch(indexer *models.Indexer, request *models.SearchRequest,
	forceSearch bool) ([]models.Release, float64) {
	startTime := time.Now()
	releases, err := s.searchIndexer(indexer, request)
	searchTime := time.Since(startTime).Seconds()

	if err != nil {
		s.logger.Error("Failed to search indexer", "indexer", indexer.Name, "error", err, "searchTime", searchTime)
		return []models.Release{}, searchTime
	}

	for i := range releases {
		releases[i].IndexerID = indexer.ID
		releases[i].Source = request.Source
		releases[i] = s.processRelease(releases[i])
	}

	if !forceSearch {
		if err := s.saveReleases(releases); err != nil {
			s.logger.Error("Failed to save releases", "error", err)
		}
	}

	s.logger.Info("Search completed", "indexer", indexer.Name, "releases", len(releases), "searchTime", searchTime)
	return releases, searchTime
}

// processSearchResults processes and filters search results
func (s *SearchService) processSearchResults(releases []models.Release,
	request *models.SearchRequest) []models.Release {
	releases = s.dedupReleases(releases)
	releases = s.applyFilters(releases, request)
	releases = s.sortReleases(releases, request)

	if request.Limit > 0 && len(releases) > request.Limit {
		if request.Offset >= len(releases) {
			return []models.Release{}
		}
		end := request.Offset + request.Limit
		if end > len(releases) {
			end = len(releases)
		}
		return releases[request.Offset:end]
	}

	return releases
}

// InteractiveSearch performs an interactive search for manual release selection
func (s *SearchService) InteractiveSearch(request *models.SearchRequest) (*models.SearchResponse, error) {
	request.Source = models.ReleaseSourceInteractiveSearch
	response, err := s.SearchReleases(request, true)
	if err != nil {
		return nil, err
	}

	for i := range response.Releases {
		response.Releases[i] = s.evaluateRelease(response.Releases[i])
	}

	return response, nil
}

// GrabRelease grabs a release and sends it to the appropriate download client
func (s *SearchService) GrabRelease(request *models.GrabRequest) (*models.GrabResponse, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	release, err := s.findReleaseForGrab(request)
	if err != nil {
		return nil, err
	}

	if !release.IsGrabbable() {
		return s.createRejectedResponse(release), nil
	}

	downloadClient, err := s.getDownloadClientForRelease(request, release)
	if err != nil {
		return nil, err
	}

	if err := s.markReleaseAsGrabbed(release, downloadClient); err != nil {
		s.logger.Error("Failed to update release status", "error", err)
	}

	s.logGrabSuccess(release, downloadClient)
	return s.createSuccessResponse(release, downloadClient), nil
}

// findReleaseForGrab finds and loads the release for grabbing
func (s *SearchService) findReleaseForGrab(request *models.GrabRequest) (*models.Release, error) {
	var release models.Release
	if err := s.db.GORM.Where("guid = ? AND indexer_id = ?", request.GUID, request.IndexerID).
		Preload("Indexer").Preload("Movie").First(&release).Error; err != nil {
		return nil, fmt.Errorf("release not found: %w", err)
	}
	return &release, nil
}

// createRejectedResponse creates a response for a rejected release
func (s *SearchService) createRejectedResponse(release *models.Release) *models.GrabResponse {
	return &models.GrabResponse{
		ID:      release.ID,
		GUID:    release.GUID,
		Title:   release.Title,
		Status:  "rejected",
		Message: fmt.Sprintf("Release not grabbable: %s", release.GetRejectionString()),
	}
}

// getDownloadClientForRelease gets the appropriate download client for a release
func (s *SearchService) getDownloadClientForRelease(
	request *models.GrabRequest, release *models.Release) (*models.DownloadClient, error) {
	downloadClientID := request.DownloadClientID
	if downloadClientID == nil && release.Indexer != nil && release.Indexer.DownloadClientID != nil {
		downloadClientID = release.Indexer.DownloadClientID
	}

	return s.downloadService.GetDownloadClientByID(*downloadClientID)
}

// markReleaseAsGrabbed updates the release status to grabbed
func (s *SearchService) markReleaseAsGrabbed(release *models.Release, downloadClient *models.DownloadClient) error {
	now := time.Now()
	release.Status = models.ReleaseStatusGrabbed
	release.GrabbedAt = &now
	release.DownloadClientID = &downloadClient.ID
	return s.db.GORM.Save(release).Error
}

// logGrabSuccess logs successful grab information
func (s *SearchService) logGrabSuccess(release *models.Release, downloadClient *models.DownloadClient) {
	if release.Movie != nil {
		s.logger.Info("Movie grabbed", "movie", release.Movie.Title, "release", release.Title)
	}
	s.logger.Info("Release grabbed successfully",
		"release", release.Title,
		"indexer", release.Indexer.Name,
		"downloadClient", downloadClient.Name)
}

// createSuccessResponse creates a successful grab response
func (s *SearchService) createSuccessResponse(
	release *models.Release, downloadClient *models.DownloadClient) *models.GrabResponse {
	return &models.GrabResponse{
		ID:               release.ID,
		GUID:             release.GUID,
		Title:            release.Title,
		Status:           "grabbed",
		DownloadClientID: &downloadClient.ID,
		Message:          fmt.Sprintf("Successfully sent to %s", downloadClient.Name),
	}
}

// GetReleases retrieves releases with optional filtering
func (s *SearchService) GetReleases(filter *models.ReleaseFilter, limit, offset int) ([]models.Release, int, error) {
	if s.db == nil {
		return nil, 0, fmt.Errorf("database not available")
	}

	query := s.db.GORM.Model(&models.Release{}).
		Preload("Indexer").
		Preload("Movie").
		Preload("DownloadClient")

	if filter != nil {
		query = s.applyReleaseFilter(query, filter)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count releases: %w", err)
	}

	var releases []models.Release
	if err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&releases).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get releases: %w", err)
	}

	return releases, int(total), nil
}

// GetReleaseByID retrieves a release by ID
func (s *SearchService) GetReleaseByID(id int) (*models.Release, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var release models.Release
	if err := s.db.GORM.Preload("Indexer").Preload("Movie").Preload("DownloadClient").
		First(&release, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	return &release, nil
}

// DeleteRelease deletes a release
func (s *SearchService) DeleteRelease(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	if err := s.db.GORM.Delete(&models.Release{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete release: %w", err)
	}

	return nil
}

// GetReleaseStats returns statistics about releases
func (s *SearchService) GetReleaseStats() (*models.ReleaseStats, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := &models.ReleaseStats{
		ProtocolBreakdown: make(map[models.Protocol]int),
		SourceBreakdown:   make(map[models.ReleaseSource]int),
		IndexerBreakdown:  make(map[int]int),
		LastUpdated:       time.Now(),
	}

	if err := s.populateTotalReleases(stats); err != nil {
		return nil, err
	}

	if err := s.populateStatusCounts(stats); err != nil {
		return nil, err
	}

	if err := s.populateSizeStats(stats); err != nil {
		return nil, err
	}

	if err := s.populateAgeStats(stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// populateTotalReleases gets the total release count
func (s *SearchService) populateTotalReleases(stats *models.ReleaseStats) error {
	var totalReleases int64
	if err := s.db.GORM.Model(&models.Release{}).Count(&totalReleases).Error; err != nil {
		return fmt.Errorf("failed to count total releases: %w", err)
	}
	stats.TotalReleases = int(totalReleases)
	return nil
}

// populateStatusCounts gets release counts by status
func (s *SearchService) populateStatusCounts(stats *models.ReleaseStats) error {
	statusCounts := []struct {
		Status models.ReleaseStatus
		Count  int64
	}{}
	if err := s.db.GORM.Model(&models.Release{}).
		Select("status, count(*) as count").
		Group("status").Scan(&statusCounts).Error; err != nil {
		return fmt.Errorf("failed to get status counts: %w", err)
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case models.ReleaseStatusAvailable:
			stats.AvailableReleases = int(sc.Count)
		case models.ReleaseStatusGrabbed:
			stats.GrabbedReleases = int(sc.Count)
		case models.ReleaseStatusRejected:
			stats.RejectedReleases = int(sc.Count)
		case models.ReleaseStatusFailed:
			stats.FailedReleases = int(sc.Count)
		}
	}
	return nil
}

// populateSizeStats gets size statistics
func (s *SearchService) populateSizeStats(stats *models.ReleaseStats) error {
	var sizeStats struct {
		AverageSize float64
		TotalSize   int64
	}
	if err := s.db.GORM.Model(&models.Release{}).
		Select("AVG(size) as average_size, SUM(size) as total_size").
		Scan(&sizeStats).Error; err != nil {
		return fmt.Errorf("failed to get size stats: %w", err)
	}
	stats.AverageSize = sizeStats.AverageSize
	stats.TotalSize = sizeStats.TotalSize
	return nil
}

// populateAgeStats gets age statistics
func (s *SearchService) populateAgeStats(stats *models.ReleaseStats) error {
	var ageStats struct {
		AverageAge float64
	}
	if err := s.db.GORM.Model(&models.Release{}).
		Select("AVG(age) as average_age").
		Scan(&ageStats).Error; err != nil {
		return fmt.Errorf("failed to get age stats: %w", err)
	}
	stats.AverageAge = ageStats.AverageAge
	return nil
}

// CleanupOldReleases removes old releases based on retention settings
func (s *SearchService) CleanupOldReleases(retentionDays int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result := s.db.GORM.Where("created_at < ? AND status NOT IN ?", cutoffDate,
		[]models.ReleaseStatus{models.ReleaseStatusGrabbed}).
		Delete(&models.Release{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old releases: %w", result.Error)
	}

	s.logger.Info("Cleaned up old releases", "deleted", result.RowsAffected, "cutoffDate", cutoffDate)
	return nil
}

// searchIndexer searches a specific indexer
func (s *SearchService) searchIndexer(indexer *models.Indexer, request *models.SearchRequest) (
	[]models.Release, error) {
	switch indexer.Type {
	case models.IndexerTypeTorznab, models.IndexerTypeNewznab:
		return s.searchNewznabIndexer(indexer, request)
	case models.IndexerTypeRSS:
		return s.searchRSSIndexer(indexer, request)
	default:
		return nil, fmt.Errorf("unsupported indexer type: %s", indexer.Type)
	}
}

// searchNewznabIndexer searches a Newznab/Torznab indexer
func (s *SearchService) searchNewznabIndexer(indexer *models.Indexer, request *models.SearchRequest) (
	[]models.Release, error) {
	searchURL, err := s.buildNewznabURL(indexer, request)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return s.parseNewznabResponse(body, indexer)
}

// searchRSSIndexer searches an RSS indexer
func (s *SearchService) searchRSSIndexer(indexer *models.Indexer, request *models.SearchRequest) (
	[]models.Release, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", indexer.BaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform RSS request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return s.parseRSSResponse(body, indexer, request)
}

// buildNewznabURL builds a Newznab/Torznab search URL
func (s *SearchService) buildNewznabURL(indexer *models.Indexer, request *models.SearchRequest) (string, error) {
	baseURL, err := url.Parse(indexer.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("t", "movie")
	params.Set("extended", "1")

	if indexer.APIKey != "" {
		params.Set("apikey", indexer.APIKey)
	}

	if request.ImdbID != "" {
		params.Set("imdbid", strings.TrimPrefix(request.ImdbID, "tt"))
	}

	if request.TmdbID != nil && *request.TmdbID > 0 {
		params.Set("tmdbid", strconv.Itoa(*request.TmdbID))
	}

	if request.Title != "" {
		params.Set("q", request.Title)
	}

	if request.Year != nil && *request.Year > 0 {
		params.Set("year", strconv.Itoa(*request.Year))
	}

	if len(request.Categories) > 0 {
		catStr := make([]string, len(request.Categories))
		for i, cat := range request.Categories {
			catStr[i] = strconv.Itoa(cat)
		}
		params.Set("cat", strings.Join(catStr, ","))
	} else if indexer.Categories != "" {
		params.Set("cat", indexer.Categories)
	}

	if request.Limit > 0 {
		params.Set("limit", strconv.Itoa(request.Limit))
	}

	if request.Offset > 0 {
		params.Set("offset", strconv.Itoa(request.Offset))
	}

	baseURL.RawQuery = params.Encode()
	return baseURL.String(), nil
}

// parseNewznabResponse parses a Newznab/Torznab XML response
func (s *SearchService) parseNewznabResponse(data []byte, _ *models.Indexer) ([]models.Release, error) {
	response, err := s.unmarshalNewznabXML(data)
	if err != nil {
		return nil, err
	}

	releases := make([]models.Release, 0, len(response.Channel.Items))
	for _, item := range response.Channel.Items {
		release := s.createReleaseFromNewznabItem(item)
		releases = append(releases, release)
	}

	return releases, nil
}

// unmarshalNewznabXML unmarshals Newznab XML response
func (s *SearchService) unmarshalNewznabXML(data []byte) (*NewznabResponse, error) {
	var response NewznabResponse
	if err := xml.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}
	return &response, nil
}

// createReleaseFromNewznabItem creates a release from a Newznab item
func (s *SearchService) createReleaseFromNewznabItem(item NewznabItem) models.Release {
	release := models.Release{
		GUID:        item.GUID,
		Title:       item.Title,
		SortTitle:   strings.ToLower(item.Title),
		Overview:    item.Description,
		DownloadURL: item.Link,
		InfoURL:     item.Link,
		Status:      models.ReleaseStatusAvailable,
	}

	s.processEnclosure(&release, item.Enclosure)
	s.processPublishDate(&release, item.PubDate)
	s.detectProtocol(&release, item)
	s.processAttributes(&release, item.Attributes)

	return release
}

// processEnclosure processes the enclosure element
func (s *SearchService) processEnclosure(release *models.Release, enclosure NewznabEnclosure) {
	if enclosure.URL != "" {
		release.DownloadURL = enclosure.URL
		if enclosure.Length != "" {
			if size, err := strconv.ParseInt(enclosure.Length, 10, 64); err == nil {
				release.Size = size
			}
		}
	}
}

// processPublishDate processes the publish date
func (s *SearchService) processPublishDate(release *models.Release, pubDateStr string) {
	pubDate, err := time.Parse(time.RFC1123Z, pubDateStr)
	if err != nil {
		pubDate, err = time.Parse(time.RFC1123, pubDateStr)
	}
	if err == nil {
		release.PublishDate = pubDate
		release.Age = int(time.Since(pubDate).Hours() / 24)
		release.AgeHours = time.Since(pubDate).Hours()
		release.AgeMinutes = time.Since(pubDate).Minutes()
	}
}

// detectProtocol detects the release protocol
func (s *SearchService) detectProtocol(release *models.Release, item NewznabItem) {
	if strings.Contains(strings.ToLower(item.Enclosure.Type), "torrent") ||
		strings.Contains(strings.ToLower(item.Link), "torrent") {
		release.Protocol = models.ProtocolTorrent
	} else {
		release.Protocol = models.ProtocolUsenet
	}
}

// processAttributes processes Newznab attributes
func (s *SearchService) processAttributes(release *models.Release, attributes []NewznabAttribute) {
	for _, attr := range attributes {
		switch attr.Name {
		case "category":
			if catID, err := strconv.Atoi(attr.Value); err == nil {
				release.Categories = append(release.Categories, catID)
			}
		case "size":
			if size, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
				release.Size = size
			}
		case "seeders":
			if seeders, err := strconv.Atoi(attr.Value); err == nil {
				release.Seeders = &seeders
			}
		case "peers":
			if peers, err := strconv.Atoi(attr.Value); err == nil {
				release.PeerCount = peers
			}
		case "leechers":
			if leechers, err := strconv.Atoi(attr.Value); err == nil {
				release.Leechers = &leechers
			}
		case "imdbid":
			release.ImdbID = attr.Value
		case "tmdbid":
			if tmdbID, err := strconv.Atoi(attr.Value); err == nil {
				release.TmdbID = &tmdbID
			}
		case "magneturl":
			release.MagnetURL = attr.Value
		}
	}
}

// parseRSSResponse parses an RSS response
func (s *SearchService) parseRSSResponse(data []byte, _ *models.Indexer,
	request *models.SearchRequest) ([]models.Release, error) {
	type RSSResponse struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []struct {
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				GUID        string `xml:"guid"`
				PubDate     string `xml:"pubDate"`
				Description string `xml:"description"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	var response RSSResponse
	if err := xml.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse RSS response: %w", err)
	}

	releases := make([]models.Release, 0)

	for _, item := range response.Channel.Items {
		if request.Title != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(request.Title)) {
			continue
		}

		release := models.Release{
			GUID:        item.GUID,
			Title:       item.Title,
			SortTitle:   strings.ToLower(item.Title),
			Overview:    item.Description,
			DownloadURL: item.Link,
			InfoURL:     item.Link,
			Protocol:    models.ProtocolTorrent,
			Status:      models.ReleaseStatusAvailable,
		}

		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubDate, err = time.Parse(time.RFC1123, item.PubDate)
		}
		if err == nil {
			release.PublishDate = pubDate
			release.Age = int(time.Since(pubDate).Hours() / 24)
			release.AgeHours = time.Since(pubDate).Hours()
			release.AgeMinutes = time.Since(pubDate).Minutes()
		}

		releases = append(releases, release)
	}

	return releases, nil
}

// processRelease processes a release to extract metadata and quality information
func (s *SearchService) processRelease(release models.Release) models.Release {
	release.Quality = s.parseQualityFromTitle(release.Title)
	release.QualityWeight = s.calculateQualityWeight(release.Quality)
	release.ReleaseInfo = s.extractReleaseInfo(release.Title)

	return release
}

// parseQualityFromTitle extracts quality information from release title
func (s *SearchService) parseQualityFromTitle(title string) models.Quality {
	titleLower := strings.ToLower(title)

	quality := models.Quality{
		Quality: models.QualityDefinition{
			ID:         1,
			Name:       "Unknown",
			Source:     "unknown",
			Resolution: 0,
		},
		Revision: models.Revision{
			Version: 1,
			Real:    0,
		},
	}

	resolutionPatterns := map[string]int{
		"2160p": 2160,
		"1080p": 1080,
		"720p":  720,
		"480p":  480,
	}

	for pattern, res := range resolutionPatterns {
		if strings.Contains(titleLower, pattern) {
			quality.Quality.Resolution = res
			quality.Quality.Name = pattern
			break
		}
	}

	sourcePatterns := map[string]string{
		"bluray":   "bluray",
		"web-dl":   "webdl",
		"webrip":   "webrip",
		"hdtv":     "hdtv",
		"dvdrip":   "dvd",
		"cam":      "cam",
		"telesync": "telesync",
	}

	for pattern, source := range sourcePatterns {
		if strings.Contains(titleLower, pattern) {
			quality.Quality.Source = source
			break
		}
	}

	return quality
}

// calculateQualityWeight calculates a weight for quality comparison
func (s *SearchService) calculateQualityWeight(quality models.Quality) int {
	weight := quality.Quality.Resolution

	sourceWeights := map[string]int{
		"bluray":   1000,
		"webdl":    800,
		"webrip":   600,
		"hdtv":     400,
		"dvd":      200,
		"cam":      50,
		"telesync": 25,
	}

	if sourceWeight, exists := sourceWeights[quality.Quality.Source]; exists {
		weight += sourceWeight
	}

	return weight
}

// extractReleaseInfo extracts detailed release information from title
func (s *SearchService) extractReleaseInfo(title string) models.ReleaseInfo {
	titleLower := strings.ToLower(title)

	info := models.ReleaseInfo{
		Title:                title,
		DownloadVolumeFactor: 1.0,
		UploadVolumeFactor:   1.0,
	}

	if strings.Contains(titleLower, "freeleech") || strings.Contains(titleLower, "fl") {
		info.Freeleech = true
		info.DownloadVolumeFactor = 0.0
	}

	codecPatterns := []string{"x264", "x265", "hevc", "h264", "h265", "xvid", "divx"}
	for _, codec := range codecPatterns {
		if strings.Contains(titleLower, codec) {
			info.Codec = codec
			break
		}
	}

	containerPatterns := []string{"mkv", "mp4", "avi", "mov"}
	for _, container := range containerPatterns {
		if strings.Contains(titleLower, "."+container) {
			info.Container = container
			break
		}
	}

	if strings.Contains(titleLower, "scene") {
		info.Scene = true
	}

	return info
}

// evaluateRelease evaluates a release and adds rejection reasons if applicable
func (s *SearchService) evaluateRelease(release models.Release) models.Release {
	var rejections []string

	if release.Size < 100*1024*1024 {
		rejections = append(rejections, "File too small")
	}

	if release.Size > 50*1024*1024*1024 {
		rejections = append(rejections, "File too large")
	}

	if release.IsTorrent() && release.Seeders != nil && *release.Seeders == 0 {
		rejections = append(rejections, "No seeders")
	}

	if release.Age > 365 {
		rejections = append(rejections, "Too old")
	}

	release.RejectionReasons = rejections
	if len(rejections) > 0 {
		release.Status = models.ReleaseStatusRejected
	}

	return release
}

// saveReleases saves releases to the database
func (s *SearchService) saveReleases(releases []models.Release) error {
	if len(releases) == 0 {
		return nil
	}

	tx := s.db.GORM.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, release := range releases {
		var existing models.Release
		if err := tx.Where("guid = ? AND indexer_id = ?", release.GUID, release.IndexerID).
			First(&existing).Error; err != nil {
			if err := tx.Create(&release).Error; err != nil {
				s.logger.Error("Failed to create release", "error", err, "title", release.Title)
				continue
			}
		} else {
			release.ID = existing.ID
			if err := tx.Save(&release).Error; err != nil {
				s.logger.Error("Failed to update release", "error", err, "title", release.Title)
				continue
			}
		}
	}

	return tx.Commit().Error
}

// dedupReleases removes duplicate releases
func (s *SearchService) dedupReleases(releases []models.Release) []models.Release {
	seen := make(map[string]bool)
	result := make([]models.Release, 0, len(releases))

	for _, release := range releases {
		key := fmt.Sprintf("%s-%d", release.GUID, release.IndexerID)
		if !seen[key] {
			seen[key] = true
			result = append(result, release)
		}
	}

	return result
}

// applyFilters applies filters to releases
func (s *SearchService) applyFilters(releases []models.Release, request *models.SearchRequest) []models.Release {
	if request.Protocol != nil {
		filtered := make([]models.Release, 0, len(releases))
		for _, release := range releases {
			if release.Protocol == *request.Protocol {
				filtered = append(filtered, release)
			}
		}
		releases = filtered
	}

	return releases
}

// sortReleases sorts releases based on request parameters
func (s *SearchService) sortReleases(releases []models.Release, request *models.SearchRequest) []models.Release {
	sortBy := request.SortBy
	if sortBy == "" {
		sortBy = "qualityWeight"
	}

	sortOrder := request.SortOrder
	if sortOrder == "" {
		sortOrder = defaultSortOrder
	}

	sort.Slice(releases, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			less = releases[i].Title < releases[j].Title
		case "size":
			less = releases[i].Size < releases[j].Size
		case "age":
			less = releases[i].Age < releases[j].Age
		case "seeders":
			seedersI := 0
			if releases[i].Seeders != nil {
				seedersI = *releases[i].Seeders
			}
			seedersJ := 0
			if releases[j].Seeders != nil {
				seedersJ = *releases[j].Seeders
			}
			less = seedersI < seedersJ
		case "publishDate":
			less = releases[i].PublishDate.Before(releases[j].PublishDate)
		default:
			less = releases[i].QualityWeight < releases[j].QualityWeight
		}

		if sortOrder == defaultSortOrder {
			return !less
		}
		return less
	})

	return releases
}

// applyReleaseFilter applies database filters to release query
func (s *SearchService) applyReleaseFilter(query *gorm.DB, filter *models.ReleaseFilter) *gorm.DB {
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if len(filter.Source) > 0 {
		query = query.Where("source IN ?", filter.Source)
	}

	if len(filter.Protocol) > 0 {
		query = query.Where("protocol IN ?", filter.Protocol)
	}

	if len(filter.IndexerIDs) > 0 {
		query = query.Where("indexer_id IN ?", filter.IndexerIDs)
	}

	if len(filter.MovieIDs) > 0 {
		query = query.Where("movie_id IN ?", filter.MovieIDs)
	}

	if filter.MinSize != nil {
		query = query.Where("size >= ?", *filter.MinSize)
	}

	if filter.MaxSize != nil {
		query = query.Where("size <= ?", *filter.MaxSize)
	}

	if filter.MinAge != nil {
		query = query.Where("age >= ?", *filter.MinAge)
	}

	if filter.MaxAge != nil {
		query = query.Where("age <= ?", *filter.MaxAge)
	}

	if filter.HasSeeders != nil && *filter.HasSeeders {
		query = query.Where("seeders > 0")
	}

	if filter.MinSeeders != nil {
		query = query.Where("seeders >= ?", *filter.MinSeeders)
	}

	if filter.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filter.CreatedAfter)
	}

	if filter.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filter.CreatedBefore)
	}

	return query
}
