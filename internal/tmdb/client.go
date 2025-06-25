// Package tmdb provides access to The Movie Database (TMDB) API for movie metadata retrieval.
package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
)

const (
	baseURL          = "https://api.themoviedb.org/3"
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "Radarr-Go/1.0"
	maxRetries       = 3
	retryDelay       = 1 * time.Second
)

// Client provides access to TMDB API
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	userAgent  string
	logger     *logger.Logger
}

// NewClient creates a new TMDB API client
func NewClient(cfg *config.Config, logger *logger.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		apiKey:    cfg.TMDB.APIKey,
		baseURL:   baseURL,
		userAgent: defaultUserAgent,
		logger:    logger,
	}
}

// Movie represents a TMDB movie
type Movie struct {
	ID                  int                 `json:"id"`
	IMDbID              string              `json:"imdb_id"`
	Title               string              `json:"title"`
	OriginalTitle       string              `json:"original_title"`
	OriginalLanguage    string              `json:"original_language"`
	Overview            string              `json:"overview"`
	ReleaseDate         string              `json:"release_date"`
	PosterPath          string              `json:"poster_path"`
	BackdropPath        string              `json:"backdrop_path"`
	Homepage            string              `json:"homepage"`
	Runtime             int                 `json:"runtime"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	VoteAverage         float64             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`
	Popularity          float64             `json:"popularity"`
	Adult               bool                `json:"adult"`
	Video               bool                `json:"video"`
	Genres              []Genre             `json:"genres"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	BelongsToCollection *Collection         `json:"belongs_to_collection"`
	Budget              int64               `json:"budget"`
	Revenue             int64               `json:"revenue"`
}

// Genre represents a movie genre
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ProductionCompany represents a production company
type ProductionCompany struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// ProductionCountry represents a production country
type ProductionCountry struct {
	ISO31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

// SpokenLanguage represents a spoken language
type SpokenLanguage struct {
	ISO6391     string `json:"iso_639_1"`
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
}

// Collection represents a movie collection
type Collection struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// SearchResponse represents a search response from TMDB
type SearchResponse struct {
	Page         int           `json:"page"`
	Results      []SearchMovie `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

// SearchMovie represents a movie in search results
type SearchMovie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIDs         []int   `json:"genre_ids"`
	Adult            bool    `json:"adult"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
}

// Credits represents movie credits
type Credits struct {
	ID   int          `json:"id"`
	Cast []CastMember `json:"cast"`
	Crew []CrewMember `json:"crew"`
}

// CastMember represents a cast member
type CastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	CreditID    string `json:"credit_id"`
	Order       int    `json:"order"`
	ProfilePath string `json:"profile_path"`
	Gender      int    `json:"gender"`
}

// CrewMember represents a crew member
type CrewMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	CreditID    string `json:"credit_id"`
	ProfilePath string `json:"profile_path"`
	Gender      int    `json:"gender"`
}

// GetMovie retrieves a movie by TMDB ID
func (c *Client) GetMovie(id int) (*Movie, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	endpoint := fmt.Sprintf("/movie/%d", id)
	params := url.Values{
		"api_key": {c.apiKey},
	}

	var movie Movie
	err := c.makeRequest(endpoint, params, &movie)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie %d: %w", id, err)
	}

	return &movie, nil
}

// SearchMovies searches for movies by query
func (c *Client) SearchMovies(query string, page int) (*SearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	endpoint := "/search/movie"
	params := url.Values{
		"api_key": {c.apiKey},
		"query":   {query},
	}

	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var response SearchResponse
	err := c.makeRequest(endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}

	return &response, nil
}

// GetCredits retrieves movie credits by TMDB ID
func (c *Client) GetCredits(id int) (*Credits, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	endpoint := fmt.Sprintf("/movie/%d/credits", id)
	params := url.Values{
		"api_key": {c.apiKey},
	}

	var credits Credits
	err := c.makeRequest(endpoint, params, &credits)
	if err != nil {
		return nil, fmt.Errorf("failed to get credits for movie %d: %w", id, err)
	}

	return &credits, nil
}

// GetPopular retrieves popular movies
func (c *Client) GetPopular(page int) (*SearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	endpoint := "/movie/popular"
	params := url.Values{
		"api_key": {c.apiKey},
	}

	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var response SearchResponse
	err := c.makeRequest(endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular movies: %w", err)
	}

	return &response, nil
}

// GetTrending retrieves trending movies
func (c *Client) GetTrending(timeWindow string, page int) (*SearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("TMDB API key not configured")
	}

	if timeWindow == "" {
		timeWindow = "week"
	}

	endpoint := fmt.Sprintf("/trending/movie/%s", timeWindow)
	params := url.Values{
		"api_key": {c.apiKey},
	}

	if page > 0 {
		params.Set("page", strconv.Itoa(page))
	}

	var response SearchResponse
	err := c.makeRequest(endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending movies: %w", err)
	}

	return &response, nil
}

// makeRequest makes an HTTP request to the TMDB API
func (c *Client) makeRequest(endpoint string, params url.Values, result interface{}) error {
	reqURL := c.baseURL + endpoint + "?" + params.Encode()

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay * time.Duration(attempt))
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			c.logger.Warn("TMDB API request failed, retrying", "attempt", attempt+1, "error", err)
			continue
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.logger.Warn("Failed to close response body", "error", err)
			}
		}()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited by TMDB API")
			c.logger.Warn("TMDB API rate limited, retrying", "attempt", attempt+1)
			time.Sleep(retryDelay * 2)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API request failed with status %d", resp.StatusCode)
			c.logger.Error("TMDB API request failed", "status", resp.StatusCode, "url", reqURL)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %w", err)
			continue
		}

		return nil
	}

	return lastErr
}
