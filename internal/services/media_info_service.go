package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// MediaInfoService provides media information extraction from video files
type MediaInfoService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewMediaInfoService creates a new instance of MediaInfoService
func NewMediaInfoService(db *database.Database, logger *logger.Logger) *MediaInfoService {
	return &MediaInfoService{
		db:     db,
		logger: logger,
	}
}

// ExtractMediaInfo extracts comprehensive media information from a video file
func (s *MediaInfoService) ExtractMediaInfo(filePath string) (*models.MediaInfo, error) {
	s.logger.Debug("Extracting media info", "file", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Try ffprobe first (most reliable)
	if mediaInfo, err := s.extractWithFFProbe(filePath); err == nil {
		return mediaInfo, nil
	}

	// Fallback to mediainfo command
	if mediaInfo, err := s.extractWithMediaInfo(filePath); err == nil {
		return mediaInfo, nil
	}

	// Fallback to basic file analysis
	mediaInfo, err := s.extractBasicInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract media info: %w", err)
	}

	return mediaInfo, nil
}

// extractWithFFProbe uses ffprobe to extract detailed media information
func (s *MediaInfoService) extractWithFFProbe(filePath string) (*models.MediaInfo, error) {
	// Check if ffprobe is available
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, fmt.Errorf("ffprobe not found")
	}

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe execution failed: %w", err)
	}

	return s.parseFFProbeOutput(string(output))
}

// extractWithMediaInfo uses mediainfo command to extract media information
func (s *MediaInfoService) extractWithMediaInfo(filePath string) (*models.MediaInfo, error) {
	// Check if mediainfo is available
	if _, err := exec.LookPath("mediainfo"); err != nil {
		return nil, fmt.Errorf("mediainfo not found")
	}

	cmd := exec.Command("mediainfo", "--Output=JSON", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("mediainfo execution failed: %w", err)
	}

	return s.parseMediaInfoOutput(string(output))
}

// extractBasicInfo extracts basic information using file properties and naming patterns
func (s *MediaInfoService) extractBasicInfo(filePath string) (*models.MediaInfo, error) {
	fileName := filepath.Base(filePath)

	mediaInfo := &models.MediaInfo{
		SchemaRevision: 1,
	}

	// Extract quality information from filename
	if resolution := s.extractResolutionFromFilename(fileName); resolution != "" {
		mediaInfo.Resolution = resolution
	}

	// Extract codec information from filename
	if videoCodec := s.extractVideoCodecFromFilename(fileName); videoCodec != "" {
		mediaInfo.VideoCodec = videoCodec
	}

	if audioCodec := s.extractAudioCodecFromFilename(fileName); audioCodec != "" {
		mediaInfo.AudioCodec = audioCodec
	}

	// Get file duration using basic methods if possible
	if fileInfo, err := os.Stat(filePath); err == nil {
		// File size in bytes
		// Note: This is just basic info, not actual media duration
		_ = fileInfo.Size()
	}

	s.logger.Debug("Extracted basic media info",
		"file", fileName,
		"resolution", mediaInfo.Resolution,
		"videoCodec", mediaInfo.VideoCodec,
		"audioCodec", mediaInfo.AudioCodec)

	return mediaInfo, nil
}

// parseFFProbeOutput parses ffprobe JSON output to extract media information
func (s *MediaInfoService) parseFFProbeOutput(output string) (*models.MediaInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would properly parse the JSON output from ffprobe

	mediaInfo := &models.MediaInfo{
		SchemaRevision: 1,
	}

	// Extract video stream info
	if matches := regexp.MustCompile(`"codec_name":\s*"([^"]+)"`).FindStringSubmatch(output); len(matches) > 1 {
		mediaInfo.VideoCodec = matches[1]
	}

	if matches := regexp.MustCompile(`"width":\s*(\d+)`).FindStringSubmatch(output); len(matches) > 1 {
		if width, err := strconv.Atoi(matches[1]); err == nil {
			if height_matches := regexp.MustCompile(`"height":\s*(\d+)`).FindStringSubmatch(output); len(height_matches) > 1 {
				if height, err := strconv.Atoi(height_matches[1]); err == nil {
					mediaInfo.Resolution = fmt.Sprintf("%dx%d", width, height)
				}
			}
		}
	}

	if matches := regexp.MustCompile(`"r_frame_rate":\s*"([^"]+)"`).FindStringSubmatch(output); len(matches) > 1 {
		if fps, err := s.parseFrameRate(matches[1]); err == nil {
			mediaInfo.VideoFps = fps
		}
	}

	if matches := regexp.MustCompile(`"bit_rate":\s*"(\d+)"`).FindStringSubmatch(output); len(matches) > 1 {
		if bitrate, err := strconv.Atoi(matches[1]); err == nil {
			mediaInfo.VideoBitrate = bitrate
		}
	}

	// Extract audio stream info (this would be more complex in reality)
	audioCodecRegex := regexp.MustCompile(`"codec_name":\s*"(aac|mp3|ac3|dts|flac|truehd|eac3)"`)
	if matches := audioCodecRegex.FindStringSubmatch(output); len(matches) > 1 {
		mediaInfo.AudioCodec = strings.ToUpper(matches[1])
	}

	if matches := regexp.MustCompile(`"channels":\s*(\d+)`).FindStringSubmatch(output); len(matches) > 1 {
		if channels, err := strconv.ParseFloat(matches[1], 64); err == nil {
			mediaInfo.AudioChannels = channels
		}
	}

	return mediaInfo, nil
}

// parseMediaInfoOutput parses mediainfo JSON output to extract media information
func (s *MediaInfoService) parseMediaInfoOutput(output string) (*models.MediaInfo, error) {
	// This is a simplified implementation
	// In a real implementation, you would properly parse the JSON output from mediainfo

	mediaInfo := &models.MediaInfo{
		SchemaRevision: 1,
	}

	// Basic parsing - in reality this would be much more comprehensive
	if matches := regexp.MustCompile(`"Format":\s*"([^"]+)"`).FindStringSubmatch(output); len(matches) > 1 {
		mediaInfo.VideoCodec = matches[1]
	}

	return mediaInfo, nil
}

// extractResolutionFromFilename attempts to extract resolution from filename
func (s *MediaInfoService) extractResolutionFromFilename(filename string) string {
	filename = strings.ToLower(filename)

	// Common resolution patterns
	resolutions := map[string]string{
		"2160p": "3840x2160",
		"1080p": "1920x1080",
		"720p":  "1280x720",
		"480p":  "720x480",
		"4k":    "3840x2160",
		"1080i": "1920x1080",
		"720i":  "1280x720",
	}

	for pattern, resolution := range resolutions {
		if strings.Contains(filename, pattern) {
			return resolution
		}
	}

	return ""
}

// extractVideoCodecFromFilename attempts to extract video codec from filename
func (s *MediaInfoService) extractVideoCodecFromFilename(filename string) string {
	filename = strings.ToLower(filename)

	// Common codec patterns
	codecs := map[string]string{
		"x264": "H.264",
		"h264": "H.264",
		"avc":  "H.264",
		"x265": "H.265",
		"h265": "H.265",
		"hevc": "H.265",
		"xvid": "Xvid",
		"divx": "DivX",
		"av1":  "AV1",
		"vp9":  "VP9",
	}

	for pattern, codec := range codecs {
		if strings.Contains(filename, pattern) {
			return codec
		}
	}

	return ""
}

// extractAudioCodecFromFilename attempts to extract audio codec from filename
func (s *MediaInfoService) extractAudioCodecFromFilename(filename string) string {
	filename = strings.ToLower(filename)

	// Common audio codec patterns
	codecs := map[string]string{
		"aac":    "AAC",
		"ac3":    "AC3",
		"dts":    "DTS",
		"dts-hd": "DTS-HD",
		"truehd": "TrueHD",
		"eac3":   "EAC3",
		"flac":   "FLAC",
		"mp3":    "MP3",
		"opus":   "Opus",
		"vorbis": "Vorbis",
		"pcm":    "PCM",
	}

	for pattern, codec := range codecs {
		if strings.Contains(filename, pattern) {
			return codec
		}
	}

	return ""
}

// parseFrameRate parses frame rate string (e.g., "24000/1001" or "25")
func (s *MediaInfoService) parseFrameRate(rateStr string) (float64, error) {
	if strings.Contains(rateStr, "/") {
		parts := strings.Split(rateStr, "/")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid frame rate format")
		}

		numerator, err1 := strconv.ParseFloat(parts[0], 64)
		denominator, err2 := strconv.ParseFloat(parts[1], 64)

		if err1 != nil || err2 != nil || denominator == 0 {
			return 0, fmt.Errorf("invalid frame rate values")
		}

		return numerator / denominator, nil
	}

	return strconv.ParseFloat(rateStr, 64)
}

// GetMediaInfo retrieves stored media info for a movie file
func (s *MediaInfoService) GetMediaInfo(movieFileID int) (*models.MediaInfo, error) {
	var movieFile models.MovieFile

	err := s.db.GORM.Where("id = ?", movieFileID).First(&movieFile).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get movie file: %w", err)
	}

	return &movieFile.MediaInfo, nil
}

// UpdateMediaInfo updates media information for a movie file
func (s *MediaInfoService) UpdateMediaInfo(movieFileID int, mediaInfo *models.MediaInfo) error {
	var movieFile models.MovieFile

	err := s.db.GORM.Where("id = ?", movieFileID).First(&movieFile).Error
	if err != nil {
		return fmt.Errorf("failed to get movie file: %w", err)
	}

	movieFile.MediaInfo = *mediaInfo

	err = s.db.GORM.Save(&movieFile).Error
	if err != nil {
		return fmt.Errorf("failed to update movie file media info: %w", err)
	}

	s.logger.Info("Updated media info for movie file", "movieFileId", movieFileID)
	return nil
}

// RefreshMediaInfo re-extracts media information for a movie file
func (s *MediaInfoService) RefreshMediaInfo(movieFileID int) error {
	var movieFile models.MovieFile

	err := s.db.GORM.Where("id = ?", movieFileID).First(&movieFile).Error
	if err != nil {
		return fmt.Errorf("failed to get movie file: %w", err)
	}

	// Extract fresh media info
	mediaInfo, err := s.ExtractMediaInfo(movieFile.Path)
	if err != nil {
		return fmt.Errorf("failed to extract media info: %w", err)
	}

	// Update the movie file record
	movieFile.MediaInfo = *mediaInfo

	err = s.db.GORM.Save(&movieFile).Error
	if err != nil {
		return fmt.Errorf("failed to save updated media info: %w", err)
	}

	s.logger.Info("Refreshed media info for movie file",
		"movieFileId", movieFileID,
		"path", movieFile.Path)

	return nil
}

// ValidateMediaInfo validates that a video file has extractable media information
func (s *MediaInfoService) ValidateMediaInfo(filePath string) error {
	mediaInfo, err := s.ExtractMediaInfo(filePath)
	if err != nil {
		return err
	}

	// Basic validation checks
	if mediaInfo.VideoCodec == "" && mediaInfo.Resolution == "" {
		return fmt.Errorf("no video information detected")
	}

	return nil
}

// GetSupportedFormats returns a list of supported video formats
func (s *MediaInfoService) GetSupportedFormats() []string {
	return []string{
		".mp4", ".mkv", ".avi", ".wmv", ".mov", ".flv",
		".m4v", ".mpg", ".mpeg", ".ts", ".webm", ".vob",
		".divx", ".xvid", ".asf", ".rm", ".rmvb", ".3gp",
	}
}

// IsVideoFile checks if a file is a supported video format
func (s *MediaInfoService) IsVideoFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	supportedFormats := s.GetSupportedFormats()
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}

	return false
}

// GetVideoFileSize gets the size of a video file
func (s *MediaInfoService) GetVideoFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	return fileInfo.Size(), nil
}

// GetVideoFileDuration attempts to get video duration from media info
func (s *MediaInfoService) GetVideoFileDuration(filePath string) (time.Duration, error) {
	// This would require more sophisticated media analysis
	// For now, return a placeholder implementation

	// In a real implementation, this would use ffprobe or mediainfo to get duration
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries",
		"format=duration", "-of", "csv=p=0", filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get duration: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return time.Duration(durationFloat * float64(time.Second)), nil
}
