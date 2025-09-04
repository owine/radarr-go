// API Response Types for Radarr Go

export interface Movie {
  id: number;
  title: string;
  originalTitle?: string;
  sortTitle?: string;
  status: string;
  overview?: string;
  inCinemas?: string;
  physicalRelease?: string;
  digitalRelease?: string;
  images?: MovieImage[];
  website?: string;
  year: number;
  hasFile: boolean;
  youTubeTrailerId?: string;
  studio?: string;
  path: string;
  pathState: string;
  qualityProfileId: number;
  monitored: boolean;
  minimumAvailability: string;
  isAvailable: boolean;
  folderName?: string;
  runtime: number;
  lastInfoSync?: string;
  cleanTitle: string;
  imdbId?: string;
  tmdbId: number;
  titleSlug: string;
  certification?: string;
  genres: string[];
  tags: number[];
  added: string;
  ratings: MovieRatings;
  movieFile?: MovieFile;
  collection?: Collection;
}

export interface MovieImage {
  coverType: string;
  url: string;
  remoteUrl?: string;
}

export interface MovieRatings {
  votes: number;
  value: number;
}

export interface MovieFile {
  id: number;
  movieId: number;
  relativePath: string;
  path: string;
  size: number;
  dateAdded: string;
  sceneName?: string;
  releaseGroup?: string;
  quality: Quality;
  mediaInfo: MediaInfo;
  originalFilePath?: string;
  qualityCutoffNotMet: boolean;
  languages?: Language[];
}

export interface Quality {
  quality: QualityDefinition;
  revision: QualityRevision;
}

export interface QualityDefinition {
  id: number;
  name: string;
  source: string;
  resolution: number;
  modifier: string;
}

export interface QualityRevision {
  version: number;
  real: number;
  isRepack: boolean;
}

export interface MediaInfo {
  audioChannels: number;
  audioCodec: string;
  audioLanguages: string;
  height: number;
  width: number;
  runtime: number;
  videoCodec: string;
  videoDynamicRange: string;
  videoDynamicRangeType: string;
}

export interface Language {
  id: number;
  name: string;
}

export interface Collection {
  id: number;
  title: string;
  sortTitle: string;
  tmdbId: number;
  images?: MovieImage[];
  overview?: string;
  monitored: boolean;
  rootFolderPath: string;
  qualityProfileId: number;
  searchOnAdd: boolean;
  minimumAvailability: string;
  movies?: Movie[];
  tags: number[];
}

export interface QualityProfile {
  id: number;
  name: string;
  upgradeAllowed: boolean;
  cutoff: number;
  items: QualityProfileItem[];
  minFormatScore: number;
  cutoffFormatScore: number;
  formatItems: FormatItem[];
  language: Language;
}

export interface QualityProfileItem {
  id?: number;
  name?: string;
  quality?: QualityDefinition;
  items?: QualityProfileItem[];
  allowed: boolean;
}

export interface FormatItem {
  format: number;
  name: string;
  score: number;
}

export interface SystemStatus {
  appName: string;
  instanceName: string;
  version: string;
  buildTime: string;
  isDebug: boolean;
  isProduction: boolean;
  isAdmin: boolean;
  isUserInteractive: boolean;
  startupPath: string;
  appData: string;
  osName: string;
  osVersion: string;
  isMonoRuntime: boolean;
  isMono: boolean;
  isLinux: boolean;
  isOsx: boolean;
  isWindows: boolean;
  isDocker: boolean;
  mode: string;
  branch: string;
  authentication: string;
  sqliteVersion: string;
  migrationVersion: number;
  urlBase: string;
  runtimeVersion: string;
  runtimeName: string;
  startTime: string;
  packageVersion: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
}

export interface HealthCheck {
  source: string;
  type: string;
  message: string;
  wikiUrl?: string;
}

export interface RootFolder {
  id: number;
  path: string;
  accessible: boolean;
  freeSpace: number;
  totalSpace: number;
  unmappedFolders?: UnmappedFolder[];
}

export interface UnmappedFolder {
  name: string;
  path: string;
  relativePath: string;
}

// API Error Types
export interface ApiError {
  message: string;
  description?: string;
  details?: string;
}

// RTK Query base query types
export interface ApiResponse<T> {
  data: T;
  status: number;
}

export interface PaginatedResponse<T> {
  page: number;
  pageSize: number;
  sortKey: string;
  sortDirection: string;
  totalRecords: number;
  records: T[];
}

// Request types
export interface MovieSearchParams {
  term?: string;
  page?: number;
  pageSize?: number;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
}

export interface AddMovieRequest {
  title: string;
  year: number;
  tmdbId: number;
  qualityProfileId: number;
  rootFolderPath: string;
  monitored: boolean;
  minimumAvailability: string;
  tags?: number[];
  searchForMovie?: boolean;
}

// Indexer Types
export interface Indexer {
  id: number;
  name: string;
  implementation: string;
  configContract: string;
  infoLink: string;
  fields: IndexerField[];
  enableRss: boolean;
  enableAutomaticSearch: boolean;
  enableInteractiveSearch: boolean;
  supportsRss: boolean;
  supportsSearch: boolean;
  protocol: string;
  priority: number;
  downloadClientId?: number;
  tags: number[];
}

export interface IndexerField {
  name: string;
  label: string;
  helpText?: string;
  value?: string | number | boolean | string[] | number[];
  type: string;
  advanced?: boolean;
  selectOptions?: SelectOption[];
}

export interface SelectOption {
  value: string | number | boolean;
  name: string;
  hint?: string;
}

// Download Client Types
export interface DownloadClient {
  id: number;
  name: string;
  implementation: string;
  configContract: string;
  infoLink: string;
  fields: DownloadClientField[];
  enable: boolean;
  protocol: string;
  priority: number;
  removeCompletedDownloads: boolean;
  removeFailedDownloads: boolean;
  tags: number[];
}

export interface DownloadClientField {
  name: string;
  label: string;
  helpText?: string;
  value?: string | number | boolean | string[] | number[];
  type: string;
  advanced?: boolean;
  selectOptions?: SelectOption[];
}

export interface DownloadClientStats {
  totalItems: number;
  downloading: number;
  completed: number;
  failed: number;
  queued: number;
}

// Import List Types
export interface ImportList {
  id: number;
  name: string;
  implementation: string;
  configContract: string;
  infoLink: string;
  fields: ImportListField[];
  enableAuto: boolean;
  shouldMonitor: boolean;
  rootFolderPath: string;
  qualityProfileId: number;
  minimumAvailability: string;
  tags: number[];
  listType: string;
  listOrder: number;
}

export interface ImportListField {
  name: string;
  label: string;
  helpText?: string;
  value?: string | number | boolean | string[] | number[];
  type: string;
  advanced?: boolean;
  selectOptions?: SelectOption[];
}

export interface ImportListStats {
  totalLists: number;
  enabledLists: number;
  lastSyncTime?: string;
  totalMoviesAdded: number;
}

// Queue Types
export interface QueueItem {
  id: number;
  movieId: number;
  movie: Movie;
  languages?: Language[];
  quality: Quality;
  customFormats?: CustomFormat[];
  size: number;
  title: string;
  sizeleft: number;
  timeleft?: string;
  estimatedCompletionTime?: string;
  status: string;
  trackedDownloadStatus: string;
  trackedDownloadState: string;
  statusMessages?: QueueStatusMessage[];
  errorMessage?: string;
  downloadId: string;
  protocol: string;
  downloadClient: string;
  indexer: string;
  outputPath?: string;
}

export interface QueueStatusMessage {
  title: string;
  messages: string[];
}

export interface QueueStats {
  totalItems: number;
  downloading: number;
  queued: number;
  completed: number;
  failed: number;
  warnings: number;
  errors: number;
}

// History Types
export interface History {
  id: number;
  movieId: number;
  movie: Movie;
  languages?: Language[];
  quality: Quality;
  customFormats?: CustomFormat[];
  date: string;
  downloadId?: string;
  eventType: string;
  data: Record<string, unknown>;
  sourceTitle?: string;
}

export interface HistoryStats {
  totalRecords: number;
  grabbed: number;
  imported: number;
  failed: number;
  deleted: number;
  renamed: number;
}

// Activity Types
export interface Activity {
  id: number;
  type: string;
  status: string;
  progress: number;
  startTime: string;
  endTime?: string;
  duration?: number;
  message?: string;
  data: Record<string, unknown>;
}

// Custom Format Types
export interface CustomFormat {
  id: number;
  name: string;
  includeCustomFormatWhenRenaming?: boolean;
  specifications: CustomFormatSpecification[];
}

export interface CustomFormatSpecification {
  name: string;
  implementation: string;
  negate: boolean;
  required: boolean;
  fields: Record<string, unknown>;
}

// Quality Definition Types
export interface QualityDefinitionResource {
  id: number;
  quality: QualityDefinition;
  title: string;
  weight: number;
  minSize?: number;
  maxSize?: number;
  preferredSize?: number;
}

// Notification Types
export interface Notification {
  id: number;
  name: string;
  implementation: string;
  configContract: string;
  infoLink: string;
  fields: NotificationField[];
  implementationName: string;
  onGrab: boolean;
  onDownload: boolean;
  onUpgrade: boolean;
  onRename: boolean;
  onMovieAdded: boolean;
  onMovieDelete: boolean;
  onMovieFileDelete: boolean;
  onMovieFileDeleteForUpgrade: boolean;
  onHealthIssue: boolean;
  onApplicationUpdate: boolean;
  includeHealthWarnings: boolean;
  tags: number[];
}

export interface NotificationField {
  name: string;
  label: string;
  helpText?: string;
  value?: string | number | boolean | string[] | number[];
  type: string;
  advanced?: boolean;
  selectOptions?: SelectOption[];
}

export interface NotificationProvider {
  implementation: string;
  implementationName: string;
  infoLink: string;
  fields: NotificationField[];
  presets?: NotificationPreset[];
}

export interface NotificationPreset {
  name: string;
  fields: Record<string, unknown>;
}

// Configuration Types
export interface HostConfig {
  bindAddress: string;
  port: number;
  sslPort: number;
  enableSsl: boolean;
  launchBrowser: boolean;
  authenticationMethod: string;
  authenticationRequired: string;
  username?: string;
  password?: string;
  logLevel: string;
  consoleLogLevel: string;
  branch: string;
  apiKey: string;
  sslCertHash?: string;
  urlBase?: string;
  instanceName: string;
  updateAutomatically: boolean;
  updateMechanism: string;
  updateScriptPath?: string;
  proxyEnabled: boolean;
  proxyType?: string;
  proxyHostname?: string;
  proxyPort?: number;
  proxyUsername?: string;
  proxyPassword?: string;
  proxyBypassFilter?: string;
  proxyBypassLocalAddresses: boolean;
  certificateValidation: string;
  backupFolder?: string;
  backupInterval: number;
  backupRetention: number;
}

export interface NamingConfig {
  renameMovies: boolean;
  replaceIllegalCharacters: boolean;
  colonReplacementFormat: string;
  movieFolderFormat: string;
  standardMovieFormat: string;
  includeQuality: boolean;
  replaceSpaces: boolean;
  separator: string;
}

export interface MediaManagementConfig {
  autoUnmonitorPreviouslyDownloadedMovies: boolean;
  recycleBin?: string;
  recycleBinCleanupDays: number;
  downloadPropersAndRepacks: string;
  createEmptyMovieFolders: boolean;
  deleteEmptyFolders: boolean;
  fileDate: string;
  rescanAfterRefresh: string;
  autoRenameFolders: boolean;
  pathsDefaultStatic: boolean;
  setPermissionsLinux: boolean;
  chmodFolder?: string;
  chownGroup?: string;
  skipFreeSpaceCheckWhenImporting: boolean;
  minimumFreeSpaceWhenImporting: number;
  copyUsingHardlinks: boolean;
  useScriptImport: boolean;
  scriptImportPath?: string;
  importExtraFiles: boolean;
  extraFileExtensions?: string;
}

// Tag Types
export interface Tag {
  id: number;
  label: string;
}

// Release Types
export interface Release {
  guid: string;
  quality: Quality;
  customFormats?: CustomFormat[];
  qualityWeight: number;
  age: number;
  ageHours: number;
  ageMinutes: number;
  size: number;
  indexerId: number;
  indexer: string;
  releaseGroup?: string;
  releaseHash?: string;
  title: string;
  sceneSource: boolean;
  movieTitles: string[];
  languages?: Language[];
  approved: boolean;
  temporarilyRejected: boolean;
  rejected: boolean;
  rejections: string[];
  publishDate: string;
  commentUrl?: string;
  downloadUrl?: string;
  infoUrl?: string;
  downloadAllowed: boolean;
  releaseWeight: number;
  preferredWordScore?: number;
  magnetUrl?: string;
  infoHash?: string;
  seeders?: number;
  leechers?: number;
  protocol: string;
}

// Calendar Types
export interface CalendarEvent {
  id: number;
  movieId: number;
  movie: Movie;
  title: string;
  physicalRelease?: string;
  inCinemas?: string;
  digitalRelease?: string;
  eventType: string;
  airDate?: string;
  hasFile: boolean;
  monitored: boolean;
  unmonitored?: boolean;
}

// Wanted Movies Types
export interface WantedMovie {
  id: number;
  movieId: number;
  movie: Movie;
  status: string;
  priority: string;
  lastSearchTime?: string;
  nextSearchTime?: string;
  searchAttempts: number;
  maxSearchAttempts: number;
  lastSearchReason?: string;
  lastSearchError?: string;
  createdAt: string;
  updatedAt: string;
}

export interface WantedMoviesStats {
  totalWanted: number;
  missing: number;
  cutoffUnmet: number;
  highPriority: number;
  mediumPriority: number;
  lowPriority: number;
  recentlySearched: number;
  neverSearched: number;
}

// Parse Types
export interface ParseResult {
  title: string;
  parsedMovieInfo: {
    movieTitle: string;
    originalTitle?: string;
    movieTitleInfo: {
      title: string;
      titleWithoutYear: string;
      year: number;
    };
    quality: Quality;
    languages?: Language[];
    releaseGroup?: string;
    releaseHash?: string;
    edition?: string;
  };
  parsedEpisodeInfo?: Record<string, unknown>;
  languages?: Language[];
  releaseTitle: string;
}

// File Organization Types
export interface FileOrganization {
  id: number;
  path: string;
  newPath?: string;
  movieId?: number;
  movie?: Movie;
  quality?: Quality;
  languages?: Language[];
  size: number;
  dateAdded: string;
  status: string;
  statusMessage?: string;
}

// Search and Discovery Types
export interface DiscoverMovie {
  tmdbId: number;
  title: string;
  originalTitle?: string;
  year: number;
  overview?: string;
  images?: MovieImage[];
  runtime: number;
  certification?: string;
  genres: string[];
  ratings: MovieRatings;
  status: string;
  inCinemas?: string;
  physicalRelease?: string;
  digitalRelease?: string;
  folder?: string;
  remotePoster?: string;
  isExcluded: boolean;
  isExisting: boolean;
  isRecommendation?: boolean;
  isRecent?: boolean;
  isTrending?: boolean;
  isPopular?: boolean;
}

// System and Config Stats
export interface ConfigStats {
  qualityProfiles: number;
  customFormats: number;
  indexers: number;
  downloadClients: number;
  importLists: number;
  notifications: number;
  rootFolders: number;
  tags: number;
}

// File Operation Types
export interface FileOperation {
  id: number;
  operationType: string;
  sourceFile: string;
  destinationFile: string;
  size: number;
  bytesProcessed: number;
  progress: number;
  status: string;
  statusMessage?: string;
  startTime: string;
  endTime?: string;
  duration?: number;
  errorMessage?: string;
}

// Media Info Types
export interface MediaInfoExtraction {
  filePath: string;
  mediaInfo: MediaInfo;
  extractedAt: string;
  isValid: boolean;
  errorMessage?: string;
}

// Command/Task Types
export interface Command {
  id: number;
  name: string;
  commandName: string;
  message?: string;
  body: Record<string, unknown>;
  priority: string;
  status: string;
  progress: number;
  queued: string;
  started?: string;
  ended?: string;
  duration?: string;
  exception?: string;
  trigger: string;
  clientUserAgent?: string;
  stateChangeTime?: string;
  sendUpdatesToClient: boolean;
  updateScheduledTask?: boolean;
  lastExecutionTime?: string;
}

// Rename Types
export interface RenamePreview {
  movieId: number;
  movieFileId: number;
  existingPath: string;
  newPath: string;
}

// Collection Types (extending the basic one)
export interface CollectionStats {
  totalMovies: number;
  availableMovies: number;
  missingMovies: number;
  monitoredMovies: number;
  totalFileSize: number;
}

// Bulk Operation Types
export interface BulkOperationResult {
  successful: number;
  failed: number;
  errors: string[];
}

// Disk Space Types
export interface DiskSpace {
  path: string;
  label: string;
  freeSpace: number;
  totalSpace: number;
  usedSpace: number;
  percentUsed: number;
}

// System Resources Types
export interface SystemResources {
  cpu: {
    usage: number;
    cores: number;
  };
  memory: {
    total: number;
    used: number;
    available: number;
    percentUsed: number;
  };
  disk: DiskSpace[];
  network: {
    bytesReceived: number;
    bytesSent: number;
  };
}

// Performance Metrics Types
export interface PerformanceMetrics {
  timestamp: string;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  databaseResponseTime: number;
  apiResponseTime: number;
  activeConnections: number;
  requestsPerSecond: number;
}

// Search Parameters Types
export interface ReleaseSearchParams {
  movieId?: number;
  term?: string;
  page?: number;
  pageSize?: number;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  includeUnknownMovieItems?: boolean;
}

export interface HistorySearchParams {
  page?: number;
  pageSize?: number;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  movieId?: number;
  eventType?: string;
  since?: string;
  until?: string;
}

export interface QueueSearchParams {
  page?: number;
  pageSize?: number;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  includeUnknownMovieItems?: boolean;
  includeMovie?: boolean;
}

export interface CalendarSearchParams {
  start: string;
  end: string;
  unmonitored?: boolean;
  tags?: string;
}

export interface WantedMoviesSearchParams {
  page?: number;
  pageSize?: number;
  sortKey?: string;
  sortDirection?: 'asc' | 'desc';
  status?: string;
  priority?: string;
  monitored?: boolean;
  hasFile?: boolean;
  movieId?: number;
}
