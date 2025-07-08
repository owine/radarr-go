package services

import (
	"os"
	"testing"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestImportListService_GetImportLists(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.GetImportLists()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_GetImportListByID(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.GetImportListByID(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_CreateImportList(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TMDB-dependent test in CI")
	}
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	list := &models.ImportList{
		Name:             "Test List",
		Implementation:   models.ImportListTypeTMDBPopular,
		QualityProfileID: 1,
		RootFolderPath:   "/movies",
	}

	// Test with nil database
	err := service.CreateImportList(list)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_UpdateImportList(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TMDB-dependent test in CI")
	}
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	list := &models.ImportList{
		ID:               1,
		Name:             "Updated List",
		Implementation:   models.ImportListTypeTMDBPopular,
		QualityProfileID: 1,
		RootFolderPath:   "/movies",
	}

	// Test with nil database
	err := service.UpdateImportList(list)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_DeleteImportList(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	err := service.DeleteImportList(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_GetEnabledImportLists(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.GetEnabledImportLists()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_TestImportList(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TMDB-dependent test in CI")
	}
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	tests := getImportListTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.TestImportList(tt.list)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errors)
		})
	}
}

type importListTestCase struct {
	name     string
	list     *models.ImportList
	expected bool
	errors   int
}

func getImportListTestCases() []importListTestCase {
	validList := &models.ImportList{
		Name: "TMDB Popular", Implementation: models.ImportListTypeTMDBPopular,
		QualityProfileID: 1, RootFolderPath: "/movies",
	}
	return []importListTestCase{
		{name: "Valid TMDB Popular list", list: validList, expected: true, errors: 0},
		{name: "Missing name", list: createImportListWithoutField("name", validList),
			expected: false, errors: 1},
		{name: "Missing implementation", list: createImportListWithoutField("implementation", validList),
			expected: false, errors: 1},
		{name: "Missing quality profile", list: createImportListWithoutField("qualityProfile", validList),
			expected: false, errors: 1},
		{name: "Missing root folder", list: createImportListWithoutField("rootFolder", validList),
			expected: false, errors: 1},
	}
}

func createImportListWithoutField(field string, base *models.ImportList) *models.ImportList {
	list := *base
	switch field {
	case "name":
		list.Name = ""
	case "implementation":
		list.Implementation = ""
	case "qualityProfile":
		list.QualityProfileID = 0
	case "rootFolder":
		list.RootFolderPath = ""
	}
	return &list
}

func TestImportListService_ValidateImportList(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TMDB-dependent test in CI")
	}
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	tests := getImportListValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateImportList(tt.list)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getImportListValidationTestCases() []struct {
	name    string
	list    *models.ImportList
	wantErr bool
	errMsg  string
} {
	validList := &models.ImportList{
		Name: "Test List", Implementation: models.ImportListTypeTMDBPopular,
		QualityProfileID: 1, RootFolderPath: "/movies",
	}
	return []struct {
		name    string
		list    *models.ImportList
		wantErr bool
		errMsg  string
	}{
		{name: "Valid list", list: validList, wantErr: false},
		{name: "Missing name", list: &models.ImportList{
			Implementation: models.ImportListTypeTMDBPopular, QualityProfileID: 1, RootFolderPath: "/movies"},
			wantErr: true, errMsg: "name is required"},
		{name: "Missing implementation", list: &models.ImportList{
			Name: "Test List", QualityProfileID: 1, RootFolderPath: "/movies"},
			wantErr: true, errMsg: "implementation is required"},
		{name: "Missing quality profile", list: &models.ImportList{
			Name: "Test List", Implementation: models.ImportListTypeTMDBPopular, RootFolderPath: "/movies"},
			wantErr: true, errMsg: "quality profile ID is required"},
		{name: "Missing root folder", list: &models.ImportList{
			Name: "Test List", Implementation: models.ImportListTypeTMDBPopular, QualityProfileID: 1},
			wantErr: true, errMsg: "root folder path is required"},
		{name: "TMDB list without list ID", list: &models.ImportList{
			Name: "Test List", Implementation: models.ImportListTypeTMDBList,
			QualityProfileID: 1, RootFolderPath: "/movies"},
			wantErr: true, errMsg: "list ID is required for TMDB lists"},
	}
}

func TestImportListService_ValidateImplementationRequirements(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	tests := getImplementationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateImplementationRequirements(tt.list)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getImplementationTestCases() []struct {
	name    string
	list    *models.ImportList
	wantErr bool
	errMsg  string
} {
	return []struct {
		name    string
		list    *models.ImportList
		wantErr bool
		errMsg  string
	}{
		{name: "TMDB list with list ID", list: &models.ImportList{
			Implementation: models.ImportListTypeTMDBList,
			Settings:       models.ImportListSettings{ListID: "123"}}, wantErr: false},
		{name: "TMDB list without list ID", list: &models.ImportList{
			Implementation: models.ImportListTypeTMDBList,
			Settings:       models.ImportListSettings{}}, wantErr: true, errMsg: "list ID is required for TMDB lists"},
		{name: "Trakt list with username", list: &models.ImportList{
			Implementation: models.ImportListTypeTrakt,
			Settings:       models.ImportListSettings{Username: "testuser"}}, wantErr: false},
		{name: "Trakt list without username", list: &models.ImportList{
			Implementation: models.ImportListTypeTrakt,
			Settings:       models.ImportListSettings{}}, wantErr: true, errMsg: "username is required for Trakt lists"},
		{name: "RSS import with URL", list: &models.ImportList{
			Implementation: models.ImportListTypeRSSImport,
			Settings:       models.ImportListSettings{URL: "https://example.com/feed.rss"}}, wantErr: false},
		{name: "RSS import without URL", list: &models.ImportList{
			Implementation: models.ImportListTypeRSSImport,
			Settings:       models.ImportListSettings{}}, wantErr: true, errMsg: "URL is required for RSS import"},
	}
}

func TestImportListService_SyncImportList(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.SyncImportList(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_GetImportListMovies(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.GetImportListMovies(nil, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_GetImportListStats(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	// Test with nil database
	_, err := service.GetImportListStats()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestImportListService_FetchSampleMovies(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewImportListService(nil, logger, nil, nil)

	list := &models.ImportList{
		ID:             1,
		Name:           "Test List",
		Implementation: models.ImportListTypeTMDBPopular,
	}

	movies, err := service.fetchSampleMovies(list, 5)
	assert.NoError(t, err)
	assert.NotEmpty(t, movies)
	assert.Equal(t, "Fight Club", movies[0].Title)
}

func TestImportList_Methods(t *testing.T) {
	list := &models.ImportList{
		Name:             "Test List",
		Implementation:   models.ImportListTypeTMDBPopular,
		Enabled:          true,
		EnableAuto:       true,
		QualityProfileID: 1,
		RootFolderPath:   "/movies",
	}

	// Test IsEnabled
	assert.True(t, list.IsEnabled())

	list.Enabled = false
	assert.False(t, list.IsEnabled())

	list.Enabled = true
	list.EnableAuto = false
	assert.False(t, list.IsEnabled())

	// Test GetListType
	assert.Equal(t, models.ImportListSourceTypeProgram, list.GetListType())

	list.ListType = models.ImportListSourceTypeAdvanced
	assert.Equal(t, models.ImportListSourceTypeAdvanced, list.GetListType())

	// Test RequiresAuthentication
	assert.False(t, list.RequiresAuthentication())

	list.Implementation = models.ImportListTypeTrakt
	assert.True(t, list.RequiresAuthentication())

	// Test GetBaseURL
	expectedURL := "https://api.trakt.tv"
	assert.Equal(t, expectedURL, list.GetBaseURL())

	list.Implementation = models.ImportListTypeTMDBPopular
	expectedTMDBURL := "https://api.themoviedb.org/3"
	assert.Equal(t, expectedTMDBURL, list.GetBaseURL())

	// Test ShouldAutoAdd
	list.EnableAuto = true
	list.Enabled = true
	assert.True(t, list.ShouldAutoAdd())

	list.EnableAuto = false
	assert.False(t, list.ShouldAutoAdd())
}

func TestImportListType_Constants(t *testing.T) {
	// Test that all import list type constants are properly defined
	assert.Equal(t, "TMDbCollectionImport", string(models.ImportListTypeTMDBCollection))
	assert.Equal(t, "TMDbCompanyImport", string(models.ImportListTypeTMDBCompany))
	assert.Equal(t, "TMDbKeywordImport", string(models.ImportListTypeTMDBKeyword))
	assert.Equal(t, "TMDbListImport", string(models.ImportListTypeTMDBList))
	assert.Equal(t, "TMDbPersonImport", string(models.ImportListTypeTMDBPerson))
	assert.Equal(t, "TMDbPopularImport", string(models.ImportListTypeTMDBPopular))
	assert.Equal(t, "TMDbUserImport", string(models.ImportListTypeTMDBUser))
	assert.Equal(t, "TraktImport", string(models.ImportListTypeTrakt))
	assert.Equal(t, "TraktListImport", string(models.ImportListTypeTraktList))
	assert.Equal(t, "TraktPopularImport", string(models.ImportListTypeTraktPopular))
	assert.Equal(t, "TraktUserImport", string(models.ImportListTypeTraktUser))
	assert.Equal(t, "PlexImport", string(models.ImportListTypePlexWatchlist))
	assert.Equal(t, "RadarrImport", string(models.ImportListTypeRadarrList))
	assert.Equal(t, "StevenLuImport", string(models.ImportListTypeStevenLu))
	assert.Equal(t, "RSSImport", string(models.ImportListTypeRSSImport))
	assert.Equal(t, "IMDbListImport", string(models.ImportListTypeIMDbList))
	assert.Equal(t, "CouchPotatoImport", string(models.ImportListTypeCouchPotato))
}

func TestImportListSourceType_Constants(t *testing.T) {
	// Test that all source type constants are properly defined
	assert.Equal(t, "program", string(models.ImportListSourceTypeProgram))
	assert.Equal(t, "other", string(models.ImportListSourceTypeOther))
	assert.Equal(t, "advanced", string(models.ImportListSourceTypeAdvanced))
}
