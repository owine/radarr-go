package notifications

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/radarr/radarr-go/internal/logger"
)

// DefaultTemplateEngine implements the TemplateEngine interface
type DefaultTemplateEngine struct {
	logger *logger.Logger
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(logger *logger.Logger) *DefaultTemplateEngine {
	return &DefaultTemplateEngine{
		logger: logger,
	}
}

// RenderTemplate renders a template with the given context
func (e *DefaultTemplateEngine) RenderTemplate(
	template *NotificationTemplate,
	message *NotificationMessage,
) (*RenderedTemplate, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	context := e.buildTemplateContext(message)

	rendered := &RenderedTemplate{
		Subject: e.renderText(template.Subject, context),
		Body:    e.renderText(template.Body, context),
	}

	if template.BodyHTML != "" {
		rendered.BodyHTML = e.renderText(template.BodyHTML, context)
	}

	return rendered, nil
}

// GetAvailableVariables returns all available template variables for an event type
func (e *DefaultTemplateEngine) GetAvailableVariables(eventType string) map[string]string {
	variables := map[string]string{
		// Basic variables
		"{eventType}": "Type of event (grab, download, etc.)",
		"{subject}":   "Notification subject",
		"{message}":   "Notification message body",
		"{timestamp}": "Event timestamp",
		"{server}":    "Server name",
		"{serverUrl}": "Server URL",

		// Movie variables
		"{movie.title}":         "Movie title",
		"{movie.year}":          "Movie release year",
		"{movie.tmdbId}":        "TMDb ID",
		"{movie.imdbId}":        "IMDb ID",
		"{movie.status}":        "Movie status",
		"{movie.overview}":      "Movie overview/plot",
		"{movie.runtime}":       "Movie runtime in minutes",
		"{movie.genres}":        "Movie genres (comma-separated)",
		"{movie.certification}": "Movie certification/rating",
		"{movie.studio}":        "Movie studio",
		"{movie.path}":          "Movie folder path",

		// File variables
		"{file.path}":          "File path",
		"{file.relativePath}":  "Relative file path",
		"{file.size}":          "File size",
		"{file.sizeFormatted}": "File size (human readable)",

		// Quality variables
		"{quality.name}":       "Quality profile name",
		"{quality.source}":     "Quality source",
		"{quality.resolution}": "Quality resolution",

		// Event-specific variables
		"{downloadClient}": "Download client name",
		"{downloadId}":     "Download ID",
		"{sourceTitle}":    "Source release title",
		"{isUpgrade}":      "Whether this is a quality upgrade",

		// Health variables
		"{health.type}":    "Health check type",
		"{health.message}": "Health check message",
		"{health.status}":  "Health check status",
		"{health.wikiUrl}": "Health check wiki URL",
	}

	// Add event-specific variables
	switch eventType {
	case "grab":
		variables["{indexer}"] = "Indexer name"
		variables["{releaseGroup}"] = "Release group"
	case "download", "upgrade":
		variables["{previousQuality}"] = "Previous quality (for upgrades)"
		variables["{importedFiles}"] = "Number of imported files"
	case "rename":
		variables["{oldPath}"] = "Old file path"
		variables["{newPath}"] = "New file path"
	case "movieDelete", "movieFileDelete":
		variables["{deletedFiles}"] = "List of deleted files"
		variables["{deleteReason}"] = "Reason for deletion"
	}

	return variables
}

// ValidateTemplate validates a template for syntax errors
func (e *DefaultTemplateEngine) ValidateTemplate(template *NotificationTemplate) error {
	if template == nil {
		return fmt.Errorf("template is nil")
	}

	if template.Subject == "" {
		return fmt.Errorf("template subject is required")
	}

	if template.Body == "" {
		return fmt.Errorf("template body is required")
	}

	// Validate variables in subject and body
	if err := e.validateVariables(template.Subject); err != nil {
		return fmt.Errorf("invalid variables in subject: %w", err)
	}

	if err := e.validateVariables(template.Body); err != nil {
		return fmt.Errorf("invalid variables in body: %w", err)
	}

	if template.BodyHTML != "" {
		if err := e.validateVariables(template.BodyHTML); err != nil {
			return fmt.Errorf("invalid variables in HTML body: %w", err)
		}
	}

	return nil
}

// buildTemplateContext builds the context for template rendering
func (e *DefaultTemplateEngine) buildTemplateContext(message *NotificationMessage) map[string]string {
	context := make(map[string]string)

	e.addBasicVariables(context, message)
	e.addMovieVariables(context, message)
	e.addFileVariables(context, message)
	e.addQualityVariables(context, message)
	e.addEventSpecificVariables(context, message)
	e.addHealthVariables(context, message)
	e.addCustomDataVariables(context, message)

	return context
}

// addBasicVariables adds basic notification variables to the context
func (e *DefaultTemplateEngine) addBasicVariables(context map[string]string, message *NotificationMessage) {
	context["{eventType}"] = message.EventType
	context["{subject}"] = message.Subject
	context["{message}"] = message.Body
	context["{timestamp}"] = message.Timestamp.Format("2006-01-02 15:04:05")
	context["{server}"] = message.ServerName
	context["{serverUrl}"] = message.ServerURL
	context["{isTest}"] = fmt.Sprintf("%t", message.IsTest)
}

// addMovieVariables adds movie-related variables to the context
func (e *DefaultTemplateEngine) addMovieVariables(context map[string]string, message *NotificationMessage) {
	if message.Movie == nil {
		return
	}

	context["{movie.title}"] = message.Movie.Title
	context["{movie.year}"] = fmt.Sprintf("%d", message.Movie.Year)
	context["{movie.tmdbId}"] = fmt.Sprintf("%d", message.Movie.TmdbID)
	context["{movie.imdbId}"] = message.Movie.ImdbID
	context["{movie.status}"] = string(message.Movie.Status)
	context["{movie.overview}"] = message.Movie.Overview
	context["{movie.runtime}"] = fmt.Sprintf("%d", message.Movie.Runtime)
	context["{movie.certification}"] = message.Movie.Certification
	context["{movie.studio}"] = message.Movie.Studio
	context["{movie.path}"] = message.Movie.Path

	if len(message.Movie.Genres) > 0 {
		context["{movie.genres}"] = strings.Join(message.Movie.Genres, ", ")
	} else {
		context["{movie.genres}"] = ""
	}
}

// addFileVariables adds file-related variables to the context
func (e *DefaultTemplateEngine) addFileVariables(context map[string]string, message *NotificationMessage) {
	if message.MovieFile == nil {
		return
	}

	context["{file.path}"] = message.MovieFile.RelativePath
	context["{file.relativePath}"] = message.MovieFile.RelativePath
	context["{file.size}"] = fmt.Sprintf("%d", message.MovieFile.Size)
	context["{file.sizeFormatted}"] = e.formatBytes(message.MovieFile.Size)
}

// addQualityVariables adds quality-related variables to the context
func (e *DefaultTemplateEngine) addQualityVariables(context map[string]string, message *NotificationMessage) {
	if message.Quality == nil {
		return
	}

	context["{quality.name}"] = message.Quality.Name
	// Add more quality fields as needed
}

// addEventSpecificVariables adds event-specific variables to the context
func (e *DefaultTemplateEngine) addEventSpecificVariables(context map[string]string, message *NotificationMessage) {
	context["{downloadClient}"] = message.DownloadClient
	context["{downloadId}"] = message.DownloadID
	context["{sourceTitle}"] = message.SourceTitle
	context["{isUpgrade}"] = fmt.Sprintf("%t", message.QualityUpgrade)
}

// addHealthVariables adds health check variables to the context
func (e *DefaultTemplateEngine) addHealthVariables(context map[string]string, message *NotificationMessage) {
	if message.HealthCheck == nil {
		return
	}

	context["{health.type}"] = message.HealthCheck.Type
	context["{health.message}"] = message.HealthCheck.Message
	context["{health.status}"] = string(message.HealthCheck.Status)
	context["{health.wikiUrl}"] = message.HealthCheck.WikiURL
}

// addCustomDataVariables adds custom data variables to the context
func (e *DefaultTemplateEngine) addCustomDataVariables(context map[string]string, message *NotificationMessage) {
	if message.Data == nil {
		return
	}

	for key, value := range message.Data {
		context[fmt.Sprintf("{data.%s}", key)] = fmt.Sprintf("%v", value)
	}
}

// renderText renders a text template with variable substitution
func (e *DefaultTemplateEngine) renderText(template string, context map[string]string) string {
	result := template

	// Replace all variables
	for variable, value := range context {
		result = strings.ReplaceAll(result, variable, value)
	}

	// Handle conditional blocks
	result = e.processConditionals(result, context)

	// Handle loops
	result = e.processLoops(result, context)

	// Clean up any remaining unreplaced variables
	result = e.cleanupUnreplacedVariables(result)

	return result
}

// validateVariables validates that all variables in the template are properly formatted
func (e *DefaultTemplateEngine) validateVariables(template string) error {
	// Find all variable patterns
	re := regexp.MustCompile(`\{[^}]+\}`)
	matches := re.FindAllString(template, -1)

	for _, match := range matches {
		// Check for basic formatting issues
		if !e.isValidVariableName(match) {
			return fmt.Errorf("invalid variable: %s", match)
		}
	}

	return nil
}

// isValidVariableName checks if a variable name is valid
func (e *DefaultTemplateEngine) isValidVariableName(variable string) bool {
	// Remove braces for validation
	name := strings.TrimPrefix(strings.TrimSuffix(variable, "}"), "{")

	// Must not be empty
	if name == "" {
		return false
	}

	// Check for valid characters (letters, numbers, dots, underscores)
	re := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	return re.MatchString(name)
}

// processConditionals handles conditional blocks in templates
func (e *DefaultTemplateEngine) processConditionals(template string, context map[string]string) string {
	// Handle {if variable}content{/if} blocks
	re := regexp.MustCompile(`\{if\s+([^}]+)\}(.*?)\{/if\}`)

	for re.MatchString(template) {
		matches := re.FindAllStringSubmatch(template, -1)
		for _, match := range matches {
			fullMatch := match[0]
			variable := "{" + strings.TrimSpace(match[1]) + "}"
			content := match[2]

			// Check if variable exists and has a value
			if value, exists := context[variable]; exists && value != "" && value != "0" && value != "false" {
				template = strings.ReplaceAll(template, fullMatch, content)
			} else {
				template = strings.ReplaceAll(template, fullMatch, "")
			}
		}
	}

	return template
}

// processLoops handles loop blocks in templates
func (e *DefaultTemplateEngine) processLoops(template string, context map[string]string) string {
	// Handle {foreach array}item content{/foreach} blocks
	// This is a simplified implementation - could be expanded
	re := regexp.MustCompile(`\{foreach\s+([^}]+)\}(.*?)\{/foreach\}`)

	for re.MatchString(template) {
		matches := re.FindAllStringSubmatch(template, -1)
		for _, match := range matches {
			fullMatch := match[0]
			arrayVar := "{" + strings.TrimSpace(match[1]) + "}"
			itemTemplate := match[2]

			// Check if the array variable exists
			if value, exists := context[arrayVar]; exists {
				// Split value by comma and create content for each item
				items := strings.Split(value, ", ")
				result := ""
				for _, item := range items {
					itemContent := strings.ReplaceAll(itemTemplate, "{item}", strings.TrimSpace(item))
					result += itemContent
				}
				template = strings.ReplaceAll(template, fullMatch, result)
			} else {
				template = strings.ReplaceAll(template, fullMatch, "")
			}
		}
	}

	return template
}

// cleanupUnreplacedVariables removes or handles unreplaced variables
func (e *DefaultTemplateEngine) cleanupUnreplacedVariables(template string) string {
	// Find unreplaced variables and replace with empty string or default value
	re := regexp.MustCompile(`\{[^}]+\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Could log unreplaced variables for debugging
		e.logger.Debug("Unreplaced template variable", "variable", match)
		return "" // Replace with empty string
	})
}

// formatBytes formats byte size into human-readable format
func (e *DefaultTemplateEngine) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// GetDefaultTemplates returns default templates for different event types
func GetDefaultTemplates() map[string]*NotificationTemplate {
	templates := make(map[string]*NotificationTemplate)

	// Add movie-related templates
	addMovieTemplates(templates)

	// Add system-related templates
	addSystemTemplates(templates)

	return templates
}

// addMovieTemplates adds movie-related notification templates
func addMovieTemplates(templates map[string]*NotificationTemplate) {
	templates["grab"] = &NotificationTemplate{
		EventType: "grab",
		Subject:   "{movie.title} ({movie.year}) - Grabbed",
		Body: "Movie '{movie.title} ({movie.year})' was grabbed from {downloadClient}.\n\n" +
			"Quality: {quality.name}\nSource: {sourceTitle}",
		DefaultSubject: "{movie.title} ({movie.year}) - Grabbed",
		DefaultBody: "Movie '{movie.title} ({movie.year})' was grabbed from {downloadClient}.\n\n" +
			"Quality: {quality.name}\nSource: {sourceTitle}",
	}

	templates["download"] = &NotificationTemplate{
		EventType: "download",
		Subject:   "{movie.title} ({movie.year}) - Downloaded",
		Body: "Movie '{movie.title} ({movie.year})' has been downloaded and imported.\n\n" +
			"Quality: {quality.name}\nFile: {file.relativePath}\nSize: {file.sizeFormatted}",
		DefaultSubject: "{movie.title} ({movie.year}) - Downloaded",
		DefaultBody: "Movie '{movie.title} ({movie.year})' has been downloaded and imported.\n\n" +
			"Quality: {quality.name}\nFile: {file.relativePath}\nSize: {file.sizeFormatted}",
	}

	templates["upgrade"] = &NotificationTemplate{
		EventType: "upgrade",
		Subject:   "{movie.title} ({movie.year}) - Upgraded",
		Body: "Movie '{movie.title} ({movie.year})' has been upgraded to better quality.\n\n" +
			"New Quality: {quality.name}\nFile: {file.relativePath}\nSize: {file.sizeFormatted}",
		DefaultSubject: "{movie.title} ({movie.year}) - Upgraded",
		DefaultBody: "Movie '{movie.title} ({movie.year})' has been upgraded to better quality.\n\n" +
			"New Quality: {quality.name}\nFile: {file.relativePath}\nSize: {file.sizeFormatted}",
	}

	templates["rename"] = &NotificationTemplate{
		EventType:      "rename",
		Subject:        "{movie.title} ({movie.year}) - Renamed",
		Body:           "Movie '{movie.title} ({movie.year})' has been renamed.\n\nFile: {file.relativePath}",
		DefaultSubject: "{movie.title} ({movie.year}) - Renamed",
		DefaultBody:    "Movie '{movie.title} ({movie.year})' has been renamed.\n\nFile: {file.relativePath}",
	}

	addMovieLifecycleTemplates(templates)
}

// addMovieLifecycleTemplates adds movie lifecycle templates
func addMovieLifecycleTemplates(templates map[string]*NotificationTemplate) {
	templates["movieAdded"] = &NotificationTemplate{
		EventType: "movieAdded",
		Subject:   "{movie.title} ({movie.year}) - Added",
		Body: "Movie '{movie.title} ({movie.year})' has been added to Radarr.\n\n" +
			"{if movie.overview}Overview: {movie.overview}\n{/if}Status: {movie.status}\nPath: {movie.path}",
		DefaultSubject: "{movie.title} ({movie.year}) - Added",
		DefaultBody: "Movie '{movie.title} ({movie.year})' has been added to Radarr.\n\n" +
			"{if movie.overview}Overview: {movie.overview}\n{/if}Status: {movie.status}\nPath: {movie.path}",
	}

	templates["movieDelete"] = &NotificationTemplate{
		EventType:      "movieDelete",
		Subject:        "{movie.title} ({movie.year}) - Deleted",
		Body:           "Movie '{movie.title} ({movie.year})' has been deleted from Radarr.",
		DefaultSubject: "{movie.title} ({movie.year}) - Deleted",
		DefaultBody:    "Movie '{movie.title} ({movie.year})' has been deleted from Radarr.",
	}

	templates["movieFileDelete"] = &NotificationTemplate{
		EventType:      "movieFileDelete",
		Subject:        "{movie.title} ({movie.year}) - File Deleted",
		Body:           "Movie file for '{movie.title} ({movie.year})' has been deleted.\n\nFile: {file.relativePath}",
		DefaultSubject: "{movie.title} ({movie.year}) - File Deleted",
		DefaultBody:    "Movie file for '{movie.title} ({movie.year})' has been deleted.\n\nFile: {file.relativePath}",
	}
}

// addSystemTemplates adds system-related notification templates
func addSystemTemplates(templates map[string]*NotificationTemplate) {
	templates["health"] = &NotificationTemplate{
		EventType: "health",
		Subject:   "Radarr Health Issue - {health.type}",
		Body: "A health issue has been detected in Radarr.\n\n" +
			"Type: {health.type}\nStatus: {health.status}\nMessage: {health.message}\n\n" +
			"{if health.wikiUrl}More info: {health.wikiUrl}{/if}",
		DefaultSubject: "Radarr Health Issue - {health.type}",
		DefaultBody: "A health issue has been detected in Radarr.\n\n" +
			"Type: {health.type}\nStatus: {health.status}\nMessage: {health.message}\n\n" +
			"{if health.wikiUrl}More info: {health.wikiUrl}{/if}",
	}

	templates["applicationUpdate"] = &NotificationTemplate{
		EventType: "applicationUpdate",
		Subject:   "Radarr Application Update Available",
		Body: "A new version of Radarr is available for update.\n\n" +
			"Current version: {server}\nPlease update when convenient.",
		DefaultSubject: "Radarr Application Update Available",
		DefaultBody: "A new version of Radarr is available for update.\n\n" +
			"Current version: {server}\nPlease update when convenient.",
	}

	templates["test"] = &NotificationTemplate{
		EventType: "test",
		Subject:   "Radarr Test Notification",
		Body: "This is a test notification from {server} to verify your " +
			"notification configuration is working correctly.\n\nSent at: {timestamp}",
		DefaultSubject: "Radarr Test Notification",
		DefaultBody: "This is a test notification from {server} to verify your " +
			"notification configuration is working correctly.\n\nSent at: {timestamp}",
	}
}
