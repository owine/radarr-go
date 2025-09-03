import { radarrApi } from '../store/api/radarrApi';
import type { AppDispatch } from '../store';

// Cache invalidation strategies and utilities
export interface InvalidationRule {
  triggerTags: string[];
  invalidateTags: string[];
  condition?: (data?: any) => boolean;
  delay?: number; // Delay in milliseconds before invalidation
}

export interface InvalidationStrategy {
  name: string;
  rules: InvalidationRule[];
  description: string;
}

// Comprehensive invalidation strategies for different operations
export const invalidationStrategies: Record<string, InvalidationStrategy> = {
  // Movie operations
  movieUpdate: {
    name: 'Movie Update',
    description: 'Invalidate related caches when a movie is updated',
    rules: [
      {
        triggerTags: ['Movie'],
        invalidateTags: ['Movie', 'WantedMovie', 'Calendar', 'Collection'],
      },
    ],
  },

  movieDelete: {
    name: 'Movie Delete',
    description: 'Invalidate all related caches when a movie is deleted',
    rules: [
      {
        triggerTags: ['Movie'],
        invalidateTags: [
          'Movie',
          'MovieFile',
          'WantedMovie',
          'Calendar',
          'Collection',
          'History',
          'Activity',
          'Config',
        ],
      },
    ],
  },

  movieFileImport: {
    name: 'Movie File Import',
    description: 'Update caches when a movie file is imported',
    rules: [
      {
        triggerTags: ['MovieFile'],
        invalidateTags: ['Movie', 'MovieFile', 'WantedMovie', 'History', 'Activity'],
      },
    ],
  },

  // Download and Queue operations
  downloadComplete: {
    name: 'Download Complete',
    description: 'Update caches when a download completes',
    rules: [
      {
        triggerTags: ['Queue'],
        invalidateTags: ['Queue', 'Movie', 'MovieFile', 'History', 'Activity', 'WantedMovie'],
        delay: 5000, // Allow time for file processing
      },
    ],
  },

  queueUpdate: {
    name: 'Queue Update',
    description: 'Update queue-related caches',
    rules: [
      {
        triggerTags: ['Queue'],
        invalidateTags: ['Queue', 'Activity'],
      },
    ],
  },

  releaseGrab: {
    name: 'Release Grab',
    description: 'Update caches when a release is grabbed',
    rules: [
      {
        triggerTags: ['Release'],
        invalidateTags: ['Queue', 'Activity', 'History'],
      },
    ],
  },

  // Configuration operations
  qualityProfileUpdate: {
    name: 'Quality Profile Update',
    description: 'Update related caches when quality profiles change',
    rules: [
      {
        triggerTags: ['QualityProfile'],
        invalidateTags: ['QualityProfile', 'Movie', 'Config'],
      },
    ],
  },

  indexerUpdate: {
    name: 'Indexer Update',
    description: 'Update caches when indexers are modified',
    rules: [
      {
        triggerTags: ['Indexer'],
        invalidateTags: ['Indexer', 'Release', 'Config'],
      },
    ],
  },

  downloadClientUpdate: {
    name: 'Download Client Update',
    description: 'Update caches when download clients are modified',
    rules: [
      {
        triggerTags: ['DownloadClient'],
        invalidateTags: ['DownloadClient', 'Queue', 'Config'],
      },
    ],
  },

  importListUpdate: {
    name: 'Import List Update',
    description: 'Update caches when import lists are modified',
    rules: [
      {
        triggerTags: ['ImportList'],
        invalidateTags: ['ImportList', 'Movie', 'Config'],
        delay: 2000, // Allow time for list processing
      },
    ],
  },

  // Notification and system operations
  notificationUpdate: {
    name: 'Notification Update',
    description: 'Update notification-related caches',
    rules: [
      {
        triggerTags: ['Notification'],
        invalidateTags: ['Notification', 'Config'],
      },
    ],
  },

  systemConfigUpdate: {
    name: 'System Config Update',
    description: 'Update system configuration caches',
    rules: [
      {
        triggerTags: ['Config'],
        invalidateTags: ['Config', 'SystemStatus', 'Health'],
      },
    ],
  },

  healthUpdate: {
    name: 'Health Update',
    description: 'Update health-related caches',
    rules: [
      {
        triggerTags: ['Health'],
        invalidateTags: ['Health', 'SystemResource'],
      },
    ],
  },

  // Batch operations
  bulkMovieOperation: {
    name: 'Bulk Movie Operation',
    description: 'Update caches for bulk movie operations',
    rules: [
      {
        triggerTags: ['Movie'],
        invalidateTags: [
          'Movie',
          'WantedMovie',
          'Calendar',
          'Collection',
          'Activity',
          'History',
        ],
        delay: 3000,
      },
    ],
  },

  // Collection operations
  collectionUpdate: {
    name: 'Collection Update',
    description: 'Update collection-related caches',
    rules: [
      {
        triggerTags: ['Collection'],
        invalidateTags: ['Collection', 'Movie'],
      },
    ],
  },

  // Task and activity operations
  taskComplete: {
    name: 'Task Complete',
    description: 'Update caches when tasks complete',
    rules: [
      {
        triggerTags: ['Command'],
        invalidateTags: ['Activity', 'Movie', 'Queue', 'Health'],
        condition: (data) => data?.status === 'completed',
        delay: 1000,
      },
    ],
  },
};

// Smart cache invalidation manager
export class CacheInvalidationManager {
  private dispatch: AppDispatch;
  private activeStrategies: Set<string> = new Set();
  private invalidationTimers: Map<string, NodeJS.Timeout> = new Map();

  constructor(dispatch: AppDispatch) {
    this.dispatch = dispatch;
    this.loadDefaultStrategies();
  }

  // Load default invalidation strategies
  private loadDefaultStrategies() {
    Object.keys(invalidationStrategies).forEach(strategyName => {
      this.activeStrategies.add(strategyName);
    });
  }

  // Execute invalidation for a specific operation
  invalidateFor(operation: string, data?: any) {
    const strategy = invalidationStrategies[operation];
    if (!strategy || !this.activeStrategies.has(operation)) {
      return;
    }

    strategy.rules.forEach(rule => {
      // Check condition if provided
      if (rule.condition && !rule.condition(data)) {
        return;
      }

      const executeInvalidation = () => {
        console.log(`Invalidating cache for ${operation}:`, rule.invalidateTags);
        this.dispatch(radarrApi.util.invalidateTags(rule.invalidateTags));
      };

      // Apply delay if specified
      if (rule.delay && rule.delay > 0) {
        const timerId = setTimeout(executeInvalidation, rule.delay);
        this.invalidationTimers.set(`${operation}-${Date.now()}`, timerId);
      } else {
        executeInvalidation();
      }
    });
  }

  // Batch invalidation for multiple operations
  batchInvalidate(operations: Array<{ operation: string; data?: any }>) {
    const allTagsToInvalidate = new Set<string>();
    let maxDelay = 0;

    operations.forEach(({ operation, data }) => {
      const strategy = invalidationStrategies[operation];
      if (!strategy || !this.activeStrategies.has(operation)) {
        return;
      }

      strategy.rules.forEach(rule => {
        if (rule.condition && !rule.condition(data)) {
          return;
        }

        rule.invalidateTags.forEach(tag => allTagsToInvalidate.add(tag));
        maxDelay = Math.max(maxDelay, rule.delay || 0);
      });
    });

    const executeInvalidation = () => {
      const tagsArray = Array.from(allTagsToInvalidate);
      console.log('Batch invalidating cache for operations:', operations.map(op => op.operation), 'Tags:', tagsArray);
      this.dispatch(radarrApi.util.invalidateTags(tagsArray));
    };

    if (maxDelay > 0) {
      const timerId = setTimeout(executeInvalidation, maxDelay);
      this.invalidationTimers.set(`batch-${Date.now()}`, timerId);
    } else {
      executeInvalidation();
    }
  }

  // Smart invalidation based on data relationships
  smartInvalidate(changedData: { type: string; id?: number; action: 'create' | 'update' | 'delete' }) {
    const { type, id, action } = changedData;

    // Define relationship mappings
    const relationships: Record<string, string[]> = {
      Movie: ['MovieFile', 'WantedMovie', 'Calendar', 'Collection', 'History'],
      QualityProfile: ['Movie', 'Config'],
      Indexer: ['Release', 'Config'],
      DownloadClient: ['Queue', 'Config'],
      ImportList: ['Movie', 'Config'],
      Collection: ['Movie'],
      Tag: ['Movie', 'QualityProfile', 'Indexer', 'DownloadClient', 'ImportList', 'Notification'],
    };

    const tagsToInvalidate = [type];
    
    // Add related tags based on relationships
    if (relationships[type]) {
      tagsToInvalidate.push(...relationships[type]);
    }

    // Add specific invalidation logic based on action
    if (action === 'delete') {
      // For deletions, invalidate more broadly
      tagsToInvalidate.push('Config', 'SystemStatus');
    }

    // Execute invalidation
    console.log(`Smart invalidation for ${type} ${action}:`, tagsToInvalidate);
    this.dispatch(radarrApi.util.invalidateTags(tagsToInvalidate));

    // If specific ID provided, also invalidate specific entity
    if (id) {
      this.dispatch(radarrApi.util.invalidateTags([{ type, id }]));
    }
  }

  // Conditional invalidation based on application state
  conditionalInvalidate(condition: () => boolean, tags: string[], delay = 0) {
    if (!condition()) {
      return;
    }

    const executeInvalidation = () => {
      console.log('Conditional invalidation:', tags);
      this.dispatch(radarrApi.util.invalidateTags(tags));
    };

    if (delay > 0) {
      const timerId = setTimeout(executeInvalidation, delay);
      this.invalidationTimers.set(`conditional-${Date.now()}`, timerId);
    } else {
      executeInvalidation();
    }
  }

  // Time-based invalidation (for data that becomes stale over time)
  scheduleInvalidation(tags: string[], delay: number, recurring = false) {
    const executeInvalidation = () => {
      console.log('Scheduled invalidation:', tags);
      this.dispatch(radarrApi.util.invalidateTags(tags));

      if (recurring) {
        this.scheduleInvalidation(tags, delay, true);
      }
    };

    const timerId = setTimeout(executeInvalidation, delay);
    const timerKey = `scheduled-${Date.now()}`;
    this.invalidationTimers.set(timerKey, timerId);

    return () => {
      const timer = this.invalidationTimers.get(timerKey);
      if (timer) {
        clearTimeout(timer);
        this.invalidationTimers.delete(timerKey);
      }
    };
  }

  // Enable/disable specific strategies
  enableStrategy(strategyName: string) {
    this.activeStrategies.add(strategyName);
  }

  disableStrategy(strategyName: string) {
    this.activeStrategies.delete(strategyName);
  }

  // Get active strategies
  getActiveStrategies(): string[] {
    return Array.from(this.activeStrategies);
  }

  // Clear all pending invalidation timers
  clearPendingInvalidations() {
    this.invalidationTimers.forEach(timer => clearTimeout(timer));
    this.invalidationTimers.clear();
  }

  // Get statistics about invalidation usage
  getStats() {
    return {
      activeStrategies: this.activeStrategies.size,
      totalStrategies: Object.keys(invalidationStrategies).length,
      pendingInvalidations: this.invalidationTimers.size,
      strategies: Array.from(this.activeStrategies),
    };
  }

  // Cleanup method
  destroy() {
    this.clearPendingInvalidations();
    this.activeStrategies.clear();
  }
}

// Global instance (will be initialized in store setup)
export let cacheInvalidationManager: CacheInvalidationManager | null = null;

// Initialize the cache invalidation manager
export function initializeCacheInvalidation(dispatch: AppDispatch) {
  cacheInvalidationManager = new CacheInvalidationManager(dispatch);
  return cacheInvalidationManager;
}

// Utility functions for common invalidation patterns
export const invalidationUtils = {
  // Invalidate after successful mutation
  afterMutation: (operation: string, data?: any) => {
    if (cacheInvalidationManager) {
      cacheInvalidationManager.invalidateFor(operation, data);
    }
  },

  // Invalidate based on WebSocket events
  onWebSocketEvent: (eventType: string, eventData: any) => {
    if (!cacheInvalidationManager) return;

    const eventToOperationMap: Record<string, string> = {
      'QueueUpdate': 'queueUpdate',
      'MovieUpdate': 'movieUpdate',
      'DownloadComplete': 'downloadComplete',
      'HealthUpdate': 'healthUpdate',
      'TaskComplete': 'taskComplete',
    };

    const operation = eventToOperationMap[eventType];
    if (operation) {
      cacheInvalidationManager.invalidateFor(operation, eventData);
    }
  },

  // Bulk operations
  onBulkOperation: (operationType: string, items: any[]) => {
    if (!cacheInvalidationManager) return;

    cacheInvalidationManager.batchInvalidate(
      items.map(item => ({ operation: operationType, data: item }))
    );
  },

  // Smart invalidation for form updates
  onFormUpdate: (entityType: string, entityId?: number) => {
    if (!cacheInvalidationManager) return;

    cacheInvalidationManager.smartInvalidate({
      type: entityType,
      id: entityId,
      action: 'update',
    });
  },

  // Periodic refresh for time-sensitive data
  setupPeriodicRefresh: (tags: string[], intervalMs: number) => {
    if (!cacheInvalidationManager) return () => {};

    return cacheInvalidationManager.scheduleInvalidation(tags, intervalMs, true);
  },
};

export default CacheInvalidationManager;