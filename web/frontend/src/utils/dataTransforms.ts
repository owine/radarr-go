import type {
  Movie,
  QueueItem,
  History,
  CalendarEvent,
  Release,
  Collection
} from '../types/api';

// Generic transformation utilities
export const transformers = {
  // Normalize arrays by id
  normalizeById: <T extends { id: number }>(items: T[]): Record<number, T> => {
    return items.reduce((acc, item) => {
      acc[item.id] = item;
      return acc;
    }, {} as Record<number, T>);
  },

  // Denormalize from id map back to array
  denormalizeToArray: <T>(normalizedData: Record<number, T>): T[] => {
    return Object.values(normalizedData);
  },

  // Group items by a key
  groupBy: <T, K extends keyof T>(items: T[], key: K): Record<string, T[]> => {
    return items.reduce((groups, item) => {
      const groupKey = String(item[key]);
      if (!groups[groupKey]) {
        groups[groupKey] = [];
      }
      groups[groupKey].push(item);
      return groups;
    }, {} as Record<string, T[]>);
  },

  // Sort items by multiple criteria
  sortBy: <T>(items: T[], ...sortFns: ((item: T) => string | number | boolean)[]): T[] => {
    return [...items].sort((a, b) => {
      for (const sortFn of sortFns) {
        const aVal = sortFn(a);
        const bVal = sortFn(b);

        if (aVal < bVal) return -1;
        if (aVal > bVal) return 1;
      }
      return 0;
    });
  },

  // Filter items by multiple conditions
  filterBy: <T>(items: T[], ...filterFns: ((item: T) => boolean)[]): T[] => {
    return items.filter(item => filterFns.every(fn => fn(item)));
  },

  // Paginate items
  paginate: <T>(items: T[], page: number, pageSize: number) => {
    const startIndex = (page - 1) * pageSize;
    const endIndex = startIndex + pageSize;

    return {
      items: items.slice(startIndex, endIndex),
      totalPages: Math.ceil(items.length / pageSize),
      totalItems: items.length,
      currentPage: page,
      hasNextPage: endIndex < items.length,
      hasPreviousPage: page > 1,
    };
  },
};

// Movie-specific transformations
export const movieTransforms = {
  // Add computed properties to movies
  enrichMovie: (movie: Movie) => ({
    ...movie,
    sizeOnDisk: movie.movieFile?.size || 0,
    qualityString: movie.movieFile?.quality?.quality?.name || 'Unknown',
    isDownloaded: movie.hasFile,
    isMonitored: movie.monitored,
    sortTitle: movie.sortTitle || movie.title.toLowerCase(),
    year: movie.year,
    runtime: movie.runtime || 0,
    imdbRating: movie.ratings?.value || 0,
    genres: movie.genres || [],
    studio: movie.studio || 'Unknown',
    certification: movie.certification || 'Unrated',
  }),

  // Group movies by status
  groupByStatus: (movies: Movie[]) => {
    return transformers.groupBy(
      movies.map(movieTransforms.enrichMovie),
      'pathState' as keyof ReturnType<typeof movieTransforms.enrichMovie>
    );
  },

  // Get movies requiring attention
  getMoviesNeedingAttention: (movies: Movie[]) => {
    return movies.filter(movie =>
      movie.monitored &&
      !movie.hasFile &&
      movie.pathState !== 'static'
    );
  },

  // Calculate collection statistics
  calculateMovieStats: (movies: Movie[]) => {
    const enrichedMovies = movies.map(movieTransforms.enrichMovie);

    return {
      total: movies.length,
      downloaded: enrichedMovies.filter(m => m.isDownloaded).length,
      monitored: enrichedMovies.filter(m => m.isMonitored).length,
      unmonitored: enrichedMovies.filter(m => !m.isMonitored).length,
      totalSize: enrichedMovies.reduce((sum, m) => sum + m.sizeOnDisk, 0),
      averageRating: enrichedMovies.reduce((sum, m) => sum + m.imdbRating, 0) / movies.length,
      byYear: transformers.groupBy(enrichedMovies, 'year'),
      byGenre: enrichedMovies.reduce((acc, movie) => {
        movie.genres.forEach(genre => {
          acc[genre] = (acc[genre] || 0) + 1;
        });
        return acc;
      }, {} as Record<string, number>),
      byQuality: transformers.groupBy(enrichedMovies, 'qualityString'),
    };
  },

  // Search and filter movies
  searchMovies: (movies: Movie[], searchTerm: string, filters: {
    genre?: string;
    year?: number;
    monitored?: boolean;
    hasFile?: boolean;
    qualityProfileId?: number;
    tags?: number[];
  } = {}) => {
    let filtered = movies;

    // Text search
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      filtered = filtered.filter(movie =>
        movie.title.toLowerCase().includes(term) ||
        movie.originalTitle?.toLowerCase().includes(term) ||
        movie.overview?.toLowerCase().includes(term) ||
        movie.studio?.toLowerCase().includes(term) ||
        movie.genres.some(genre => genre.toLowerCase().includes(term))
      );
    }

    // Apply filters
    if (filters.genre) {
      filtered = filtered.filter(movie => movie.genres.includes(filters.genre!));
    }

    if (filters.year) {
      filtered = filtered.filter(movie => movie.year === filters.year);
    }

    if (filters.monitored !== undefined) {
      filtered = filtered.filter(movie => movie.monitored === filters.monitored);
    }

    if (filters.hasFile !== undefined) {
      filtered = filtered.filter(movie => movie.hasFile === filters.hasFile);
    }

    if (filters.qualityProfileId) {
      filtered = filtered.filter(movie => movie.qualityProfileId === filters.qualityProfileId);
    }

    if (filters.tags && filters.tags.length > 0) {
      filtered = filtered.filter(movie =>
        filters.tags!.some(tagId => movie.tags.includes(tagId))
      );
    }

    return filtered;
  },
};

// Queue-specific transformations
export const queueTransforms = {
  // Enrich queue items with computed properties
  enrichQueueItem: (item: QueueItem) => ({
    ...item,
    progressPercent: item.size > 0 ? ((item.size - item.sizeleft) / item.size) * 100 : 0,
    downloadSpeed: item.timeleft ? (item.sizeleft / parseFloat(item.timeleft)) : 0,
    eta: item.timeleft ? new Date(Date.now() + parseFloat(item.timeleft) * 1000) : null,
    isStalled: item.status === 'downloading' && (!item.timeleft || parseFloat(item.timeleft) > 86400),
    hasErrors: !!(item.errorMessage || (item.statusMessages && item.statusMessages.length > 0)),
    isCompleted: item.status === 'completed',
    isPaused: item.status === 'paused',
    isDownloading: item.status === 'downloading',
  }),

  // Group queue items by status
  groupByStatus: (items: QueueItem[]) => {
    return transformers.groupBy(
      items.map(queueTransforms.enrichQueueItem),
      'status' as keyof ReturnType<typeof queueTransforms.enrichQueueItem>
    );
  },

  // Calculate queue statistics
  calculateQueueStats: (items: QueueItem[]) => {
    const enriched = items.map(queueTransforms.enrichQueueItem);

    return {
      total: items.length,
      downloading: enriched.filter(i => i.isDownloading).length,
      completed: enriched.filter(i => i.isCompleted).length,
      paused: enriched.filter(i => i.isPaused).length,
      stalled: enriched.filter(i => i.isStalled).length,
      withErrors: enriched.filter(i => i.hasErrors).length,
      totalSize: enriched.reduce((sum, i) => sum + i.size, 0),
      remainingSize: enriched.reduce((sum, i) => sum + i.sizeleft, 0),
      averageProgress: enriched.reduce((sum, i) => sum + i.progressPercent, 0) / items.length,
      estimatedTimeRemaining: enriched
        .filter(i => i.eta)
        .reduce((latest, i) => i.eta && (!latest || i.eta > latest) ? i.eta : latest, null as Date | null),
    };
  },
};

// Calendar-specific transformations
export const calendarTransforms = {
  // Enrich calendar events
  enrichCalendarEvent: (event: CalendarEvent) => ({
    ...event,
    isPast: event.airDate ? new Date(event.airDate) < new Date() : false,
    isToday: event.airDate ?
      new Date(event.airDate).toDateString() === new Date().toDateString() : false,
    isThisWeek: event.airDate ? isWithinDays(new Date(event.airDate), 7) : false,
    isThisMonth: event.airDate ? isWithinDays(new Date(event.airDate), 30) : false,
    displayDate: event.airDate || event.physicalRelease || event.inCinemas || event.digitalRelease,
    eventTypeDisplay: getEventTypeDisplay(event.eventType),
  }),

  // Group events by date
  groupByDate: (events: CalendarEvent[]) => {
    const enriched = events.map(calendarTransforms.enrichCalendarEvent);
    return transformers.groupBy(enriched, 'displayDate' as keyof typeof enriched[0]);
  },

  // Filter events by date range
  filterByDateRange: (events: CalendarEvent[], startDate: Date, endDate: Date) => {
    return events.filter(event => {
      const eventDate = new Date(event.airDate || event.physicalRelease || '');
      return eventDate >= startDate && eventDate <= endDate;
    });
  },
};

// History-specific transformations
export const historyTransforms = {
  // Enrich history records
  enrichHistory: (record: History) => ({
    ...record,
    displayDate: new Date(record.date).toLocaleDateString(),
    displayTime: new Date(record.date).toLocaleTimeString(),
    eventTypeDisplay: getEventTypeDisplay(record.eventType),
    wasSuccessful: ['movieFileImported', 'downloadFolderImported', 'grabbed'].includes(record.eventType),
    wasFailed: ['downloadFailed', 'importFailed'].includes(record.eventType),
    wasDeleted: ['movieFileDeleted', 'movieFileRenamed'].includes(record.eventType),
  }),

  // Group history by event type
  groupByEventType: (records: History[]) => {
    return transformers.groupBy(
      records.map(historyTransforms.enrichHistory),
      'eventType' as keyof ReturnType<typeof historyTransforms.enrichHistory>
    );
  },

  // Calculate history statistics
  calculateHistoryStats: (records: History[]) => {
    const enriched = records.map(historyTransforms.enrichHistory);

    return {
      total: records.length,
      successful: enriched.filter(r => r.wasSuccessful).length,
      failed: enriched.filter(r => r.wasFailed).length,
      deleted: enriched.filter(r => r.wasDeleted).length,
      byEventType: transformers.groupBy(enriched, 'eventType'),
      recentActivity: enriched
        .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
        .slice(0, 10),
    };
  },
};

// Collection transformations
export const collectionTransforms = {
  // Enrich collection with computed properties
  enrichCollection: (collection: Collection) => ({
    ...collection,
    movieCount: collection.movies?.length || 0,
    downloadedCount: collection.movies?.filter(m => m.hasFile).length || 0,
    monitoredCount: collection.movies?.filter(m => m.monitored).length || 0,
    completionPercentage: collection.movies?.length ?
      ((collection.movies.filter(m => m.hasFile).length / collection.movies.length) * 100) : 0,
    totalSize: collection.movies?.reduce((sum, movie) => sum + (movie.movieFile?.size || 0), 0) || 0,
    genres: Array.from(new Set(collection.movies?.flatMap(m => m.genres) || [])),
    averageRating: collection.movies?.length ?
      collection.movies.reduce((sum, m) => sum + (m.ratings?.value || 0), 0) / collection.movies.length : 0,
  }),
};

// Release transformations
export const releaseTransforms = {
  // Enrich release with computed properties
  enrichRelease: (release: Release) => ({
    ...release,
    ageInDays: release.age,
    ageDisplay: getAgeDisplay(release.age),
    sizeDisplay: formatFileSize(release.size),
    qualityDisplay: release.quality?.quality?.name || 'Unknown',
    isApproved: release.approved,
    isRejected: release.rejected || release.temporarilyRejected,
    rejectionReasons: release.rejections || [],
    seedersDisplay: release.seeders ? `${release.seeders} seeders` : 'Unknown',
    leechersDisplay: release.leechers ? `${release.leechers} leechers` : 'Unknown',
    protocolIcon: getProtocolIcon(release.protocol),
  }),

  // Filter releases by quality
  filterByQuality: (releases: Release[], minQuality: number, maxQuality: number) => {
    return releases.filter(release => {
      const qualityId = release.quality?.quality?.id || 0;
      return qualityId >= minQuality && qualityId <= maxQuality;
    });
  },

  // Sort releases by preference
  sortByPreference: (releases: Release[]) => {
    return transformers.sortBy(
      releases,
      (r) => r.approved ? 0 : 1, // Approved first
      (r) => -r.qualityWeight, // Higher quality weight first
      (r) => r.age, // Newer releases first
      (r) => -r.size, // Larger releases first
      (r) => -(r.seeders || 0) // More seeders first
    );
  },
};

// Utility functions
function isWithinDays(date: Date, days: number): boolean {
  const now = new Date();
  const diffTime = Math.abs(date.getTime() - now.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return diffDays <= days;
}

function getEventTypeDisplay(eventType: string): string {
  const eventTypeMap: Record<string, string> = {
    'grabbed': 'Grabbed',
    'movieFileImported': 'Imported',
    'downloadFailed': 'Download Failed',
    'importFailed': 'Import Failed',
    'movieFileDeleted': 'File Deleted',
    'movieFileRenamed': 'File Renamed',
    'downloadFolderImported': 'Folder Imported',
    'physicalRelease': 'Physical Release',
    'inCinemas': 'In Cinemas',
    'digitalRelease': 'Digital Release',
  };

  return eventTypeMap[eventType] || eventType;
}

function getAgeDisplay(ageInDays: number): string {
  if (ageInDays < 1) {
    const hours = Math.floor(ageInDays * 24);
    return `${hours}h`;
  } else if (ageInDays < 7) {
    return `${Math.floor(ageInDays)}d`;
  } else if (ageInDays < 30) {
    return `${Math.floor(ageInDays / 7)}w`;
  } else {
    return `${Math.floor(ageInDays / 30)}m`;
  }
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function getProtocolIcon(protocol: string): string {
  const icons: Record<string, string> = {
    'torrent': '🧲',
    'usenet': '📡',
    'http': '🌐',
  };

  return icons[protocol.toLowerCase()] || '📄';
}

// Export all transformations as a single object
export const dataTransforms = {
  ...transformers,
  movie: movieTransforms,
  queue: queueTransforms,
  calendar: calendarTransforms,
  history: historyTransforms,
  collection: collectionTransforms,
  release: releaseTransforms,
};
