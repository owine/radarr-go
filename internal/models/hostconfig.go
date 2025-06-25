package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// HostConfig represents the host/system configuration for Radarr
type HostConfig struct {
	ID                     int             `json:"id" gorm:"primaryKey;autoIncrement"`
	BindAddress            string          `json:"bindAddress" gorm:"not null;default:'*'"`
	Port                   int             `json:"port" gorm:"not null;default:7878"`
	URLBase                string          `json:"urlBase" gorm:"default:''"`
	EnableSSL              bool            `json:"enableSsl" gorm:"default:false"`
	SSLPort                int             `json:"sslPort" gorm:"default:6969"`
	SSLCertPath            string          `json:"sslCertPath" gorm:"default:''"`
	SSLKeyPath             string          `json:"sslKeyPath" gorm:"default:''"`
	Username               string          `json:"username" gorm:"default:''"`
	Password               string          `json:"password" gorm:"default:''"`
	AuthenticationMethod   AuthMethod      `json:"authenticationMethod" gorm:"default:'none'"`
	AuthenticationRequired AuthRequired    `json:"authenticationRequired" gorm:"default:'enabled'"`
	LogLevel               LogLevel        `json:"logLevel" gorm:"default:'info'"`
	LaunchBrowser          bool            `json:"launchBrowser" gorm:"default:true"`
	EnableColorImpared     bool            `json:"enableColorImpairedMode" gorm:"default:false"`
	ProxySettings          ProxySettings   `json:"proxySettings" gorm:"type:text"`
	UpdateMechanism        UpdateMechanism `json:"updateMechanism" gorm:"default:'builtin'"`
	UpdateBranch           string          `json:"updateBranch" gorm:"default:'master'"`
	UpdateAutomatically    bool            `json:"updateAutomatically" gorm:"default:false"`
	UpdateScriptPath       string          `json:"updateScriptPath" gorm:"default:''"`
	AnalyticsEnabled       bool            `json:"analyticsEnabled" gorm:"default:true"`
	BackupFolder           string          `json:"backupFolder" gorm:"default:''"`
	BackupInterval         int             `json:"backupInterval" gorm:"default:7"`
	BackupRetention        int             `json:"backupRetention" gorm:"default:28"`
	CertificateValidation  CertValidation  `json:"certificateValidation" gorm:"default:'enabled'"`
	CreatedAt              time.Time       `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt              time.Time       `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the HostConfig model
func (HostConfig) TableName() string {
	return "host_config"
}

// AuthMethod represents the authentication method
type AuthMethod string

const (
	// AuthMethodNone represents no authentication
	AuthMethodNone AuthMethod = "none"
	// AuthMethodBasic represents basic HTTP authentication
	AuthMethodBasic AuthMethod = "basic"
	// AuthMethodForms represents forms-based authentication
	AuthMethodForms AuthMethod = "forms"
	// AuthMethodExternal represents external authentication
	AuthMethodExternal AuthMethod = "external"
)

// AuthRequired represents when authentication is required
type AuthRequired string

const (
	// AuthRequiredEnabled requires authentication for all requests
	AuthRequiredEnabled AuthRequired = "enabled"
	// AuthRequiredForExternalRequests requires authentication only for external requests
	AuthRequiredForExternalRequests AuthRequired = "disabledForLocalAddresses"
)

// LogLevel represents the application log level
type LogLevel string

const (
	// LogLevelTrace represents trace level logging
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug represents debug level logging
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info level logging
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn represents warning level logging
	LogLevelWarn LogLevel = "warn"
	// LogLevelError represents error level logging
	LogLevelError LogLevel = "error"
	// LogLevelFatal represents fatal level logging
	LogLevelFatal LogLevel = "fatal"
)

// UpdateMechanism represents how updates are handled
type UpdateMechanism string

const (
	// UpdateMechanismBuiltIn represents built-in update mechanism
	UpdateMechanismBuiltIn UpdateMechanism = "builtin"
	// UpdateMechanismScript represents script-based updates
	UpdateMechanismScript UpdateMechanism = "script"
	// UpdateMechanismExternal represents external update management
	UpdateMechanismExternal UpdateMechanism = "external"
	// UpdateMechanismDocker represents Docker-based updates
	UpdateMechanismDocker UpdateMechanism = "docker"
)

// CertValidation represents certificate validation settings
type CertValidation string

const (
	// CertValidationEnabled enables certificate validation
	CertValidationEnabled CertValidation = "enabled"
	// CertValidationForExternalRequests enables validation only for external requests
	CertValidationForExternalRequests CertValidation = "disabledForLocalAddresses"
	// CertValidationDisabled disables certificate validation
	CertValidationDisabled CertValidation = "disabled"
)

// ProxySettings represents proxy configuration
type ProxySettings struct {
	Type         ProxyType `json:"type"`
	Hostname     string    `json:"hostname"`
	Port         int       `json:"port"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	BypassFilter string    `json:"bypassFilter"`
	BypassLocal  bool      `json:"bypassLocalAddress"`
}

// ProxyType represents the type of proxy
type ProxyType string

const (
	// ProxyTypeHTTP represents HTTP proxy
	ProxyTypeHTTP ProxyType = "http"
	// ProxyTypeSocks4 represents SOCKS4 proxy
	ProxyTypeSocks4 ProxyType = "socks4"
	// ProxyTypeSocks5 represents SOCKS5 proxy
	ProxyTypeSocks5 ProxyType = "socks5"
)

// Value implements the driver.Valuer interface for database storage
func (ps ProxySettings) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

// Scan implements the sql.Scanner interface for database retrieval
func (ps *ProxySettings) Scan(value interface{}) error {
	if value == nil {
		*ps = ProxySettings{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ps)
	case string:
		return json.Unmarshal([]byte(v), ps)
	default:
		return fmt.Errorf("cannot scan %T into ProxySettings", value)
	}
}

// GetDefaultHostConfig returns the default host configuration
func GetDefaultHostConfig() *HostConfig {
	return &HostConfig{
		BindAddress:            "*",
		Port:                   7878,
		URLBase:                "",
		EnableSSL:              false,
		SSLPort:                6969,
		AuthenticationMethod:   AuthMethodNone,
		AuthenticationRequired: AuthRequiredEnabled,
		LogLevel:               LogLevelInfo,
		LaunchBrowser:          true,
		EnableColorImpared:     false,
		ProxySettings:          ProxySettings{},
		UpdateMechanism:        UpdateMechanismBuiltIn,
		UpdateBranch:           "master",
		UpdateAutomatically:    false,
		AnalyticsEnabled:       true,
		BackupInterval:         7,
		BackupRetention:        28,
		CertificateValidation:  CertValidationEnabled,
	}
}

// IsAuthenticationEnabled returns true if authentication is enabled
func (hc *HostConfig) IsAuthenticationEnabled() bool {
	return hc.AuthenticationMethod != AuthMethodNone
}

// IsSSLEnabled returns true if SSL is enabled
func (hc *HostConfig) IsSSLEnabled() bool {
	return hc.EnableSSL
}

// HasProxyConfigured returns true if proxy is configured
func (hc *HostConfig) HasProxyConfigured() bool {
	return hc.ProxySettings.Hostname != ""
}

// GetEffectivePort returns the port that should be used (SSL or regular)
func (hc *HostConfig) GetEffectivePort() int {
	if hc.EnableSSL {
		return hc.SSLPort
	}
	return hc.Port
}

// ValidateConfiguration validates the host configuration
func (hc *HostConfig) ValidateConfiguration() []string {
	var errors []string

	if hc.Port < 1 || hc.Port > 65535 {
		errors = append(errors, "Port must be between 1 and 65535")
	}

	if hc.EnableSSL {
		if hc.SSLPort < 1 || hc.SSLPort > 65535 {
			errors = append(errors, "SSL Port must be between 1 and 65535")
		}
		if hc.SSLCertPath == "" {
			errors = append(errors, "SSL Certificate path is required when SSL is enabled")
		}
		if hc.SSLKeyPath == "" {
			errors = append(errors, "SSL Key path is required when SSL is enabled")
		}
	}

	if hc.AuthenticationMethod == AuthMethodBasic || hc.AuthenticationMethod == AuthMethodForms {
		if hc.Username == "" {
			errors = append(errors, "Username is required for authentication")
		}
		if hc.Password == "" {
			errors = append(errors, "Password is required for authentication")
		}
	}

	if hc.UpdateMechanism == UpdateMechanismScript && hc.UpdateScriptPath == "" {
		errors = append(errors, "Update script path is required when using script update mechanism")
	}

	if hc.BackupInterval < 1 {
		errors = append(errors, "Backup interval must be at least 1 day")
	}

	if hc.BackupRetention < 1 {
		errors = append(errors, "Backup retention must be at least 1 day")
	}

	return errors
}
