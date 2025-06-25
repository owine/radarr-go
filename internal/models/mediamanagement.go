package models

import (
	"time"
)

// MediaManagementConfig represents the media management configuration for Radarr
type MediaManagementConfig struct {
	ID                         int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	AutoUnmonitorPreviousMovie bool                   `json:"autoUnmonitorPreviouslyDownloadedMovies" gorm:"default:false"`
	RecycleBin                 string                 `json:"recycleBin" gorm:"default:''"`
	RecycleBinCleanup          int                    `json:"recycleBinCleanupDays" gorm:"default:7"`
	DownloadPropersAndRepacks  ProperRepackSetting    `json:"downloadPropersAndRepacks" gorm:"default:'preferAndUpgrade'"`
	CreateEmptyFolders         bool                   `json:"createEmptyMovieFolders" gorm:"default:false"`
	DeleteEmptyFolders         bool                   `json:"deleteEmptyFolders" gorm:"default:false"`
	FileDate                   FileDateType           `json:"fileDate" gorm:"default:'none'"`
	RescanAfterRefresh         RescanAfterRefreshType `json:"rescanAfterRefresh" gorm:"default:'always'"`
	AllowFingerprinting        FingerprintingType     `json:"allowFingerprinting" gorm:"default:'newFiles'"`
	SetPermissions             bool                   `json:"setPermissionsLinux" gorm:"default:false"`
	ChmodFolder                string                 `json:"chmodFolder" gorm:"default:'755'"`
	ChownGroup                 string                 `json:"chownGroup" gorm:"default:''"`
	SkipFreeSpaceCheck         bool                   `json:"skipFreeSpaceCheckWhenImporting" gorm:"default:false"`
	MinimumFreeSpace           int64                  `json:"minimumFreeSpaceWhenImporting" gorm:"default:100"`
	CopyUsingHardlinks         bool                   `json:"copyUsingHardlinks" gorm:"default:true"`
	UseScriptImport            bool                   `json:"useScriptImport" gorm:"default:false"`
	ScriptImportPath           string                 `json:"scriptImportPath" gorm:"default:''"`
	ImportExtraFiles           bool                   `json:"importExtraFiles" gorm:"default:false"`
	ExtraFileExtensions        StringArray            `json:"extraFileExtensions" gorm:"type:text"`
	EnableMediaInfo            bool                   `json:"enableMediaInfo" gorm:"default:true"`
	ImportMechanism            ImportMechanism        `json:"importMechanism" gorm:"default:'move'"`
	WatchLibraryForChanges     bool                   `json:"watchLibraryForChanges" gorm:"default:true"`
	CreatedAt                  time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt                  time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the MediaManagementConfig model
func (MediaManagementConfig) TableName() string {
	return "media_management_config"
}

// ProperRepackSetting represents how to handle propers and repacks
type ProperRepackSetting string

const (
	// ProperRepackPreferAndUpgrade downloads propers/repacks and upgrades existing files
	ProperRepackPreferAndUpgrade ProperRepackSetting = "preferAndUpgrade"
	// ProperRepackDoNotUpgrade downloads propers/repacks but doesn't upgrade existing files
	ProperRepackDoNotUpgrade ProperRepackSetting = "doNotUpgrade"
	// ProperRepackDoNotPrefer doesn't download propers/repacks automatically
	ProperRepackDoNotPrefer ProperRepackSetting = "doNotPrefer"
)

// FileDateType represents how file dates should be set
type FileDateType string

const (
	// FileDateNone keeps original file dates
	FileDateNone FileDateType = "none"
	// FileDateLocalAirDate sets file date to local air date
	FileDateLocalAirDate FileDateType = "localAirDate"
	// FileDateUTCAirDate sets file date to UTC air date
	FileDateUTCAirDate FileDateType = "utcAirDate"
)

// RescanAfterRefreshType represents when to rescan after refresh
type RescanAfterRefreshType string

const (
	// RescanAfterRefreshAlways always rescans after refresh
	RescanAfterRefreshAlways RescanAfterRefreshType = "always"
	// RescanAfterRefreshAfterManual only rescans after manual refresh
	RescanAfterRefreshAfterManual RescanAfterRefreshType = "afterManual"
	// RescanAfterRefreshNever never rescans after refresh
	RescanAfterRefreshNever RescanAfterRefreshType = "never"
)

// FingerprintingType represents when to allow fingerprinting
type FingerprintingType string

const (
	// FingerprintingAlways always allows fingerprinting
	FingerprintingAlways FingerprintingType = "allFiles"
	// FingerprintingNewFiles only allows fingerprinting for new files
	FingerprintingNewFiles FingerprintingType = "newFiles"
	// FingerprintingNever never allows fingerprinting
	FingerprintingNever FingerprintingType = "never"
)

// ImportMechanism represents how files should be imported
type ImportMechanism string

const (
	// ImportMechanismMove moves files during import
	ImportMechanismMove ImportMechanism = "move"
	// ImportMechanismCopy copies files during import
	ImportMechanismCopy ImportMechanism = "copy"
	// ImportMechanismHardlink creates hardlinks during import
	ImportMechanismHardlink ImportMechanism = "hardlink"
)

// RootFolder represents a root folder configuration
type RootFolder struct {
	ID                           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Path                         string    `json:"path" gorm:"not null;uniqueIndex"`
	Accessible                   bool      `json:"accessible" gorm:"default:true"`
	FreeSpace                    int64     `json:"freeSpace" gorm:"default:0"`
	TotalSpace                   int64     `json:"totalSpace" gorm:"default:0"`
	UnmappedFolders              IntArray  `json:"unmappedFolders" gorm:"type:text"`
	DefaultTags                  IntArray  `json:"defaultTags" gorm:"type:text"`
	DefaultQualityProfileID      int       `json:"defaultQualityProfileId" gorm:"default:0"`
	DefaultMonitorOption         string    `json:"defaultMonitorOption" gorm:"default:'movieOnly'"`
	DefaultSearchForMissingMovie bool      `json:"defaultSearchForMissingMovie" gorm:"default:false"`
	CreatedAt                    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt                    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the RootFolder model
func (RootFolder) TableName() string {
	return "root_folders"
}

// UnmappedFolder represents a folder that exists but isn't mapped to a movie
type UnmappedFolder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// GetDefaultMediaManagementConfig returns the default media management configuration
func GetDefaultMediaManagementConfig() *MediaManagementConfig {
	return &MediaManagementConfig{
		AutoUnmonitorPreviousMovie: false,
		RecycleBin:                 "",
		RecycleBinCleanup:          7,
		DownloadPropersAndRepacks:  ProperRepackPreferAndUpgrade,
		CreateEmptyFolders:         false,
		DeleteEmptyFolders:         false,
		FileDate:                   FileDateNone,
		RescanAfterRefresh:         RescanAfterRefreshAlways,
		AllowFingerprinting:        FingerprintingNewFiles,
		SetPermissions:             false,
		ChmodFolder:                "755",
		ChownGroup:                 "",
		SkipFreeSpaceCheck:         false,
		MinimumFreeSpace:           100,
		CopyUsingHardlinks:         true,
		UseScriptImport:            false,
		ScriptImportPath:           "",
		ImportExtraFiles:           false,
		ExtraFileExtensions:        StringArray{"srt", "nfo"},
		EnableMediaInfo:            true,
		ImportMechanism:            ImportMechanismMove,
		WatchLibraryForChanges:     true,
	}
}

// IsRecycleBinEnabled returns true if recycle bin is configured
func (mmc *MediaManagementConfig) IsRecycleBinEnabled() bool {
	return mmc.RecycleBin != ""
}

// IsPermissionsEnabled returns true if permission setting is enabled
func (mmc *MediaManagementConfig) IsPermissionsEnabled() bool {
	return mmc.SetPermissions
}

// IsScriptImportEnabled returns true if script import is enabled
func (mmc *MediaManagementConfig) IsScriptImportEnabled() bool {
	return mmc.UseScriptImport && mmc.ScriptImportPath != ""
}

// ValidateConfiguration validates the media management configuration
func (mmc *MediaManagementConfig) ValidateConfiguration() []string {
	var errors []string

	// Validate recycle bin cleanup days
	if mmc.RecycleBinCleanup < 0 {
		errors = append(errors, "Recycle bin cleanup days cannot be negative")
	}

	// Validate minimum free space
	if mmc.MinimumFreeSpace < 0 {
		errors = append(errors, "Minimum free space cannot be negative")
	}

	// Validate chmod folder format
	if mmc.SetPermissions && mmc.ChmodFolder != "" {
		if len(mmc.ChmodFolder) != 3 {
			errors = append(errors, "Chmod folder must be 3 digits (e.g., 755)")
		}
		for _, char := range mmc.ChmodFolder {
			if char < '0' || char > '7' {
				errors = append(errors, "Chmod folder must contain only octal digits (0-7)")
				break
			}
		}
	}

	// Validate script import path
	if mmc.UseScriptImport && mmc.ScriptImportPath == "" {
		errors = append(errors, "Script import path is required when script import is enabled")
	}

	// Validate extra file extensions
	for _, ext := range mmc.ExtraFileExtensions {
		if ext == "" {
			errors = append(errors, "Extra file extensions cannot be empty")
		}
	}

	return errors
}

// ValidateRootFolder validates a root folder configuration
func ValidateRootFolder(rf *RootFolder) []string {
	var errors []string

	if rf.Path == "" {
		errors = append(errors, "Root folder path is required")
	}

	if rf.DefaultQualityProfileID < 0 {
		errors = append(errors, "Default quality profile ID cannot be negative")
	}

	// Validate default monitor option
	validMonitorOptions := []string{"movieOnly", "movieAndCollection", "none"}
	validOption := false
	for _, option := range validMonitorOptions {
		if rf.DefaultMonitorOption == option {
			validOption = true
			break
		}
	}
	if !validOption {
		errors = append(errors, "Invalid default monitor option")
	}

	return errors
}

// RootFolderStats represents statistics about a root folder
type RootFolderStats struct {
	FreeSpace      int64   `json:"freeSpace"`
	TotalSpace     int64   `json:"totalSpace"`
	MovieCount     int     `json:"movieCount"`
	UnmappedCount  int     `json:"unmappedFolderCount"`
	PercentageFree float64 `json:"percentageFree"`
}

// CalculateStats calculates statistics for a root folder
func (rf *RootFolder) CalculateStats(movieCount int) RootFolderStats {
	var percentageFree float64
	if rf.TotalSpace > 0 {
		percentageFree = float64(rf.FreeSpace) / float64(rf.TotalSpace) * 100
	}

	return RootFolderStats{
		FreeSpace:      rf.FreeSpace,
		TotalSpace:     rf.TotalSpace,
		MovieCount:     movieCount,
		UnmappedCount:  len(rf.UnmappedFolders),
		PercentageFree: percentageFree,
	}
}

// IsAccessible returns true if the root folder is accessible
func (rf *RootFolder) IsAccessible() bool {
	return rf.Accessible
}

// HasEnoughSpace checks if the root folder has enough free space (in MB)
func (rf *RootFolder) HasEnoughSpace(requiredMB int64) bool {
	if rf.FreeSpace == 0 {
		return true // Unknown space, assume it's fine
	}
	return rf.FreeSpace >= requiredMB
}

// GetSpaceUsagePercentage returns the percentage of space used
func (rf *RootFolder) GetSpaceUsagePercentage() float64 {
	if rf.TotalSpace == 0 {
		return 0
	}
	used := rf.TotalSpace - rf.FreeSpace
	return float64(used) / float64(rf.TotalSpace) * 100
}
