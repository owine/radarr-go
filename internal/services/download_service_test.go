package services

import (
	"testing"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDownloadService_GetDownloadClients(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloadClients()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_GetDownloadClientByID(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloadClientByID(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_CreateDownloadClient(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	client := &models.DownloadClient{
		Name:     "Test Client",
		Type:     models.DownloadClientTypeQBittorrent,
		Protocol: models.DownloadProtocolTorrent,
		Host:     "localhost",
		Port:     8080,
	}

	// Test with nil database
	err := service.CreateDownloadClient(client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_UpdateDownloadClient(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	client := &models.DownloadClient{
		ID:       1,
		Name:     "Updated Client",
		Type:     models.DownloadClientTypeQBittorrent,
		Protocol: models.DownloadProtocolTorrent,
		Host:     "localhost",
		Port:     8080,
	}

	// Test with nil database
	err := service.UpdateDownloadClient(client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_DeleteDownloadClient(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	err := service.DeleteDownloadClient(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_GetEnabledDownloadClients(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetEnabledDownloadClients()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_TestDownloadClient(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	tests := getDownloadClientTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.TestDownloadClient(tt.client)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.IsValid)
			assert.Len(t, result.Errors, tt.errors)
		})
	}
}

func getDownloadClientTestCases() []struct {
	name     string
	client   *models.DownloadClient
	expected bool
	errors   int
} {
	return []struct {
		name     string
		client   *models.DownloadClient
		expected bool
		errors   int
	}{
		{
			name: "Valid qBittorrent client",
			client: &models.DownloadClient{
				Name:     "qBittorrent",
				Type:     models.DownloadClientTypeQBittorrent,
				Protocol: models.DownloadProtocolTorrent,
				Host:     "localhost",
				Port:     8080,
			},
			expected: false, // Will fail connection test
			errors:   1,
		},
		{
			name: "Missing name",
			client: &models.DownloadClient{
				Type:     models.DownloadClientTypeQBittorrent,
				Protocol: models.DownloadProtocolTorrent,
				Host:     "localhost",
				Port:     8080,
			},
			expected: false,
			errors:   1,
		},
		{
			name: "Missing host",
			client: &models.DownloadClient{
				Name:     "Test Client",
				Type:     models.DownloadClientTypeQBittorrent,
				Protocol: models.DownloadProtocolTorrent,
				Port:     8080,
			},
			expected: false,
			errors:   1,
		},
		{
			name: "Invalid port",
			client: &models.DownloadClient{
				Name:     "Test Client",
				Type:     models.DownloadClientTypeQBittorrent,
				Protocol: models.DownloadProtocolTorrent,
				Host:     "localhost",
				Port:     0,
			},
			expected: false,
			errors:   1,
		},
	}
}

func TestDownloadService_ValidateDownloadClient(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	tests := getValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateDownloadClient(tt.client)
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

func getValidationTestCases() []struct {
	name    string
	client  *models.DownloadClient
	wantErr bool
	errMsg  string
} {
	validClient := &models.DownloadClient{
		Name: "Test Client", Type: models.DownloadClientTypeQBittorrent,
		Protocol: models.DownloadProtocolTorrent, Host: "localhost", Port: 8080,
	}
	return []struct {
		name    string
		client  *models.DownloadClient
		wantErr bool
		errMsg  string
	}{
		{name: "Valid client", client: validClient, wantErr: false},
		{name: "Missing name", client: &models.DownloadClient{
			Type: models.DownloadClientTypeQBittorrent, Protocol: models.DownloadProtocolTorrent,
			Host: "localhost", Port: 8080}, wantErr: true, errMsg: "name is required"},
		{name: "Missing host", client: &models.DownloadClient{
			Name: "Test Client", Type: models.DownloadClientTypeQBittorrent,
			Protocol: models.DownloadProtocolTorrent, Port: 8080}, wantErr: true, errMsg: "host is required"},
		{name: "Invalid port", client: &models.DownloadClient{
			Name: "Test Client", Type: models.DownloadClientTypeQBittorrent,
			Protocol: models.DownloadProtocolTorrent, Host: "localhost", Port: 0},
			wantErr: true, errMsg: "port must be between 1 and 65535"},
		{name: "Missing type", client: &models.DownloadClient{
			Name: "Test Client", Protocol: models.DownloadProtocolTorrent,
			Host: "localhost", Port: 8080}, wantErr: true, errMsg: "client type is required"},
		{name: "Invalid protocol", client: &models.DownloadClient{
			Name: "Test Client", Type: models.DownloadClientTypeQBittorrent,
			Protocol: models.DownloadProtocolUsenet, Host: "localhost", Port: 8080},
			wantErr: true, errMsg: "protocol usenet is not supported"},
	}
}

func TestDownloadService_IsValidProtocolForClientType(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	tests := []struct {
		name       string
		clientType models.DownloadClientType
		protocol   models.DownloadProtocol
		expected   bool
	}{
		{
			name:       "qBittorrent with torrent",
			clientType: models.DownloadClientTypeQBittorrent,
			protocol:   models.DownloadProtocolTorrent,
			expected:   true,
		},
		{
			name:       "qBittorrent with usenet",
			clientType: models.DownloadClientTypeQBittorrent,
			protocol:   models.DownloadProtocolUsenet,
			expected:   false,
		},
		{
			name:       "SABnzbd with usenet",
			clientType: models.DownloadClientTypeSABnzbd,
			protocol:   models.DownloadProtocolUsenet,
			expected:   true,
		},
		{
			name:       "SABnzbd with torrent",
			clientType: models.DownloadClientTypeSABnzbd,
			protocol:   models.DownloadProtocolTorrent,
			expected:   false,
		},
		{
			name:       "NZBGet with usenet",
			clientType: models.DownloadClientTypeNZBGet,
			protocol:   models.DownloadProtocolUsenet,
			expected:   true,
		},
		{
			name:       "Transmission with torrent",
			clientType: models.DownloadClientTypeTransmission,
			protocol:   models.DownloadProtocolTorrent,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidProtocolForClientType(tt.clientType, tt.protocol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDownloadService_GetDownloadHistory(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloadHistory(10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_AddToHistory(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	history := &models.DownloadHistory{
		SourceTitle: "Test Movie",
		Protocol:    models.DownloadProtocolTorrent,
		Successful:  true,
	}

	// Test with nil database
	err := service.AddToHistory(history)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_GetDownloadClientStats(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloadClientStats()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_GetDownloadClientsByProtocol(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloadClientsByProtocol(models.DownloadProtocolTorrent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadService_GetDownloads(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewDownloadService(nil, logger)

	// Test with nil database
	_, err := service.GetDownloads()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestDownloadClientType_Constants(t *testing.T) {
	// Test that all download client type constants are properly defined
	assert.Equal(t, "qbittorrent", string(models.DownloadClientTypeQBittorrent))
	assert.Equal(t, "transmission", string(models.DownloadClientTypeTransmission))
	assert.Equal(t, "deluge", string(models.DownloadClientTypeDeluge))
	assert.Equal(t, "sabnzbd", string(models.DownloadClientTypeSABnzbd))
	assert.Equal(t, "nzbget", string(models.DownloadClientTypeNZBGet))
	assert.Equal(t, "rtorrent", string(models.DownloadClientTypeRTorrent))
	assert.Equal(t, "utorrent", string(models.DownloadClientTypeUtorrent))
}

func TestDownloadClient_Methods(t *testing.T) {
	client := &models.DownloadClient{
		Name:     "Test Client",
		Type:     models.DownloadClientTypeQBittorrent,
		Protocol: models.DownloadProtocolTorrent,
		Host:     "localhost",
		Port:     8080,
		Enable:   true,
		UseSsl:   false,
	}

	// Test IsEnabled
	assert.True(t, client.IsEnabled())

	client.Enable = false
	assert.False(t, client.IsEnabled())

	// Test SupportsProtocol
	assert.True(t, client.SupportsProtocol(models.DownloadProtocolTorrent))
	assert.False(t, client.SupportsProtocol(models.DownloadProtocolUsenet))

	// Test GetBaseURL
	expectedURL := "http://localhost:8080"
	assert.Equal(t, expectedURL, client.GetBaseURL())

	client.UseSsl = true
	expectedURLSSL := "https://localhost:8080"
	assert.Equal(t, expectedURLSSL, client.GetBaseURL())
}
