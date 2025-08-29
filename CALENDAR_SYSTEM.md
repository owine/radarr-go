# Calendar and Scheduling System

The Radarr Go calendar system provides comprehensive calendar functionality for tracking movie release dates and events, with full API compatibility with Radarr v3 and additional features for external calendar integration.

## Features

### Core Calendar Functionality
- **Movie Release Tracking**: Automatically tracks cinema, physical, and digital release dates
- **Event Generation**: Creates calendar events from movie metadata and release information
- **Multiple View Types**: Support for month, week, agenda, and forecast views
- **Filtering**: Filter events by movie, tags, monitored status, date range, and event type
- **Real-time Updates**: Events update automatically based on movie status changes

### iCal Feed Generation
- **RFC 5545 Compliant**: Generates standards-compliant iCal feeds for external calendar applications
- **Configurable Content**: Customize event titles, descriptions, and included event types
- **Authentication Support**: Optional passkey protection for feed access
- **Flexible Time Ranges**: Configure how far in the past and future to include events
- **Tag Filtering**: Include only movies with specific tags in the feed

### Performance Optimization
- **Event Caching**: Intelligent caching of calendar events with configurable expiration
- **Efficient Queries**: Optimized database queries with proper indexing
- **Lazy Loading**: Events generated on-demand to reduce memory usage
- **Cache Management**: Automatic cleanup of expired cache entries

## API Endpoints

### Calendar Events
- `GET /api/v3/calendar` - Retrieve calendar events with filtering
- `GET /api/v3/calendar/stats` - Get calendar statistics and summary
- `POST /api/v3/calendar/refresh` - Force refresh calendar events and clear cache

### iCal Feed
- `GET /api/v3/calendar/feed.ics` - Generate iCal feed for external calendar apps
- `GET /api/v3/calendar/feed/url` - Generate iCal feed URL with parameters

### Configuration
- `GET /api/v3/calendar/config` - Get calendar configuration settings
- `PUT /api/v3/calendar/config` - Update calendar configuration

## Calendar Event Types

The system supports several types of calendar events:

- **Cinema Release** (`cinemaRelease`) - Movie released in theaters
- **Physical Release** (`physicalRelease`) - Physical media (DVD/Blu-ray) release
- **Digital Release** (`digitalRelease`) - Digital/streaming platform release
- **Availability** (`availability`) - When movie becomes available for download based on minimum availability settings
- **Announcement** (`announcement`) - Movie announcement events
- **Monitoring** (`monitoring`) - Monitoring status change events

## API Usage Examples

### Get Calendar Events
```bash
# Get events for current month
curl "http://localhost:7878/api/v3/calendar"

# Get events for specific date range
curl "http://localhost:7878/api/v3/calendar?start=2023-12-01&end=2023-12-31"

# Get only cinema releases for monitored movies
curl "http://localhost:7878/api/v3/calendar?eventTypes=cinemaRelease&monitored=true"

# Get events for specific movies
curl "http://localhost:7878/api/v3/calendar?movieIds=1,2,3"

# Get events with movie information included
curl "http://localhost:7878/api/v3/calendar?includeMovieInformation=true"
```

### iCal Feed Examples
```bash
# Basic iCal feed
curl "http://localhost:7878/api/v3/calendar/feed.ics"

# Customized feed with authentication
curl "http://localhost:7878/api/v3/calendar/feed.ics?eventTypes=cinemaRelease,physicalRelease&daysInFuture=180&passKey=secret123"

# Feed for specific tags only
curl "http://localhost:7878/api/v3/calendar/feed.ics?tags=1,2,3"

# Feed including only monitored movies
curl "http://localhost:7878/api/v3/calendar/feed.ics?monitored=true"
```

### Get Calendar Statistics
```bash
curl "http://localhost:7878/api/v3/calendar/stats"
```

Response includes:
- Total movies count
- Monitored vs unmonitored movies
- Movies with/without files
- Upcoming releases in next 30 days
- Event counts by type

## Configuration Options

### Calendar Display Settings
- **Default View**: Month, week, agenda, or forecast view
- **First Day of Week**: Sunday (0) through Saturday (6)
- **Colored Events**: Enable color-coding for different event types
- **Movie Information**: Include detailed movie info in event tooltips
- **Event Filtering**: Enable full calendar event filtering
- **Multiple Events**: Collapse multiple events for same movie on same day

### iCal Feed Settings
- **Enable Feed**: Enable/disable iCal feed generation
- **Authentication**: Require passkey for feed access
- **Time Range**: Configure days in past and future to include
- **Event Types**: Default event types to include in feeds
- **Tags**: Default tags to filter by
- **Format Templates**: Customize event title and description formats

### Caching Settings
- **Enable Caching**: Enable/disable event caching
- **Cache Duration**: How long to cache events (in minutes)
- **Cache Cleanup**: Automatic cleanup of expired cache entries

## Database Schema

### Calendar Events Table
Stores generated calendar events with full movie context:
- Event metadata (type, date, status, description)
- Movie information (title, year, overview, images)
- Display settings (all-day, location, reminders)
- Relationships to movies table

### Calendar Configuration Table
Stores calendar settings and preferences:
- Display preferences (view type, colors, information level)
- iCal feed configuration (authentication, time ranges, formats)
- Caching settings (enabled, duration)
- Event type filters and defaults

### Calendar Event Cache Table
Performance optimization cache:
- Cached event data with expiration
- Summary statistics
- Cache key management
- Automatic cleanup

## Performance Characteristics

### Event Generation
- Events generated on-demand from movie data
- Efficient filtering at database level
- Minimal memory footprint through streaming

### Caching Strategy
- Intelligent cache keys based on request parameters
- Configurable cache expiration (default 1 hour)
- Automatic cleanup of expired entries
- Cache bypass for real-time updates

### Database Optimization
- Proper indexing on event dates, movie IDs, and types
- Foreign key constraints for data integrity
- Optimized queries with minimal JOIN operations
- Support for both PostgreSQL and MySQL/MariaDB

## Integration with Radarr Features

### Movie Management
- Events automatically update when movie metadata changes
- Release date changes trigger event regeneration
- Monitoring status affects event visibility

### Notification System
- Calendar events can trigger notifications
- Upcoming release notifications
- Availability change notifications

### Task System
- Calendar refresh tasks for periodic updates
- Cache cleanup tasks for maintenance
- Event generation tasks for new movies

## External Calendar Integration

### Supported Applications
The iCal feed works with:
- Google Calendar
- Apple Calendar (macOS/iOS)
- Outlook/Office 365
- Thunderbird
- Any RFC 5545 compliant calendar application

### Setup Examples

#### Google Calendar
1. Get the iCal feed URL from `/api/v3/calendar/feed/url`
2. In Google Calendar, go to "Other calendars" → "Add by URL"
3. Paste the iCal feed URL
4. Calendar updates automatically (typically every 8-12 hours)

#### Apple Calendar
1. Open Calendar app
2. File → New Calendar Subscription
3. Enter the iCal feed URL
4. Configure refresh frequency and other settings

#### Outlook
1. Go to Calendar section
2. Add Calendar → From Internet
3. Enter the iCal feed URL
4. Customize display name and other settings

## Monitoring and Maintenance

### Health Checks
The calendar system includes health monitoring:
- Database connectivity checks
- Cache performance monitoring
- Event generation validation
- Feed generation status

### Maintenance Tasks
Regular maintenance includes:
- Cache cleanup (removes expired entries)
- Event refresh (updates from movie changes)
- Configuration validation
- Performance metric collection

### Troubleshooting

#### Common Issues
1. **Events not appearing**: Check movie monitoring status and date filters
2. **iCal feed not updating**: Verify cache settings and external app refresh intervals
3. **Performance issues**: Review cache configuration and database indexing
4. **Authentication failures**: Verify passkey configuration in feed settings

#### Debug Information
Enable debug logging to see:
- Event generation queries
- Cache hit/miss ratios
- iCal feed generation details
- API request parameters

## Future Enhancements

### Planned Features
- Custom event creation for manual tracking
- Event templates for recurring patterns
- Advanced filtering with complex queries
- Integration with external metadata sources
- Mobile-optimized calendar views

### API Extensions
- WebSocket support for real-time updates
- Bulk event operations
- Event history and audit trail
- Advanced analytics and reporting

## Security Considerations

### Authentication
- Optional passkey protection for iCal feeds
- Integration with Radarr's existing API key system
- No sensitive information exposed in calendar events

### Data Privacy
- Events contain only public movie information
- No user-specific data in calendar feeds
- Configurable information levels for events

### Access Control
- Same access controls as Radarr API
- Calendar-specific permissions
- Feed-level access controls
