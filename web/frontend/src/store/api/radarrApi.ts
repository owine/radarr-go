import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
  Movie,
  SystemStatus,
  HealthCheck,
  QualityProfile,
  RootFolder,
  MovieSearchParams,
  AddMovieRequest,
  Indexer,
  DownloadClient,
  DownloadClientStats,
  ImportList,
  ImportListStats,
  QueueItem,
  QueueStats,
  QueueSearchParams,
  History,
  HistoryStats,
  HistorySearchParams,
  Activity,
  CustomFormat,
  QualityDefinitionResource,
  Notification,
  NotificationProvider,
  HostConfig,
  NamingConfig,
  MediaManagementConfig,
  Tag,
  Release,
  ReleaseSearchParams,
  CalendarEvent,
  CalendarSearchParams,
  WantedMovie,
  WantedMoviesStats,
  WantedMoviesSearchParams,
  ParseResult,
  DiscoverMovie,
  ConfigStats,
  Command,
  Collection,
  CollectionStats,
  DiskSpace,
  SystemResources,
  PerformanceMetrics,
  PaginatedResponse
} from '../../types/api';
import type { RootState } from '../index';

// Base query with authentication
const baseQuery = fetchBaseQuery({
  baseUrl: '/api/v3/',
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).auth.apiKey;

    if (token) {
      headers.set('X-Api-Key', token);
    }

    headers.set('Content-Type', 'application/json');
    return headers;
  },
});

export const radarrApi = createApi({
  reducerPath: 'radarrApi',
  baseQuery,
  tagTypes: [
    'Movie',
    'MovieFile',
    'QualityProfile',
    'QualityDefinition',
    'CustomFormat',
    'RootFolder',
    'SystemStatus',
    'Health',
    'Collection',
    'Indexer',
    'DownloadClient',
    'ImportList',
    'Queue',
    'History',
    'Activity',
    'Notification',
    'Config',
    'Tag',
    'Release',
    'Calendar',
    'WantedMovie',
    'Parse',
    'FileOrganization',
    'Command',
    'SystemResource'
  ],
  endpoints: (builder) => ({
    // System endpoints
    getSystemStatus: builder.query<SystemStatus, void>({
      query: () => 'system/status',
      providesTags: ['SystemStatus'],
    }),

    getHealth: builder.query<HealthCheck[], void>({
      query: () => 'health',
      providesTags: ['Health'],
    }),

    // Movie endpoints
    getMovies: builder.query<Movie[], MovieSearchParams | void>({
      query: (params) => ({
        url: 'movie',
        params: params ? {
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'title',
          sortDirection: params.sortDirection || 'asc',
          term: params.term,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'title',
          sortDirection: 'asc',
        },
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Movie' as const, id })),
              { type: 'Movie', id: 'LIST' },
            ]
          : [{ type: 'Movie', id: 'LIST' }],
    }),

    getMovie: builder.query<Movie, number>({
      query: (id) => `movie/${id}`,
      providesTags: (_, __, id) => [{ type: 'Movie', id }],
    }),

    addMovie: builder.mutation<Movie, AddMovieRequest>({
      query: (movie) => ({
        url: 'movie',
        method: 'POST',
        body: movie,
      }),
      invalidatesTags: [{ type: 'Movie', id: 'LIST' }],
    }),

    updateMovie: builder.mutation<Movie, Partial<Movie> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `movie/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Movie', id },
        { type: 'Movie', id: 'LIST' },
      ],
    }),

    deleteMovie: builder.mutation<void, { id: number; deleteFiles?: boolean; addImportExclusion?: boolean }>({
      query: ({ id, deleteFiles = false, addImportExclusion = false }) => ({
        url: `movie/${id}`,
        method: 'DELETE',
        params: { deleteFiles, addImportExclusion },
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Movie', id },
        { type: 'Movie', id: 'LIST' },
      ],
    }),

    // Search for movies to add
    searchMovies: builder.query<Movie[], string>({
      query: (term) => ({
        url: 'movie/lookup',
        params: { term },
      }),
    }),

    // Quality Profile endpoints
    getQualityProfiles: builder.query<QualityProfile[], void>({
      query: () => 'qualityprofile',
      providesTags: [{ type: 'QualityProfile', id: 'LIST' }],
    }),

    getQualityProfile: builder.query<QualityProfile, number>({
      query: (id) => `qualityprofile/${id}`,
      providesTags: (_, __, id) => [{ type: 'QualityProfile', id }],
    }),

    // Root Folder endpoints
    getRootFolders: builder.query<RootFolder[], void>({
      query: () => 'rootfolder',
      providesTags: [{ type: 'RootFolder', id: 'LIST' }],
    }),

    // Movie refresh and search
    refreshMovie: builder.mutation<void, number>({
      query: (id) => ({
        url: `command`,
        method: 'POST',
        body: {
          name: 'RefreshMovie',
          movieId: id,
        },
      }),
      invalidatesTags: (_, __, id) => [{ type: 'Movie', id }],
    }),

    searchMovie: builder.mutation<void, number>({
      query: (id) => ({
        url: `command`,
        method: 'POST',
        body: {
          name: 'MoviesSearch',
          movieIds: [id],
        },
      }),
    }),

    toggleMovieMonitor: builder.mutation<Movie, { movieId: number; monitored: boolean }>({
      query: ({ movieId, monitored }) => ({
        url: `movie/${movieId}`,
        method: 'PUT',
        body: { monitored },
      }),
      invalidatesTags: (_, __, { movieId }) => [
        { type: 'Movie', id: movieId },
        { type: 'Movie', id: 'LIST' },
      ],
    }),

    // Bulk operations
    refreshAllMovies: builder.mutation<void, void>({
      query: () => ({
        url: `command`,
        method: 'POST',
        body: {
          name: 'RefreshMovie',
        },
      }),
      invalidatesTags: [{ type: 'Movie', id: 'LIST' }],
    }),


    // Queue endpoints
    getQueue: builder.query<PaginatedResponse<QueueItem>, QueueSearchParams | void>({
      query: (params) => ({
        url: 'queue',
        params: params ? {
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'timeleft',
          sortDirection: params.sortDirection || 'asc',
          includeUnknownMovieItems: params.includeUnknownMovieItems || false,
          includeMovie: params.includeMovie || true,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'timeleft',
          sortDirection: 'asc',
          includeUnknownMovieItems: false,
          includeMovie: true,
        },
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.records.map(({ id }) => ({ type: 'Queue' as const, id })),
              { type: 'Queue', id: 'LIST' },
            ]
          : [{ type: 'Queue', id: 'LIST' }],
    }),

    getQueueItem: builder.query<QueueItem, number>({
      query: (id) => `queue/${id}`,
      providesTags: (_, __, id) => [{ type: 'Queue', id }],
    }),

    removeQueueItem: builder.mutation<void, { id: number; removeFromClient?: boolean; blocklist?: boolean }>({
      query: ({ id, removeFromClient = true, blocklist = false }) => ({
        url: `queue/${id}`,
        method: 'DELETE',
        params: { removeFromClient, blocklist },
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Queue', id },
        { type: 'Queue', id: 'LIST' },
      ],
    }),

    removeQueueItems: builder.mutation<void, { ids: number[]; removeFromClient?: boolean; blocklist?: boolean }>({
      query: ({ ids, removeFromClient = true, blocklist = false }) => ({
        url: 'queue/bulk',
        method: 'DELETE',
        body: { ids, removeFromClient, blocklist },
      }),
      invalidatesTags: [{ type: 'Queue', id: 'LIST' }],
    }),

    getQueueStats: builder.query<QueueStats, void>({
      query: () => 'queue/stats',
      providesTags: [{ type: 'Queue', id: 'STATS' }],
    }),

    // History endpoints
    getHistory: builder.query<PaginatedResponse<History>, HistorySearchParams | void>({
      query: (params) => ({
        url: 'history',
        params: params ? {
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'date',
          sortDirection: params.sortDirection || 'desc',
          movieId: params.movieId,
          eventType: params.eventType,
          since: params.since,
          until: params.until,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'date',
          sortDirection: 'desc',
        },
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.records.map(({ id }) => ({ type: 'History' as const, id })),
              { type: 'History', id: 'LIST' },
            ]
          : [{ type: 'History', id: 'LIST' }],
    }),

    getHistoryStats: builder.query<HistoryStats, void>({
      query: () => 'history/stats',
      providesTags: [{ type: 'History', id: 'STATS' }],
    }),

    // Activity endpoints
    getActivity: builder.query<Activity[], void>({
      query: () => 'activity',
      providesTags: [{ type: 'Activity', id: 'LIST' }],
    }),

    getRunningActivities: builder.query<Activity[], void>({
      query: () => 'activity/running',
      providesTags: [{ type: 'Activity', id: 'RUNNING' }],
    }),

    // Indexer endpoints
    getIndexers: builder.query<Indexer[], void>({
      query: () => 'indexer',
      providesTags: [{ type: 'Indexer', id: 'LIST' }],
    }),

    getIndexer: builder.query<Indexer, number>({
      query: (id) => `indexer/${id}`,
      providesTags: (_, __, id) => [{ type: 'Indexer', id }],
    }),

    createIndexer: builder.mutation<Indexer, Partial<Indexer>>({
      query: (indexer) => ({
        url: 'indexer',
        method: 'POST',
        body: indexer,
      }),
      invalidatesTags: [{ type: 'Indexer', id: 'LIST' }],
    }),

    updateIndexer: builder.mutation<Indexer, Partial<Indexer> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `indexer/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Indexer', id },
        { type: 'Indexer', id: 'LIST' },
      ],
    }),

    deleteIndexer: builder.mutation<void, number>({
      query: (id) => ({
        url: `indexer/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'Indexer', id },
        { type: 'Indexer', id: 'LIST' },
      ],
    }),

    testIndexer: builder.mutation<void, number>({
      query: (id) => ({
        url: `indexer/${id}/test`,
        method: 'POST',
      }),
    }),

    // Download Client endpoints
    getDownloadClients: builder.query<DownloadClient[], void>({
      query: () => 'downloadclient',
      providesTags: [{ type: 'DownloadClient', id: 'LIST' }],
    }),

    getDownloadClient: builder.query<DownloadClient, number>({
      query: (id) => `downloadclient/${id}`,
      providesTags: (_, __, id) => [{ type: 'DownloadClient', id }],
    }),

    createDownloadClient: builder.mutation<DownloadClient, Partial<DownloadClient>>({
      query: (client) => ({
        url: 'downloadclient',
        method: 'POST',
        body: client,
      }),
      invalidatesTags: [{ type: 'DownloadClient', id: 'LIST' }],
    }),

    updateDownloadClient: builder.mutation<DownloadClient, Partial<DownloadClient> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `downloadclient/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'DownloadClient', id },
        { type: 'DownloadClient', id: 'LIST' },
      ],
    }),

    deleteDownloadClient: builder.mutation<void, number>({
      query: (id) => ({
        url: `downloadclient/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'DownloadClient', id },
        { type: 'DownloadClient', id: 'LIST' },
      ],
    }),

    testDownloadClient: builder.mutation<void, Partial<DownloadClient>>({
      query: (client) => ({
        url: 'downloadclient/test',
        method: 'POST',
        body: client,
      }),
    }),

    getDownloadClientStats: builder.query<DownloadClientStats, void>({
      query: () => 'downloadclient/stats',
      providesTags: [{ type: 'DownloadClient', id: 'STATS' }],
    }),

    // Import List endpoints
    getImportLists: builder.query<ImportList[], void>({
      query: () => 'importlist',
      providesTags: [{ type: 'ImportList', id: 'LIST' }],
    }),

    getImportList: builder.query<ImportList, number>({
      query: (id) => `importlist/${id}`,
      providesTags: (_, __, id) => [{ type: 'ImportList', id }],
    }),

    createImportList: builder.mutation<ImportList, Partial<ImportList>>({
      query: (list) => ({
        url: 'importlist',
        method: 'POST',
        body: list,
      }),
      invalidatesTags: [{ type: 'ImportList', id: 'LIST' }],
    }),

    updateImportList: builder.mutation<ImportList, Partial<ImportList> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `importlist/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'ImportList', id },
        { type: 'ImportList', id: 'LIST' },
      ],
    }),

    deleteImportList: builder.mutation<void, number>({
      query: (id) => ({
        url: `importlist/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'ImportList', id },
        { type: 'ImportList', id: 'LIST' },
      ],
    }),

    testImportList: builder.mutation<void, Partial<ImportList>>({
      query: (list) => ({
        url: 'importlist/test',
        method: 'POST',
        body: list,
      }),
    }),

    syncImportList: builder.mutation<void, number>({
      query: (id) => ({
        url: `importlist/${id}/sync`,
        method: 'POST',
      }),
      invalidatesTags: [{ type: 'ImportList', id: 'LIST' }, { type: 'Movie', id: 'LIST' }],
    }),

    syncAllImportLists: builder.mutation<void, void>({
      query: () => ({
        url: 'importlist/sync',
        method: 'POST',
      }),
      invalidatesTags: [{ type: 'ImportList', id: 'LIST' }, { type: 'Movie', id: 'LIST' }],
    }),

    getImportListStats: builder.query<ImportListStats, void>({
      query: () => 'importlist/stats',
      providesTags: [{ type: 'ImportList', id: 'STATS' }],
    }),

    // Quality management endpoints
    getQualityDefinitions: builder.query<QualityDefinitionResource[], void>({
      query: () => 'qualitydefinition',
      providesTags: [{ type: 'QualityDefinition', id: 'LIST' }],
    }),

    getQualityDefinition: builder.query<QualityDefinitionResource, number>({
      query: (id) => `qualitydefinition/${id}`,
      providesTags: (_, __, id) => [{ type: 'QualityDefinition', id }],
    }),

    updateQualityDefinition: builder.mutation<QualityDefinitionResource, Partial<QualityDefinitionResource> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `qualitydefinition/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'QualityDefinition', id },
        { type: 'QualityDefinition', id: 'LIST' },
      ],
    }),

    // Custom Format endpoints
    getCustomFormats: builder.query<CustomFormat[], void>({
      query: () => 'customformat',
      providesTags: [{ type: 'CustomFormat', id: 'LIST' }],
    }),

    getCustomFormat: builder.query<CustomFormat, number>({
      query: (id) => `customformat/${id}`,
      providesTags: (_, __, id) => [{ type: 'CustomFormat', id }],
    }),

    createCustomFormat: builder.mutation<CustomFormat, Partial<CustomFormat>>({
      query: (format) => ({
        url: 'customformat',
        method: 'POST',
        body: format,
      }),
      invalidatesTags: [{ type: 'CustomFormat', id: 'LIST' }],
    }),

    updateCustomFormat: builder.mutation<CustomFormat, Partial<CustomFormat> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `customformat/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'CustomFormat', id },
        { type: 'CustomFormat', id: 'LIST' },
      ],
    }),

    deleteCustomFormat: builder.mutation<void, number>({
      query: (id) => ({
        url: `customformat/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'CustomFormat', id },
        { type: 'CustomFormat', id: 'LIST' },
      ],
    }),

    // Notification endpoints
    getNotifications: builder.query<Notification[], void>({
      query: () => 'notification',
      providesTags: [{ type: 'Notification', id: 'LIST' }],
    }),

    getNotification: builder.query<Notification, number>({
      query: (id) => `notification/${id}`,
      providesTags: (_, __, id) => [{ type: 'Notification', id }],
    }),

    createNotification: builder.mutation<Notification, Partial<Notification>>({
      query: (notification) => ({
        url: 'notification',
        method: 'POST',
        body: notification,
      }),
      invalidatesTags: [{ type: 'Notification', id: 'LIST' }],
    }),

    updateNotification: builder.mutation<Notification, Partial<Notification> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `notification/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Notification', id },
        { type: 'Notification', id: 'LIST' },
      ],
    }),

    deleteNotification: builder.mutation<void, number>({
      query: (id) => ({
        url: `notification/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'Notification', id },
        { type: 'Notification', id: 'LIST' },
      ],
    }),

    testNotification: builder.mutation<void, Partial<Notification>>({
      query: (notification) => ({
        url: 'notification/test',
        method: 'POST',
        body: notification,
      }),
    }),

    getNotificationProviders: builder.query<NotificationProvider[], void>({
      query: () => 'notification/schema',
    }),

    // Configuration endpoints
    getHostConfig: builder.query<HostConfig, void>({
      query: () => 'config/host',
      providesTags: [{ type: 'Config', id: 'HOST' }],
    }),

    updateHostConfig: builder.mutation<HostConfig, Partial<HostConfig>>({
      query: (config) => ({
        url: 'config/host',
        method: 'PUT',
        body: config,
      }),
      invalidatesTags: [{ type: 'Config', id: 'HOST' }],
    }),

    getNamingConfig: builder.query<NamingConfig, void>({
      query: () => 'config/naming',
      providesTags: [{ type: 'Config', id: 'NAMING' }],
    }),

    updateNamingConfig: builder.mutation<NamingConfig, Partial<NamingConfig>>({
      query: (config) => ({
        url: 'config/naming',
        method: 'PUT',
        body: config,
      }),
      invalidatesTags: [{ type: 'Config', id: 'NAMING' }],
    }),

    getMediaManagementConfig: builder.query<MediaManagementConfig, void>({
      query: () => 'config/mediamanagement',
      providesTags: [{ type: 'Config', id: 'MEDIA' }],
    }),

    updateMediaManagementConfig: builder.mutation<MediaManagementConfig, Partial<MediaManagementConfig>>({
      query: (config) => ({
        url: 'config/mediamanagement',
        method: 'PUT',
        body: config,
      }),
      invalidatesTags: [{ type: 'Config', id: 'MEDIA' }],
    }),

    getConfigStats: builder.query<ConfigStats, void>({
      query: () => 'config/stats',
      providesTags: [{ type: 'Config', id: 'STATS' }],
    }),

    // Additional configuration endpoints
    getNamingTokens: builder.query<string[], void>({
      query: () => 'config/naming/tokens',
    }),

    previewNaming: builder.query<{ filename: string; folder?: string }, number>({
      query: (movieId) => `config/naming/preview/${movieId}`,
    }),

    // Tag endpoints
    getTags: builder.query<Tag[], void>({
      query: () => 'tag',
      providesTags: [{ type: 'Tag', id: 'LIST' }],
    }),

    createTag: builder.mutation<Tag, { label: string }>({
      query: (tag) => ({
        url: 'tag',
        method: 'POST',
        body: tag,
      }),
      invalidatesTags: [{ type: 'Tag', id: 'LIST' }],
    }),

    updateTag: builder.mutation<Tag, Partial<Tag> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `tag/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Tag', id },
        { type: 'Tag', id: 'LIST' },
      ],
    }),

    deleteTag: builder.mutation<void, number>({
      query: (id) => ({
        url: `tag/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'Tag', id },
        { type: 'Tag', id: 'LIST' },
      ],
    }),

    // Release and Search endpoints
    getReleases: builder.query<Release[], ReleaseSearchParams | void>({
      query: (params) => ({
        url: 'release',
        params: params ? {
          movieId: params.movieId,
          term: params.term,
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'releaseWeight',
          sortDirection: params.sortDirection || 'desc',
          includeUnknownMovieItems: params.includeUnknownMovieItems || false,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'releaseWeight',
          sortDirection: 'desc',
          includeUnknownMovieItems: false,
        },
      }),
      providesTags: [{ type: 'Release', id: 'LIST' }],
    }),

    searchMovieReleases: builder.query<Release[], number>({
      query: (movieId) => `search/movie/${movieId}`,
      providesTags: [{ type: 'Release', id: 'SEARCH' }],
    }),

    grabRelease: builder.mutation<void, { guid: string; indexerId: number }>({
      query: (release) => ({
        url: 'release/grab',
        method: 'POST',
        body: release,
      }),
      invalidatesTags: [{ type: 'Queue', id: 'LIST' }, { type: 'Activity', id: 'LIST' }],
    }),

    // Calendar endpoints
    getCalendar: builder.query<CalendarEvent[], CalendarSearchParams>({
      query: (params) => ({
        url: 'calendar',
        params: {
          start: params.start,
          end: params.end,
          unmonitored: params.unmonitored || false,
          tags: params.tags,
        },
      }),
      providesTags: [{ type: 'Calendar', id: 'LIST' }],
    }),

    // Wanted Movies endpoints
    getMissingMovies: builder.query<PaginatedResponse<WantedMovie>, WantedMoviesSearchParams | void>({
      query: (params) => ({
        url: 'wanted/missing',
        params: params ? {
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'title',
          sortDirection: params.sortDirection || 'asc',
          status: params.status,
          priority: params.priority,
          monitored: params.monitored,
          hasFile: params.hasFile,
          movieId: params.movieId,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'title',
          sortDirection: 'asc',
        },
      }),
      providesTags: [{ type: 'WantedMovie', id: 'MISSING' }],
    }),

    getCutoffUnmetMovies: builder.query<PaginatedResponse<WantedMovie>, WantedMoviesSearchParams | void>({
      query: (params) => ({
        url: 'wanted/cutoff',
        params: params ? {
          page: params.page || 1,
          pageSize: params.pageSize || 20,
          sortKey: params.sortKey || 'title',
          sortDirection: params.sortDirection || 'asc',
          status: params.status,
          priority: params.priority,
          monitored: params.monitored,
          hasFile: params.hasFile,
          movieId: params.movieId,
        } : {
          page: 1,
          pageSize: 20,
          sortKey: 'title',
          sortDirection: 'asc',
        },
      }),
      providesTags: [{ type: 'WantedMovie', id: 'CUTOFF' }],
    }),

    getWantedStats: builder.query<WantedMoviesStats, void>({
      query: () => 'wanted/stats',
      providesTags: [{ type: 'WantedMovie', id: 'STATS' }],
    }),

    // Movie Discovery endpoints
    getPopularMovies: builder.query<DiscoverMovie[], void>({
      query: () => 'movie/popular',
    }),

    getTrendingMovies: builder.query<DiscoverMovie[], void>({
      query: () => 'movie/trending',
    }),

    // Parse endpoints
    parseReleaseTitle: builder.query<ParseResult, string>({
      query: (title) => ({
        url: 'parse',
        params: { title },
      }),
      providesTags: [{ type: 'Parse', id: 'SINGLE' }],
    }),

    // Command/Task endpoints
    getCommands: builder.query<Command[], void>({
      query: () => 'command',
      providesTags: [{ type: 'Command', id: 'LIST' }],
    }),

    getCommand: builder.query<Command, number>({
      query: (id) => `command/${id}`,
      providesTags: (_, __, id) => [{ type: 'Command', id }],
    }),

    queueCommand: builder.mutation<Command, { name: string; [key: string]: unknown }>({
      query: (command) => ({
        url: 'command',
        method: 'POST',
        body: command,
      }),
      invalidatesTags: [{ type: 'Command', id: 'LIST' }, { type: 'Activity', id: 'LIST' }],
    }),

    cancelCommand: builder.mutation<void, number>({
      query: (id) => ({
        url: `command/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'Command', id },
        { type: 'Command', id: 'LIST' },
      ],
    }),

    // System Resource endpoints
    getSystemResources: builder.query<SystemResources, void>({
      query: () => 'health/system/resources',
      providesTags: [{ type: 'SystemResource', id: 'CURRENT' }],
    }),

    getDiskSpace: builder.query<DiskSpace[], void>({
      query: () => 'health/system/diskspace',
      providesTags: [{ type: 'SystemResource', id: 'DISK' }],
    }),

    getPerformanceMetrics: builder.query<PerformanceMetrics[], { since?: string; until?: string }>({
      query: (params) => ({
        url: 'health/metrics',
        params,
      }),
      providesTags: [{ type: 'SystemResource', id: 'METRICS' }],
    }),

    // Collection management
    getCollections: builder.query<Collection[], void>({
      query: () => 'collection',
      providesTags: [{ type: 'Collection', id: 'LIST' }],
    }),

    getCollection: builder.query<Collection, number>({
      query: (id) => `collection/${id}`,
      providesTags: (_, __, id) => [{ type: 'Collection', id }],
    }),

    createCollection: builder.mutation<Collection, Partial<Collection>>({
      query: (collection) => ({
        url: 'collection',
        method: 'POST',
        body: collection,
      }),
      invalidatesTags: [{ type: 'Collection', id: 'LIST' }],
    }),

    updateCollection: builder.mutation<Collection, Partial<Collection> & { id: number }>({
      query: ({ id, ...patch }) => ({
        url: `collection/${id}`,
        method: 'PUT',
        body: patch,
      }),
      invalidatesTags: (_, __, { id }) => [
        { type: 'Collection', id },
        { type: 'Collection', id: 'LIST' },
      ],
    }),

    deleteCollection: builder.mutation<void, number>({
      query: (id) => ({
        url: `collection/${id}`,
        method: 'DELETE',
      }),
      invalidatesTags: (_, __, id) => [
        { type: 'Collection', id },
        { type: 'Collection', id: 'LIST' },
      ],
    }),

    getCollectionStats: builder.query<CollectionStats, number>({
      query: (id) => `collection/${id}/statistics`,
      providesTags: (_, __, id) => [{ type: 'Collection', id: `${id}_STATS` }],
    }),
  }),
});

export const {
  // System hooks
  useGetSystemStatusQuery,
  useGetHealthQuery,

  // Movie hooks
  useGetMoviesQuery,
  useGetMovieQuery,
  useAddMovieMutation,
  useUpdateMovieMutation,
  useDeleteMovieMutation,
  useSearchMoviesQuery,
  useGetPopularMoviesQuery,
  useGetTrendingMoviesQuery,

  // Quality Profile hooks
  useGetQualityProfilesQuery,
  useGetQualityProfileQuery,

  // Quality Definition hooks
  useGetQualityDefinitionsQuery,
  useGetQualityDefinitionQuery,
  useUpdateQualityDefinitionMutation,

  // Custom Format hooks
  useGetCustomFormatsQuery,
  useGetCustomFormatQuery,
  useCreateCustomFormatMutation,
  useUpdateCustomFormatMutation,
  useDeleteCustomFormatMutation,

  // Root Folder hooks
  useGetRootFoldersQuery,

  // Movie operations hooks
  useRefreshMovieMutation,
  useSearchMovieMutation,
  useToggleMovieMonitorMutation,
  useRefreshAllMoviesMutation,

  // Calendar hooks
  useGetCalendarQuery,

  // History hooks
  useGetHistoryQuery,
  useGetHistoryStatsQuery,

  // Activity hooks
  useGetActivityQuery,
  useGetRunningActivitiesQuery,

  // Queue hooks
  useGetQueueQuery,
  useGetQueueItemQuery,
  useRemoveQueueItemMutation,
  useRemoveQueueItemsMutation,
  useGetQueueStatsQuery,

  // Indexer hooks
  useGetIndexersQuery,
  useGetIndexerQuery,
  useCreateIndexerMutation,
  useUpdateIndexerMutation,
  useDeleteIndexerMutation,
  useTestIndexerMutation,

  // Download Client hooks
  useGetDownloadClientsQuery,
  useGetDownloadClientQuery,
  useCreateDownloadClientMutation,
  useUpdateDownloadClientMutation,
  useDeleteDownloadClientMutation,
  useTestDownloadClientMutation,
  useGetDownloadClientStatsQuery,

  // Import List hooks
  useGetImportListsQuery,
  useGetImportListQuery,
  useCreateImportListMutation,
  useUpdateImportListMutation,
  useDeleteImportListMutation,
  useTestImportListMutation,
  useSyncImportListMutation,
  useSyncAllImportListsMutation,
  useGetImportListStatsQuery,

  // Notification hooks
  useGetNotificationsQuery,
  useGetNotificationQuery,
  useCreateNotificationMutation,
  useUpdateNotificationMutation,
  useDeleteNotificationMutation,
  useTestNotificationMutation,
  useGetNotificationProvidersQuery,

  // Configuration hooks
  useGetHostConfigQuery,
  useUpdateHostConfigMutation,
  useGetNamingConfigQuery,
  useUpdateNamingConfigMutation,
  useGetMediaManagementConfigQuery,
  useUpdateMediaManagementConfigMutation,
  useGetConfigStatsQuery,
  useGetNamingTokensQuery,
  usePreviewNamingQuery,

  // Tag hooks
  useGetTagsQuery,
  useCreateTagMutation,
  useUpdateTagMutation,
  useDeleteTagMutation,

  // Release and Search hooks
  useGetReleasesQuery,
  useSearchMovieReleasesQuery,
  useGrabReleaseMutation,

  // Wanted Movies hooks
  useGetMissingMoviesQuery,
  useGetCutoffUnmetMoviesQuery,
  useGetWantedStatsQuery,

  // Parse hooks
  useParseReleaseTitleQuery,

  // Command/Task hooks
  useGetCommandsQuery,
  useGetCommandQuery,
  useQueueCommandMutation,
  useCancelCommandMutation,

  // System Resource hooks
  useGetSystemResourcesQuery,
  useGetDiskSpaceQuery,
  useGetPerformanceMetricsQuery,

  // Collection hooks
  useGetCollectionsQuery,
  useGetCollectionQuery,
  useCreateCollectionMutation,
  useUpdateCollectionMutation,
  useDeleteCollectionMutation,
  useGetCollectionStatsQuery,
} = radarrApi;
