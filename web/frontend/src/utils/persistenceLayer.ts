import { cacheManager } from './cacheManager';
import { radarrApi } from '../store/api/radarrApi';
import type { AppDispatch, RootState } from '../store';

// User preferences interface
export interface UserPreferences {
  theme: 'light' | 'dark' | 'system';
  language: string;
  timezone: string;

  // UI preferences
  sidebarCollapsed: boolean;
  compactMode: boolean;
  showPosters: boolean;
  posterSize: 'small' | 'medium' | 'large';

  // Table preferences
  movieTableColumns: string[];
  queueTableColumns: string[];
  historyTableColumns: string[];

  // Default values
  defaultQualityProfile: number;
  defaultRootFolder: string;
  defaultMinimumAvailability: string;

  // Notification preferences
  enableNotifications: boolean;
  notificationDuration: number;

  // Performance preferences
  enableAnimations: boolean;
  enableBackgroundRefresh: boolean;
  refreshInterval: number;

  // Search preferences
  searchResultsPerPage: number;
  defaultSearchSort: string;

  // Calendar preferences
  calendarStartDay: 0 | 1; // 0 = Sunday, 1 = Monday
  calendarDefaultView: 'month' | 'week' | 'agenda';

  // Advanced preferences
  enableDebugMode: boolean;
  enableBetaFeatures: boolean;
  maxConcurrentDownloads: number;
}

// Application state that should be persisted
export interface PersistedAppState {
  lastActiveTab: string;
  searchHistory: string[];
  recentlyViewedMovies: number[];
  bookmarks: Array<{ id: string; name: string; url: string; timestamp: number }>;
  dashboardLayout: Array<{ id: string; x: number; y: number; w: number; h: number }>;
  filterPresets: Record<string, unknown>;
  sortPreferences: Record<string, { key: string; direction: 'asc' | 'desc' }>;
}

// Session data that should be persisted for the current session
export interface SessionData {
  currentPage: string;
  scrollPositions: Record<string, number>;
  formDrafts: Record<string, unknown>;
  expandedSections: Record<string, boolean>;
  activeFilters: Record<string, unknown>;
  selectedItems: Record<string, number[]>;
}

// Default user preferences
const defaultPreferences: UserPreferences = {
  theme: 'system',
  language: 'en',
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,

  sidebarCollapsed: false,
  compactMode: false,
  showPosters: true,
  posterSize: 'medium',

  movieTableColumns: ['title', 'year', 'status', 'quality', 'size', 'added'],
  queueTableColumns: ['title', 'status', 'progress', 'size', 'eta', 'actions'],
  historyTableColumns: ['date', 'movie', 'event', 'quality', 'source'],

  defaultQualityProfile: 1,
  defaultRootFolder: '',
  defaultMinimumAvailability: 'announced',

  enableNotifications: true,
  notificationDuration: 5000,

  enableAnimations: true,
  enableBackgroundRefresh: true,
  refreshInterval: 30000,

  searchResultsPerPage: 20,
  defaultSearchSort: 'title',

  calendarStartDay: 0,
  calendarDefaultView: 'month',

  enableDebugMode: false,
  enableBetaFeatures: false,
  maxConcurrentDownloads: 3,
};

// Persistence layer class
export class PersistenceLayer {
  private dispatch: AppDispatch;
  private getState: () => RootState;
  private preferences: UserPreferences = { ...defaultPreferences };
  private appState: PersistedAppState = {
    lastActiveTab: 'movies',
    searchHistory: [],
    recentlyViewedMovies: [],
    bookmarks: [],
    dashboardLayout: [],
    filterPresets: {},
    sortPreferences: {},
  };
  private sessionData: SessionData = {
    currentPage: '/',
    scrollPositions: {},
    formDrafts: {},
    expandedSections: {},
    activeFilters: {},
    selectedItems: {},
  };

  constructor(dispatch: AppDispatch, getState: () => RootState) {
    this.dispatch = dispatch;
    this.getState = getState;
    this.loadPersistedData();
  }

  // Load all persisted data
  private async loadPersistedData() {
    try {
      // Load user preferences
      const savedPreferences = await cacheManager.getUserPreference<UserPreferences>('preferences');
      if (savedPreferences) {
        this.preferences = { ...defaultPreferences, ...savedPreferences };
      }

      // Load app state
      const savedAppState = await cacheManager.loadAppState();
      if (savedAppState) {
        this.appState = { ...this.appState, ...savedAppState };
      }

      // Load session data
      const savedSessionData = await cacheManager.retrieve<SessionData>('session-data', {
        storageType: 'sessionStorage',
      });
      if (savedSessionData) {
        this.sessionData = { ...this.sessionData, ...savedSessionData };
      }

      console.log('Persistence layer loaded successfully');
    } catch (error) {
      console.error('Failed to load persisted data:', error);
    }
  }

  // User preferences methods
  async getPreferences(): Promise<UserPreferences> {
    return { ...this.preferences };
  }

  async updatePreferences(updates: Partial<UserPreferences>) {
    this.preferences = { ...this.preferences, ...updates };
    await cacheManager.setUserPreference('preferences', this.preferences);

    // Dispatch preference update event
    this.dispatch({ type: 'ui/setPreferences', payload: this.preferences });
  }

  async getPreference<K extends keyof UserPreferences>(key: K): Promise<UserPreferences[K]> {
    return this.preferences[key];
  }

  async setPreference<K extends keyof UserPreferences>(key: K, value: UserPreferences[K]) {
    await this.updatePreferences({ [key]: value } as Partial<UserPreferences>);
  }

  // App state methods
  async getAppState(): Promise<PersistedAppState> {
    return { ...this.appState };
  }

  async updateAppState(updates: Partial<PersistedAppState>) {
    this.appState = { ...this.appState, ...updates };
    await cacheManager.saveAppState(this.appState);
  }

  // Search history management
  async addToSearchHistory(query: string) {
    const history = [...this.appState.searchHistory];
    const existingIndex = history.indexOf(query);

    if (existingIndex !== -1) {
      history.splice(existingIndex, 1);
    }

    history.unshift(query);

    // Keep only last 50 searches
    if (history.length > 50) {
      history.splice(50);
    }

    await this.updateAppState({ searchHistory: history });
  }

  async getSearchHistory(): Promise<string[]> {
    return [...this.appState.searchHistory];
  }

  async clearSearchHistory() {
    await this.updateAppState({ searchHistory: [] });
  }

  // Recently viewed movies
  async addToRecentlyViewed(movieId: number) {
    const recent = [...this.appState.recentlyViewedMovies];
    const existingIndex = recent.indexOf(movieId);

    if (existingIndex !== -1) {
      recent.splice(existingIndex, 1);
    }

    recent.unshift(movieId);

    // Keep only last 20 movies
    if (recent.length > 20) {
      recent.splice(20);
    }

    await this.updateAppState({ recentlyViewedMovies: recent });
  }

  async getRecentlyViewed(): Promise<number[]> {
    return [...this.appState.recentlyViewedMovies];
  }

  // Bookmark management
  async addBookmark(name: string, url: string) {
    const bookmarks = [...this.appState.bookmarks];
    const bookmark = {
      id: `bookmark-${Date.now()}`,
      name,
      url,
      timestamp: Date.now(),
    };

    bookmarks.push(bookmark);
    await this.updateAppState({ bookmarks });
  }

  async removeBookmark(id: string) {
    const bookmarks = this.appState.bookmarks.filter(b => b.id !== id);
    await this.updateAppState({ bookmarks });
  }

  async getBookmarks() {
    return [...this.appState.bookmarks];
  }

  // Filter presets
  async saveFilterPreset(name: string, filters: unknown, context: string) {
    const presets = { ...this.appState.filterPresets };
    if (!presets[context]) {
      presets[context] = {};
    }
    presets[context][name] = filters;

    await this.updateAppState({ filterPresets: presets });
  }

  async loadFilterPreset(name: string, context: string) {
    return this.appState.filterPresets[context]?.[name] || null;
  }

  async getFilterPresets(context: string) {
    return this.appState.filterPresets[context] || {};
  }

  async deleteFilterPreset(name: string, context: string) {
    const presets = { ...this.appState.filterPresets };
    if (presets[context]) {
      delete presets[context][name];
      await this.updateAppState({ filterPresets: presets });
    }
  }

  // Sort preferences
  async setSortPreference(context: string, key: string, direction: 'asc' | 'desc') {
    const sortPrefs = { ...this.appState.sortPreferences };
    sortPrefs[context] = { key, direction };
    await this.updateAppState({ sortPreferences: sortPrefs });
  }

  async getSortPreference(context: string) {
    return this.appState.sortPreferences[context] || { key: 'title', direction: 'asc' };
  }

  // Session data methods
  async getSessionData(): Promise<SessionData> {
    return { ...this.sessionData };
  }

  async updateSessionData(updates: Partial<SessionData>) {
    this.sessionData = { ...this.sessionData, ...updates };
    await cacheManager.persist('session-data', this.sessionData, {
      storageType: 'sessionStorage',
      ttl: 24 * 60 * 60 * 1000, // 24 hours
    });
  }

  // Scroll position management
  async saveScrollPosition(page: string, position: number) {
    const scrollPositions = { ...this.sessionData.scrollPositions };
    scrollPositions[page] = position;
    await this.updateSessionData({ scrollPositions });
  }

  async getScrollPosition(page: string): Promise<number> {
    return this.sessionData.scrollPositions[page] || 0;
  }

  // Form draft management
  async saveFormDraft(formId: string, data: unknown) {
    const formDrafts = { ...this.sessionData.formDrafts };
    formDrafts[formId] = { data, timestamp: Date.now() };
    await this.updateSessionData({ formDrafts });
  }

  async getFormDraft(formId: string) {
    const draft = this.sessionData.formDrafts[formId];
    if (!draft) return null;

    // Check if draft is too old (1 hour)
    if (Date.now() - draft.timestamp > 60 * 60 * 1000) {
      await this.removeFormDraft(formId);
      return null;
    }

    return draft.data;
  }

  async removeFormDraft(formId: string) {
    const formDrafts = { ...this.sessionData.formDrafts };
    delete formDrafts[formId];
    await this.updateSessionData({ formDrafts });
  }

  // Expanded sections state
  async setExpandedSection(sectionId: string, expanded: boolean) {
    const expandedSections = { ...this.sessionData.expandedSections };
    expandedSections[sectionId] = expanded;
    await this.updateSessionData({ expandedSections });
  }

  async isExpandedSection(sectionId: string): Promise<boolean> {
    return this.sessionData.expandedSections[sectionId] || false;
  }

  // Selected items management
  async setSelectedItems(context: string, items: number[]) {
    const selectedItems = { ...this.sessionData.selectedItems };
    selectedItems[context] = items;
    await this.updateSessionData({ selectedItems });
  }

  async getSelectedItems(context: string): Promise<number[]> {
    return this.sessionData.selectedItems[context] || [];
  }

  async clearSelectedItems(context: string) {
    await this.setSelectedItems(context, []);
  }

  // Offline data management
  async enableOfflineMode() {
    // Cache critical data for offline access
    const criticalQueries = [
      'getMovies',
      'getQualityProfiles',
      'getRootFolders',
      'getSystemStatus',
      'getHealth',
    ];

    for (const queryType of criticalQueries) {
      try {
        const endpoint = radarrApi.endpoints[queryType as keyof typeof radarrApi.endpoints];
        if (endpoint) {
          // This would need to be implemented with actual query execution
          console.log(`Caching ${queryType} for offline access`);
        }
      } catch (error) {
        console.error(`Failed to cache ${queryType}:`, error);
      }
    }
  }

  // Import/Export functionality
  async exportData(): Promise<string> {
    const exportData = {
      preferences: this.preferences,
      appState: this.appState,
      timestamp: Date.now(),
      version: '1.0.0',
    };

    return JSON.stringify(exportData, null, 2);
  }

  async importData(jsonData: string): Promise<void> {
    try {
      const importData = JSON.parse(jsonData);

      // Validate data structure
      if (!importData.preferences || !importData.appState) {
        throw new Error('Invalid import data format');
      }

      // Merge with existing data
      await this.updatePreferences(importData.preferences);
      await this.updateAppState(importData.appState);

      console.log('Data imported successfully');
    } catch (error) {
      console.error('Failed to import data:', error);
      throw error;
    }
  }

  // Clear all persisted data
  async clearAllData() {
    this.preferences = { ...defaultPreferences };
    this.appState = {
      lastActiveTab: 'movies',
      searchHistory: [],
      recentlyViewedMovies: [],
      bookmarks: [],
      dashboardLayout: [],
      filterPresets: {},
      sortPreferences: {},
    };
    this.sessionData = {
      currentPage: '/',
      scrollPositions: {},
      formDrafts: {},
      expandedSections: {},
      activeFilters: {},
      selectedItems: {},
    };

    await cacheManager.clearPersisted();
    console.log('All persisted data cleared');
  }

  // Statistics
  getStats() {
    return {
      preferences: Object.keys(this.preferences).length,
      searchHistory: this.appState.searchHistory.length,
      recentlyViewed: this.appState.recentlyViewedMovies.length,
      bookmarks: this.appState.bookmarks.length,
      filterPresets: Object.keys(this.appState.filterPresets).length,
      formDrafts: Object.keys(this.sessionData.formDrafts).length,
      cacheStats: cacheManager.getStats(),
    };
  }
}

// Global persistence layer instance
export let persistenceLayer: PersistenceLayer | null = null;

// Initialize persistence layer
export function initializePersistenceLayer(dispatch: AppDispatch, getState: () => RootState) {
  persistenceLayer = new PersistenceLayer(dispatch, getState);
  return persistenceLayer;
}

// Utility hooks for using persistence in components
export const persistenceUtils = {
  // Save preference shortcut
  savePreference: async <K extends keyof UserPreferences>(key: K, value: UserPreferences[K]) => {
    if (persistenceLayer) {
      await persistenceLayer.setPreference(key, value);
    }
  },

  // Get preference shortcut
  getPreference: async <K extends keyof UserPreferences>(key: K): Promise<UserPreferences[K] | null> => {
    if (persistenceLayer) {
      return await persistenceLayer.getPreference(key);
    }
    return null;
  },

  // Auto-save form drafts
  setupAutoSave: (formId: string, getFormData: () => unknown, interval = 30000) => {
    let autoSaveTimer: NodeJS.Timeout;

    const save = async () => {
      if (persistenceLayer) {
        const data = getFormData();
        if (data && Object.keys(data).length > 0) {
          await persistenceLayer.saveFormDraft(formId, data);
        }
      }
    };

    const startAutoSave = () => {
      autoSaveTimer = setInterval(save, interval);
    };

    const stopAutoSave = () => {
      if (autoSaveTimer) {
        clearInterval(autoSaveTimer);
      }
    };

    const clearDraft = async () => {
      if (persistenceLayer) {
        await persistenceLayer.removeFormDraft(formId);
      }
    };

    return { startAutoSave, stopAutoSave, save, clearDraft };
  },
};

export default PersistenceLayer;
