package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type MovieFile struct {
	ID                  int         `json:"id" db:"id" gorm:"primaryKey"`
	MovieID             int         `json:"movieId" db:"movie_id" gorm:"index"`
	RelativePath        string      `json:"relativePath" db:"relative_path"`
	Path                string      `json:"path" db:"path"`
	Size                int64       `json:"size" db:"size"`
	DateAdded           time.Time   `json:"dateAdded" db:"date_added"`
	SceneName           string      `json:"sceneName" db:"scene_name"`
	IndexerFlags        int         `json:"indexerFlags" db:"indexer_flags"`
	Quality             Quality     `json:"quality" db:"quality" gorm:"type:text"`
	CustomFormats       IntArray    `json:"customFormats" db:"custom_formats" gorm:"type:text"`
	CustomFormatScore   int         `json:"customFormatScore" db:"custom_format_score"`
	MediaInfo           MediaInfo   `json:"mediaInfo" db:"media_info" gorm:"type:text"`
	OriginalFilePath    string      `json:"originalFilePath" db:"original_file_path"`
	Languages           []Language  `json:"languages" db:"languages" gorm:"type:text"`
	ReleaseGroup        string      `json:"releaseGroup" db:"release_group"`
	Edition             string      `json:"edition" db:"edition"`
	
	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

type Quality struct {
	Quality  QualityDefinition `json:"quality"`
	Revision Revision          `json:"revision"`
}

func (q Quality) Value() (driver.Value, error) {
	return json.Marshal(q)
}

func (q *Quality) Scan(value interface{}) error {
	if value == nil {
		*q = Quality{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, q)
}

type QualityDefinition struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

type Revision struct {
	Version  int  `json:"version"`
	Real     int  `json:"real"`
	IsRepack bool `json:"isRepack"`
}

type MediaInfo struct {
	AudioBitrate         int     `json:"audioBitrate"`
	AudioChannels        float64 `json:"audioChannels"`
	AudioCodec           string  `json:"audioCodec"`
	AudioLanguages       string  `json:"audioLanguages"`
	AudioStreamCount     int     `json:"audioStreamCount"`
	VideoBitDepth        int     `json:"videoBitDepth"`
	VideoBitrate         int     `json:"videoBitrate"`
	VideoCodec           string  `json:"videoCodec"`
	VideoFps             float64 `json:"videoFps"`
	Resolution           string  `json:"resolution"`
	RunTime              string  `json:"runTime"`
	ScanType             string  `json:"scanType"`
	Subtitles            string  `json:"subtitles"`
	VideoMultiViewCount  int     `json:"videoMultiViewCount"`
	VideoColourPrimaries string  `json:"videoColourPrimaries"`
	VideoTransferCharacteristics string `json:"videoTransferCharacteristics"`
	SchemaRevision       int     `json:"schemaRevision"`
}

func (m MediaInfo) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MediaInfo) Scan(value interface{}) error {
	if value == nil {
		*m = MediaInfo{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, m)
}