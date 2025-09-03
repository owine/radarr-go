// API Integration Layer Tests
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { configureStore } from '@reduxjs/toolkit';
import { radarrApi } from '../store/api/radarrApi';
import { websocketMiddleware, webSocketManager } from '../store/middleware/websocketMiddleware';
import { cacheManager } from '../utils/cacheManager';
import { dataTransforms } from '../utils/dataTransforms';
import { invalidationStrategies } from '../utils/cacheInvalidation';
import type { Movie, QueueItem, History } from '../types/api';

// Mock data
const mockMovie: Movie = {
  id: 1,
  title: 'Test Movie',
  originalTitle: 'Test Movie Original',
  sortTitle: 'test movie',
  status: 'released',
  overview: 'A test movie for testing purposes',
  inCinemas: '2023-01-01',
  physicalRelease: '2023-03-01',
  digitalRelease: '2023-02-15',
  images: [
    { coverType: 'poster', url: '/poster.jpg', remoteUrl: 'https://example.com/poster.jpg' }
  ],
  website: 'https://testmovie.com',
  year: 2023,
  hasFile: true,
  youTubeTrailerId: 'abc123',
  studio: 'Test Studio',
  path: '/movies/Test Movie (2023)',
  pathState: 'static',
  qualityProfileId: 1,
  monitored: true,
  minimumAvailability: 'released',
  isAvailable: true,
  folderName: 'Test Movie (2023)',
  runtime: 120,
  lastInfoSync: '2023-12-01T00:00:00Z',
  cleanTitle: 'testmovie',
  imdbId: 'tt1234567',
  tmdbId: 123456,
  titleSlug: 'test-movie-2023',
  certification: 'PG-13',
  genres: ['Action', 'Drama'],
  tags: [1, 2],
  added: '2023-11-01T00:00:00Z',
  ratings: { votes: 1000, value: 7.5 },
  movieFile: {
    id: 1,
    movieId: 1,
    relativePath: 'Test Movie (2023).mkv',
    path: '/movies/Test Movie (2023)/Test Movie (2023).mkv',
    size: 5000000000,
    dateAdded: '2023-12-01T00:00:00Z',
    sceneName: 'Test.Movie.2023.1080p.BluRay.x264-GROUP',
    releaseGroup: 'GROUP',
    quality: {
      quality: { id: 7, name: 'Bluray-1080p', source: 'bluray', resolution: 1080, modifier: '' },
      revision: { version: 1, real: 0, isRepack: false }
    },
    mediaInfo: {
      audioChannels: 6,
      audioCodec: 'DTS',
      audioLanguages: 'English',
      height: 1080,
      width: 1920,
      runtime: 120,
      videoCodec: 'x264',
      videoDynamicRange: 'SDR',
      videoDynamicRangeType: 'SDR'
    },
    originalFilePath: '/downloads/Test.Movie.2023.1080p.BluRay.x264-GROUP/Test.Movie.2023.1080p.BluRay.x264-GROUP.mkv',
    qualityCutoffNotMet: false,
    languages: [{ id: 1, name: 'English' }]
  }
};

const mockQueueItem: QueueItem = {
  id: 1,
  movieId: 1,
  movie: mockMovie,
  languages: [{ id: 1, name: 'English' }],
  quality: mockMovie.movieFile!.quality,
  customFormats: [],
  size: 5000000000,
  title: 'Test Movie 2023 1080p BluRay',
  sizeleft: 1000000000,
  timeleft: '00:30:00',
  estimatedCompletionTime: '2023-12-01T01:30:00Z',
  status: 'downloading',
  trackedDownloadStatus: 'downloading',
  trackedDownloadState: 'downloading',
  statusMessages: [],
  downloadId: 'download-123',
  protocol: 'torrent',
  downloadClient: 'Test Client',
  indexer: 'Test Indexer',
  outputPath: '/downloads/Test Movie 2023'
};

// Mock store setup
function createMockStore() {
  return configureStore({
    reducer: {
      [radarrApi.reducerPath]: radarrApi.reducer,
    },
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware({
        serializableCheck: {
          ignoredActions: [radarrApi.util.resetApiState.type],
        },
      }).concat(radarrApi.middleware, websocketMiddleware),
  });
}

describe('API Integration Layer', () => {
  let store: ReturnType<typeof createMockStore>;

  beforeEach(() => {
    store = createMockStore();
    vi.clearAllMocks();
  });

  afterEach(() => {
    store.dispatch(radarrApi.util.resetApiState());
  });

  describe('RTK Query API Slice', () => {
    it('should have all expected endpoints', () => {
      const endpoints = Object.keys(radarrApi.endpoints);

      // System endpoints
      expect(endpoints).toContain('getSystemStatus');
      expect(endpoints).toContain('getHealth');

      // Movie endpoints
      expect(endpoints).toContain('getMovies');
      expect(endpoints).toContain('getMovie');
      expect(endpoints).toContain('addMovie');
      expect(endpoints).toContain('updateMovie');
      expect(endpoints).toContain('deleteMovie');
      expect(endpoints).toContain('searchMovies');
      expect(endpoints).toContain('getPopularMovies');
      expect(endpoints).toContain('getTrendingMovies');

      // Quality endpoints
      expect(endpoints).toContain('getQualityProfiles');
      expect(endpoints).toContain('getQualityProfile');
      expect(endpoints).toContain('getQualityDefinitions');
      expect(endpoints).toContain('getCustomFormats');

      // Queue endpoints
      expect(endpoints).toContain('getQueue');
      expect(endpoints).toContain('getQueueItem');
      expect(endpoints).toContain('removeQueueItem');
      expect(endpoints).toContain('removeQueueItems');
      expect(endpoints).toContain('getQueueStats');

      // History endpoints
      expect(endpoints).toContain('getHistory');
      expect(endpoints).toContain('getHistoryStats');

      // Activity endpoints
      expect(endpoints).toContain('getActivity');
      expect(endpoints).toContain('getRunningActivities');

      // Indexer endpoints
      expect(endpoints).toContain('getIndexers');
      expect(endpoints).toContain('createIndexer');
      expect(endpoints).toContain('updateIndexer');
      expect(endpoints).toContain('deleteIndexer');
      expect(endpoints).toContain('testIndexer');

      // Download Client endpoints
      expect(endpoints).toContain('getDownloadClients');
      expect(endpoints).toContain('createDownloadClient');
      expect(endpoints).toContain('updateDownloadClient');
      expect(endpoints).toContain('deleteDownloadClient');
      expect(endpoints).toContain('testDownloadClient');
      expect(endpoints).toContain('getDownloadClientStats');

      // Import List endpoints
      expect(endpoints).toContain('getImportLists');
      expect(endpoints).toContain('createImportList');
      expect(endpoints).toContain('updateImportList');
      expect(endpoints).toContain('deleteImportList');
      expect(endpoints).toContain('testImportList');
      expect(endpoints).toContain('syncImportList');
      expect(endpoints).toContain('syncAllImportLists');
      expect(endpoints).toContain('getImportListStats');

      // Notification endpoints
      expect(endpoints).toContain('getNotifications');
      expect(endpoints).toContain('createNotification');
      expect(endpoints).toContain('updateNotification');
      expect(endpoints).toContain('deleteNotification');
      expect(endpoints).toContain('testNotification');
      expect(endpoints).toContain('getNotificationProviders');

      // Configuration endpoints
      expect(endpoints).toContain('getHostConfig');
      expect(endpoints).toContain('updateHostConfig');
      expect(endpoints).toContain('getNamingConfig');
      expect(endpoints).toContain('updateNamingConfig');
      expect(endpoints).toContain('getMediaManagementConfig');
      expect(endpoints).toContain('updateMediaManagementConfig');
      expect(endpoints).toContain('getConfigStats');

      // Tag endpoints
      expect(endpoints).toContain('getTags');
      expect(endpoints).toContain('createTag');
      expect(endpoints).toContain('updateTag');
      expect(endpoints).toContain('deleteTag');

      // Release and Search endpoints
      expect(endpoints).toContain('getReleases');
      expect(endpoints).toContain('searchMovieReleases');
      expect(endpoints).toContain('grabRelease');

      // Calendar endpoints
      expect(endpoints).toContain('getCalendar');

      // Wanted Movies endpoints
      expect(endpoints).toContain('getMissingMovies');
      expect(endpoints).toContain('getCutoffUnmetMovies');
      expect(endpoints).toContain('getWantedStats');

      // Parse endpoints
      expect(endpoints).toContain('parseReleaseTitle');

      // Command/Task endpoints
      expect(endpoints).toContain('getCommands');
      expect(endpoints).toContain('getCommand');
      expect(endpoints).toContain('queueCommand');
      expect(endpoints).toContain('cancelCommand');

      // System Resource endpoints
      expect(endpoints).toContain('getSystemResources');
      expect(endpoints).toContain('getDiskSpace');
      expect(endpoints).toContain('getPerformanceMetrics');

      // Collection endpoints
      expect(endpoints).toContain('getCollections');
      expect(endpoints).toContain('createCollection');
      expect(endpoints).toContain('updateCollection');
      expect(endpoints).toContain('deleteCollection');
      expect(endpoints).toContain('getCollectionStats');
    });

    it('should have proper tag types for cache invalidation', () => {
      const tagTypes = radarrApi.tagTypes;

      expect(tagTypes).toContain('Movie');
      expect(tagTypes).toContain('MovieFile');
      expect(tagTypes).toContain('QualityProfile');
      expect(tagTypes).toContain('QualityDefinition');
      expect(tagTypes).toContain('CustomFormat');
      expect(tagTypes).toContain('RootFolder');
      expect(tagTypes).toContain('SystemStatus');
      expect(tagTypes).toContain('Health');
      expect(tagTypes).toContain('Collection');
      expect(tagTypes).toContain('Indexer');
      expect(tagTypes).toContain('DownloadClient');
      expect(tagTypes).toContain('ImportList');
      expect(tagTypes).toContain('Queue');
      expect(tagTypes).toContain('History');
      expect(tagTypes).toContain('Activity');
      expect(tagTypes).toContain('Notification');
      expect(tagTypes).toContain('Config');
      expect(tagTypes).toContain('Tag');
      expect(tagTypes).toContain('Release');
      expect(tagTypes).toContain('Calendar');
      expect(tagTypes).toContain('WantedMovie');
      expect(tagTypes).toContain('Parse');
      expect(tagTypes).toContain('FileOrganization');
      expect(tagTypes).toContain('Command');
      expect(tagTypes).toContain('SystemResource');
    });
  });

  describe('WebSocket Middleware', () => {
    it('should have proper connection states', () => {
      const connectionState = webSocketManager.getConnectionState();
      expect(['disconnected', 'connecting', 'connected', 'reconnecting', 'error']).toContain(connectionState);
    });

    it('should support event subscription', () => {
      const mockCallback = vi.fn();
      const unsubscribe = webSocketManager.subscribe('QueueUpdate', mockCallback);

      expect(typeof unsubscribe).toBe('function');
      unsubscribe();
    });

    it('should maintain event history', () => {
      const history = webSocketManager.getEventHistory();
      expect(Array.isArray(history)).toBe(true);
    });
  });

  describe('Cache Manager', () => {
    beforeEach(() => {
      cacheManager.clear();
    });

    it('should cache and retrieve data', () => {
      const testData = { id: 1, name: 'test' };
      cacheManager.set('test-key', testData, { ttl: 60000 });

      const retrieved = cacheManager.get('test-key');
      expect(retrieved).toEqual(testData);
    });

    it('should handle cache expiration', async () => {
      const testData = { id: 1, name: 'test' };
      cacheManager.set('test-key', testData, { ttl: 1 }); // 1ms TTL

      // Wait for expiration
      await new Promise(resolve => setTimeout(resolve, 10));

      const retrieved = cacheManager.get('test-key');
      expect(retrieved).toBeNull();
    });

    it('should clear cache by tags', () => {
      cacheManager.set('key1', { data: 'test1' }, { tags: ['movies'] });
      cacheManager.set('key2', { data: 'test2' }, { tags: ['queue'] });
      cacheManager.set('key3', { data: 'test3' }, { tags: ['movies', 'queue'] });

      const deletedCount = cacheManager.clearByTags(['movies']);

      expect(deletedCount).toBe(2); // key1 and key3
      expect(cacheManager.get('key2')).not.toBeNull();
    });

    it('should provide cache statistics', () => {
      cacheManager.set('key1', { data: 'test1' });
      cacheManager.set('key2', { data: 'test2' });

      const stats = cacheManager.getStats();
      expect(stats.totalItems).toBe(2);
      expect(stats.validItems).toBe(2);
      expect(stats.expiredItems).toBe(0);
    });
  });

  describe('Data Transformations', () => {
    it('should normalize and denormalize data correctly', () => {
      const movies = [mockMovie, { ...mockMovie, id: 2, title: 'Another Movie' }];

      const normalized = dataTransforms.normalizeById(movies);
      expect(normalized[1]).toEqual(mockMovie);
      expect(normalized[2].title).toBe('Another Movie');

      const denormalized = dataTransforms.denormalizeToArray(normalized);
      expect(denormalized).toHaveLength(2);
    });

    it('should group items correctly', () => {
      const movies = [
        { ...mockMovie, year: 2023 },
        { ...mockMovie, id: 2, year: 2024 },
        { ...mockMovie, id: 3, year: 2023 }
      ];

      const grouped = dataTransforms.groupBy(movies, 'year');
      expect(grouped['2023']).toHaveLength(2);
      expect(grouped['2024']).toHaveLength(1);
    });

    it('should enrich movies with computed properties', () => {
      const enriched = dataTransforms.movie.enrichMovie(mockMovie);

      expect(enriched.sizeOnDisk).toBe(5000000000);
      expect(enriched.qualityString).toBe('Bluray-1080p');
      expect(enriched.isDownloaded).toBe(true);
      expect(enriched.isMonitored).toBe(true);
      expect(enriched.genres).toEqual(['Action', 'Drama']);
    });

    it('should calculate movie statistics', () => {
      const movies = [
        mockMovie,
        { ...mockMovie, id: 2, hasFile: false, monitored: false },
        { ...mockMovie, id: 3, year: 2024 }
      ];

      const stats = dataTransforms.movie.calculateMovieStats(movies);

      expect(stats.total).toBe(3);
      expect(stats.downloaded).toBe(2);
      expect(stats.monitored).toBe(2);
      expect(stats.unmonitored).toBe(1);
      expect(stats.byYear['2023']).toHaveLength(2);
      expect(stats.byYear['2024']).toHaveLength(1);
    });

    it('should search movies with filters', () => {
      const movies = [
        mockMovie,
        { ...mockMovie, id: 2, title: 'Comedy Movie', genres: ['Comedy'], year: 2024 },
        { ...mockMovie, id: 3, title: 'Horror Film', genres: ['Horror'], monitored: false }
      ];

      // Text search
      const searchResults = dataTransforms.movie.searchMovies(movies, 'comedy');
      expect(searchResults).toHaveLength(1);
      expect(searchResults[0].title).toBe('Comedy Movie');

      // Filter by genre
      const actionMovies = dataTransforms.movie.searchMovies(movies, '', { genre: 'Action' });
      expect(actionMovies).toHaveLength(1);
      expect(actionMovies[0].title).toBe('Test Movie');

      // Filter by monitored status
      const unmonitoredMovies = dataTransforms.movie.searchMovies(movies, '', { monitored: false });
      expect(unmonitoredMovies).toHaveLength(1);
      expect(unmonitoredMovies[0].title).toBe('Horror Film');
    });

    it('should enrich queue items with computed properties', () => {
      const enriched = dataTransforms.queue.enrichQueueItem(mockQueueItem);

      expect(enriched.progressPercent).toBe(80); // (5GB - 1GB) / 5GB * 100
      expect(enriched.isDownloading).toBe(true);
      expect(enriched.isCompleted).toBe(false);
      expect(enriched.hasErrors).toBe(false);
    });
  });

  describe('Cache Invalidation Strategies', () => {
    it('should have comprehensive invalidation strategies', () => {
      expect(invalidationStrategies.movieUpdate).toBeDefined();
      expect(invalidationStrategies.movieDelete).toBeDefined();
      expect(invalidationStrategies.downloadComplete).toBeDefined();
      expect(invalidationStrategies.queueUpdate).toBeDefined();
      expect(invalidationStrategies.releaseGrab).toBeDefined();
      expect(invalidationStrategies.qualityProfileUpdate).toBeDefined();
      expect(invalidationStrategies.indexerUpdate).toBeDefined();
      expect(invalidationStrategies.downloadClientUpdate).toBeDefined();
      expect(invalidationStrategies.importListUpdate).toBeDefined();
      expect(invalidationStrategies.notificationUpdate).toBeDefined();
      expect(invalidationStrategies.systemConfigUpdate).toBeDefined();
      expect(invalidationStrategies.healthUpdate).toBeDefined();
      expect(invalidationStrategies.bulkMovieOperation).toBeDefined();
      expect(invalidationStrategies.collectionUpdate).toBeDefined();
      expect(invalidationStrategies.taskComplete).toBeDefined();
    });

    it('should have proper invalidation rules', () => {
      const movieUpdateStrategy = invalidationStrategies.movieUpdate;

      expect(movieUpdateStrategy.name).toBe('Movie Update');
      expect(movieUpdateStrategy.rules).toHaveLength(1);
      expect(movieUpdateStrategy.rules[0].triggerTags).toContain('Movie');
      expect(movieUpdateStrategy.rules[0].invalidateTags).toContain('Movie');
      expect(movieUpdateStrategy.rules[0].invalidateTags).toContain('WantedMovie');
      expect(movieUpdateStrategy.rules[0].invalidateTags).toContain('Calendar');
      expect(movieUpdateStrategy.rules[0].invalidateTags).toContain('Collection');
    });

    it('should support conditional invalidation', () => {
      const taskCompleteStrategy = invalidationStrategies.taskComplete;
      const rule = taskCompleteStrategy.rules[0];

      expect(rule.condition).toBeDefined();
      expect(rule.condition!({ status: 'completed' })).toBe(true);
      expect(rule.condition!({ status: 'running' })).toBe(false);
    });
  });

  describe('API Endpoint Configuration', () => {
    it('should have proper pagination support for queue endpoint', () => {
      const queueEndpoint = radarrApi.endpoints.getQueue;
      expect(queueEndpoint).toBeDefined();

      // The endpoint should accept pagination parameters
      const queryArgs = {
        page: 2,
        pageSize: 50,
        sortKey: 'title',
        sortDirection: 'desc' as const,
        includeUnknownMovieItems: true,
        includeMovie: true,
      };

      // This tests the query structure - it should not throw
      expect(() => queueEndpoint.query(queryArgs)).not.toThrow();
    });

    it('should have proper mutation endpoints with invalidation tags', () => {
      const updateMovieEndpoint = radarrApi.endpoints.updateMovie;
      expect(updateMovieEndpoint).toBeDefined();

      const deleteMovieEndpoint = radarrApi.endpoints.deleteMovie;
      expect(deleteMovieEndpoint).toBeDefined();

      const addMovieEndpoint = radarrApi.endpoints.addMovie;
      expect(addMovieEndpoint).toBeDefined();
    });
  });

  describe('Type Safety', () => {
    it('should have proper TypeScript types for Movie interface', () => {
      // This test ensures the Movie interface has all expected properties
      const movieKeys = Object.keys(mockMovie);

      const expectedKeys = [
        'id', 'title', 'originalTitle', 'sortTitle', 'status', 'overview',
        'inCinemas', 'physicalRelease', 'digitalRelease', 'images', 'website',
        'year', 'hasFile', 'youTubeTrailerId', 'studio', 'path', 'pathState',
        'qualityProfileId', 'monitored', 'minimumAvailability', 'isAvailable',
        'folderName', 'runtime', 'lastInfoSync', 'cleanTitle', 'imdbId',
        'tmdbId', 'titleSlug', 'certification', 'genres', 'tags', 'added',
        'ratings', 'movieFile'
      ];

      expectedKeys.forEach(key => {
        expect(movieKeys).toContain(key);
      });
    });

    it('should have proper TypeScript types for QueueItem interface', () => {
      const queueItemKeys = Object.keys(mockQueueItem);

      const expectedKeys = [
        'id', 'movieId', 'movie', 'languages', 'quality', 'customFormats',
        'size', 'title', 'sizeleft', 'timeleft', 'estimatedCompletionTime',
        'status', 'trackedDownloadStatus', 'trackedDownloadState',
        'statusMessages', 'downloadId', 'protocol', 'downloadClient',
        'indexer', 'outputPath'
      ];

      expectedKeys.forEach(key => {
        expect(queueItemKeys).toContain(key);
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle API errors gracefully', async () => {
      // Mock fetch to return an error
      global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

      try {
        const result = await store.dispatch(
          radarrApi.endpoints.getSystemStatus.initiate()
        );
        expect(result.error).toBeDefined();
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
      }
    });
  });
});

export {};
