// Advanced caching and persistence utilities for Radarr Go frontend

export interface CacheItem<T = unknown> {
  data: T;
  timestamp: number;
  ttl: number; // Time to live in milliseconds
  version: string;
  tags: string[];
}

export interface CacheOptions {
  ttl?: number; // Default 5 minutes
  version?: string;
  tags?: string[];
  compress?: boolean;
  encrypt?: boolean;
}

export interface PersistenceOptions extends CacheOptions {
  storageType?: 'localStorage' | 'sessionStorage' | 'indexedDB';
  namespace?: string;
}

class CacheManager {
  private cache: Map<string, CacheItem> = new Map();
  private defaultTTL = 5 * 60 * 1000; // 5 minutes
  private maxSize = 1000;
  private currentVersion = '1.0.0';
  private defaultNamespace = 'radarr-go';

  // In-memory caching
  set<T>(key: string, data: T, options: CacheOptions = {}): void {
    const ttl = options.ttl || this.defaultTTL;
    const item: CacheItem<T> = {
      data,
      timestamp: Date.now(),
      ttl,
      version: options.version || this.currentVersion,
      tags: options.tags || [],
    };

    // Compress data if requested
    if (options.compress && typeof data === 'object') {
      try {
        item.data = this.compress(JSON.stringify(data)) as T;
      } catch (error) {
        console.warn('Failed to compress cache data:', error);
      }
    }

    this.cache.set(key, item);
    this.cleanup();
  }

  get<T>(key: string, options: { decompress?: boolean } = {}): T | null {
    const item = this.cache.get(key);

    if (!item) return null;

    // Check if expired
    if (Date.now() > item.timestamp + item.ttl) {
      this.cache.delete(key);
      return null;
    }

    // Check version compatibility
    if (item.version !== this.currentVersion) {
      this.cache.delete(key);
      return null;
    }

    let data = item.data;

    // Decompress if needed
    if (options.decompress && typeof data === 'string') {
      try {
        data = JSON.parse(this.decompress(data));
      } catch (error) {
        console.warn('Failed to decompress cache data:', error);
        return null;
      }
    }

    return data;
  }

  has(key: string): boolean {
    const item = this.cache.get(key);
    if (!item) return false;

    // Check if expired
    if (Date.now() > item.timestamp + item.ttl) {
      this.cache.delete(key);
      return false;
    }

    return true;
  }

  delete(key: string): boolean {
    return this.cache.delete(key);
  }

  clear(): void {
    this.cache.clear();
  }

  // Clear cache items by tags
  clearByTags(tags: string[]): number {
    let deletedCount = 0;

    for (const [key, item] of this.cache.entries()) {
      if (tags.some(tag => item.tags.includes(tag))) {
        this.cache.delete(key);
        deletedCount++;
      }
    }

    return deletedCount;
  }

  // Get cache statistics
  getStats() {
    const now = Date.now();
    let validItems = 0;
    let expiredItems = 0;
    let totalSize = 0;

    for (const [, item] of this.cache.entries()) {
      if (now > item.timestamp + item.ttl) {
        expiredItems++;
      } else {
        validItems++;
      }

      totalSize += this.getItemSize(item);
    }

    return {
      totalItems: this.cache.size,
      validItems,
      expiredItems,
      totalSize: this.formatBytes(totalSize),
      hitRate: this.getHitRate(),
    };
  }

  // Persistence methods
  async persist<T>(key: string, data: T, options: PersistenceOptions = {}): Promise<void> {
    const storageType = options.storageType || 'localStorage';
    const namespace = options.namespace || this.defaultNamespace;
    const fullKey = `${namespace}:${key}`;

    const item: CacheItem<T> = {
      data,
      timestamp: Date.now(),
      ttl: options.ttl || this.defaultTTL,
      version: options.version || this.currentVersion,
      tags: options.tags || [],
    };

    try {
      let serializedData = JSON.stringify(item);

      // Compress if requested
      if (options.compress) {
        serializedData = this.compress(serializedData);
      }

      // Encrypt if requested (basic encryption - not for sensitive data)
      if (options.encrypt) {
        serializedData = this.encrypt(serializedData);
      }

      switch (storageType) {
        case 'localStorage':
          localStorage.setItem(fullKey, serializedData);
          break;
        case 'sessionStorage':
          sessionStorage.setItem(fullKey, serializedData);
          break;
        case 'indexedDB':
          await this.setIndexedDB(fullKey, serializedData);
          break;
      }
    } catch (error) {
      console.error('Failed to persist cache item:', error);
      throw error;
    }
  }

  async retrieve<T>(key: string, options: PersistenceOptions = {}): Promise<T | null> {
    const storageType = options.storageType || 'localStorage';
    const namespace = options.namespace || this.defaultNamespace;
    const fullKey = `${namespace}:${key}`;

    try {
      let serializedData: string | null = null;

      switch (storageType) {
        case 'localStorage':
          serializedData = localStorage.getItem(fullKey);
          break;
        case 'sessionStorage':
          serializedData = sessionStorage.getItem(fullKey);
          break;
        case 'indexedDB':
          serializedData = await this.getIndexedDB(fullKey);
          break;
      }

      if (!serializedData) return null;

      // Decrypt if encrypted
      if (options.encrypt) {
        serializedData = this.decrypt(serializedData);
      }

      // Decompress if compressed
      if (options.compress) {
        serializedData = this.decompress(serializedData);
      }

      const item: CacheItem<T> = JSON.parse(serializedData);

      // Check expiration
      if (Date.now() > item.timestamp + item.ttl) {
        await this.removePersisted(key, options);
        return null;
      }

      // Check version compatibility
      if (item.version !== this.currentVersion) {
        await this.removePersisted(key, options);
        return null;
      }

      return item.data;
    } catch (error) {
      console.error('Failed to retrieve persisted item:', error);
      return null;
    }
  }

  async removePersisted(key: string, options: PersistenceOptions = {}): Promise<void> {
    const storageType = options.storageType || 'localStorage';
    const namespace = options.namespace || this.defaultNamespace;
    const fullKey = `${namespace}:${key}`;

    try {
      switch (storageType) {
        case 'localStorage':
          localStorage.removeItem(fullKey);
          break;
        case 'sessionStorage':
          sessionStorage.removeItem(fullKey);
          break;
        case 'indexedDB':
          await this.removeIndexedDB(fullKey);
          break;
      }
    } catch (error) {
      console.error('Failed to remove persisted item:', error);
    }
  }

  // Clear all persisted items in namespace
  async clearPersisted(options: PersistenceOptions = {}): Promise<void> {
    const storageType = options.storageType || 'localStorage';
    const namespace = options.namespace || this.defaultNamespace;

    try {
      switch (storageType) {
        case 'localStorage':
          Object.keys(localStorage).forEach(key => {
            if (key.startsWith(`${namespace}:`)) {
              localStorage.removeItem(key);
            }
          });
          break;
        case 'sessionStorage':
          Object.keys(sessionStorage).forEach(key => {
            if (key.startsWith(`${namespace}:`)) {
              sessionStorage.removeItem(key);
            }
          });
          break;
        case 'indexedDB':
          await this.clearIndexedDB(namespace);
          break;
      }
    } catch (error) {
      console.error('Failed to clear persisted items:', error);
    }
  }

  // User preferences specific methods
  async setUserPreference<T>(key: string, value: T): Promise<void> {
    await this.persist(`user-prefs:${key}`, value, {
      ttl: 365 * 24 * 60 * 60 * 1000, // 1 year
      storageType: 'localStorage',
      tags: ['user-preferences'],
    });
  }

  async getUserPreference<T>(key: string, defaultValue?: T): Promise<T> {
    const value = await this.retrieve<T>(`user-prefs:${key}`, {
      storageType: 'localStorage',
    });
    return value !== null ? value : (defaultValue as T);
  }

  // App state persistence
  async saveAppState(state: Record<string, unknown>): Promise<void> {
    await this.persist('app-state', state, {
      ttl: 24 * 60 * 60 * 1000, // 24 hours
      storageType: 'localStorage',
      tags: ['app-state'],
      compress: true,
    });
  }

  async loadAppState(): Promise<Record<string, unknown> | null> {
    return await this.retrieve('app-state', {
      storageType: 'localStorage',
      compress: true,
    });
  }

  // Private utility methods
  private cleanup(): void {
    if (this.cache.size <= this.maxSize) return;

    // Sort by timestamp and remove oldest items
    const entries = Array.from(this.cache.entries())
      .sort(([, a], [, b]) => a.timestamp - b.timestamp);

    const itemsToRemove = this.cache.size - this.maxSize + Math.floor(this.maxSize * 0.1);

    for (let i = 0; i < itemsToRemove; i++) {
      this.cache.delete(entries[i][0]);
    }
  }

  private getItemSize(item: CacheItem): number {
    try {
      return new Blob([JSON.stringify(item)]).size;
    } catch {
      return JSON.stringify(item).length * 2; // Approximate size
    }
  }

  private formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  private hitCount = 0;
  private missCount = 0;

  private getHitRate(): number {
    const total = this.hitCount + this.missCount;
    return total > 0 ? (this.hitCount / total) * 100 : 0;
  }

  // Simple compression (using LZ-string-like algorithm)
  private compress(str: string): string {
    // Simple run-length encoding for demonstration
    // In production, use a proper compression library
    return str.replace(/(.)\1+/g, (match, char) => `${char}${match.length}`);
  }

  private decompress(str: string): string {
    // Reverse of simple compression
    return str.replace(/(.)\d+/g, (match, char) => {
      const count = parseInt(match.slice(1));
      return char.repeat(count);
    });
  }

  // Basic encryption (NOT secure - for demo only)
  private encrypt(str: string): string {
    // Simple XOR encryption with fixed key (NOT secure)
    const key = 'radarr-go-key';
    let result = '';

    for (let i = 0; i < str.length; i++) {
      result += String.fromCharCode(
        str.charCodeAt(i) ^ key.charCodeAt(i % key.length)
      );
    }

    return btoa(result);
  }

  private decrypt(str: string): string {
    try {
      const decoded = atob(str);
      const key = 'radarr-go-key';
      let result = '';

      for (let i = 0; i < decoded.length; i++) {
        result += String.fromCharCode(
          decoded.charCodeAt(i) ^ key.charCodeAt(i % key.length)
        );
      }

      return result;
    } catch {
      throw new Error('Failed to decrypt data');
    }
  }

  // IndexedDB utilities
  private async setIndexedDB(key: string, value: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('RadarrGoCache', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const db = request.result;
        const transaction = db.transaction(['cache'], 'readwrite');
        const store = transaction.objectStore('cache');

        store.put({ key, value });
        transaction.oncomplete = () => resolve();
        transaction.onerror = () => reject(transaction.error);
      };

      request.onupgradeneeded = () => {
        const db = request.result;
        if (!db.objectStoreNames.contains('cache')) {
          db.createObjectStore('cache', { keyPath: 'key' });
        }
      };
    });
  }

  private async getIndexedDB(key: string): Promise<string | null> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('RadarrGoCache', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const db = request.result;
        const transaction = db.transaction(['cache'], 'readonly');
        const store = transaction.objectStore('cache');
        const getRequest = store.get(key);

        getRequest.onsuccess = () => {
          resolve(getRequest.result?.value || null);
        };
        getRequest.onerror = () => reject(getRequest.error);
      };
    });
  }

  private async removeIndexedDB(key: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('RadarrGoCache', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const db = request.result;
        const transaction = db.transaction(['cache'], 'readwrite');
        const store = transaction.objectStore('cache');

        store.delete(key);
        transaction.oncomplete = () => resolve();
        transaction.onerror = () => reject(transaction.error);
      };
    });
  }

  private async clearIndexedDB(namespace: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('RadarrGoCache', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const db = request.result;
        const transaction = db.transaction(['cache'], 'readwrite');
        const store = transaction.objectStore('cache');

        const cursorRequest = store.openCursor();
        cursorRequest.onsuccess = (event) => {
          const cursor = (event.target as IDBRequest).result;
          if (cursor) {
            if (cursor.key.toString().startsWith(`${namespace}:`)) {
              cursor.delete();
            }
            cursor.continue();
          }
        };

        transaction.oncomplete = () => resolve();
        transaction.onerror = () => reject(transaction.error);
      };
    });
  }
}

// Global cache manager instance
export const cacheManager = new CacheManager();

// Utility functions for common caching patterns
export const cacheUtils = {
  // Generate cache key with parameters
  generateKey: (prefix: string, params: Record<string, unknown> = {}): string => {
    const sortedParams = Object.keys(params)
      .sort()
      .map(key => `${key}:${params[key]}`)
      .join('|');

    return sortedParams ? `${prefix}:${sortedParams}` : prefix;
  },

  // Cache with automatic retry on stale data
  withFallback: async <T>(
    primaryFn: () => Promise<T>,
    fallbackKey: string,
    options: CacheOptions = {}
  ): Promise<T> => {
    try {
      const result = await primaryFn();
      cacheManager.set(fallbackKey, result, options);
      return result;
    } catch (error) {
      const cached = cacheManager.get<T>(fallbackKey);
      if (cached !== null) {
        console.warn('Using cached data due to primary source failure:', error);
        return cached;
      }
      throw error;
    }
  },

  // Batch cache operations
  setBatch: <T>(items: Array<{ key: string; data: T; options?: CacheOptions }>): void => {
    items.forEach(({ key, data, options }) => {
      cacheManager.set(key, data, options);
    });
  },

  getBatch: <T>(keys: string[]): Array<{ key: string; data: T | null }> => {
    return keys.map(key => ({
      key,
      data: cacheManager.get<T>(key)
    }));
  },

  // Cache middleware for API responses
  cacheApiResponse: <T>(
    key: string,
    apiFn: () => Promise<T>,
    options: CacheOptions = {}
  ): Promise<T> => {
    const cached = cacheManager.get<T>(key);

    if (cached !== null) {
      return Promise.resolve(cached);
    }

    return apiFn().then(result => {
      cacheManager.set(key, result, options);
      return result;
    });
  },
};

export default cacheManager;
