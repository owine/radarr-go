import { useEffect, useRef, useState, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import type {
  QueryDefinition,
  QueryHooks,
  QueryActionCreatorResult,
  QueryDefinition as RTKQueryDefinition
} from '@reduxjs/toolkit/query/react';
import { radarrApi } from '../store/api/radarrApi';
import { cacheManager, cacheUtils } from '../utils/cacheManager';
import { useWebSocketConnection } from '../store/middleware/websocketMiddleware';

// Enhanced query options
export interface EnhancedQueryOptions {
  // Polling options
  pollingInterval?: number;
  skipPollingWhenHidden?: boolean;

  // Retry options
  maxRetries?: number;
  retryDelay?: number;
  exponentialBackoff?: boolean;

  // Cache options
  cacheTime?: number;
  staleTime?: number;
  useOfflineCache?: boolean;

  // Background updates
  refetchOnMount?: boolean;
  refetchOnFocus?: boolean;
  refetchOnReconnect?: boolean;
  refetchOnVisibilityChange?: boolean;

  // Optimistic updates
  optimisticUpdate?: boolean;
  optimisticData?: any;

  // Real-time updates
  subscribeToUpdates?: boolean;
  updateEvents?: string[];

  // Error handling
  onError?: (error: any) => void;
  onSuccess?: (data: any) => void;

  // Transform data
  select?: (data: any) => any;
  transformError?: (error: any) => any;
}

// Enhanced query result
export interface EnhancedQueryResult<T> {
  data: T | undefined;
  error: any;
  isLoading: boolean;
  isFetching: boolean;
  isSuccess: boolean;
  isError: boolean;
  isStale: boolean;
  isOffline: boolean;
  lastFetchedAt?: Date;

  // Actions
  refetch: () => Promise<any>;
  invalidate: () => void;
  reset: () => void;

  // Status helpers
  hasData: boolean;
  isEmpty: boolean;

  // Retry functionality
  retry: () => void;
  retryCount: number;
  canRetry: boolean;
}

// Enhanced mutation options
export interface EnhancedMutationOptions<T> {
  // Optimistic updates
  optimisticUpdate?: (currentData: any, variables: any) => any;
  rollbackOnError?: boolean;

  // Cache invalidation
  invalidatesTags?: string[];
  updatesCaches?: Array<{
    queryKey: string;
    updater: (currentData: any, result: T, variables: any) => any;
  }>;

  // UI feedback
  showSuccessMessage?: boolean | string;
  showErrorMessage?: boolean | string;

  // Callbacks
  onMutate?: (variables: any) => void;
  onSuccess?: (result: T, variables: any) => void;
  onError?: (error: any, variables: any) => void;
  onSettled?: (result: T | undefined, error: any, variables: any) => void;
}

// Enhanced mutation result
export interface EnhancedMutationResult<T, V> {
  mutate: (variables: V) => Promise<T>;
  mutateAsync: (variables: V) => Promise<T>;
  reset: () => void;

  data: T | undefined;
  error: any;
  isLoading: boolean;
  isSuccess: boolean;
  isError: boolean;
  isIdle: boolean;

  // Status helpers
  hasData: boolean;
  canMutate: boolean;
}

// Main enhanced query hook
export function useEnhancedQuery<T>(
  queryKey: string,
  queryFn: () => Promise<T>,
  options: EnhancedQueryOptions = {}
): EnhancedQueryResult<T> {
  const dispatch = useDispatch();
  const wsConnection = useWebSocketConnection();

  // State management
  const [data, setData] = useState<T | undefined>(undefined);
  const [error, setError] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isFetching, setIsFetching] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [lastFetchedAt, setLastFetchedAt] = useState<Date | undefined>();
  const [isStale, setIsStale] = useState(false);

  // Refs for managing state and cleanup
  const abortControllerRef = useRef<AbortController | null>(null);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const staleTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const isOnlineRef = useRef(navigator.onLine);

  // Computed states
  const isSuccess = !isLoading && !error && data !== undefined;
  const isError = !isLoading && error !== null;
  const hasData = data !== undefined;
  const isEmpty = hasData && (Array.isArray(data) ? data.length === 0 : Object.keys(data as any).length === 0);
  const canRetry = isError && retryCount < (options.maxRetries || 3);
  const isOffline = !isOnlineRef.current;

  // Execute query function
  const executeQuery = useCallback(async (showLoading = true): Promise<void> => {
    // Cancel previous request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new abort controller
    abortControllerRef.current = new AbortController();

    if (showLoading) {
      setIsFetching(true);
      if (data === undefined) {
        setIsLoading(true);
      }
    }

    setError(null);

    try {
      // Check offline cache first if enabled
      if (options.useOfflineCache && isOffline) {
        const cachedData = cacheManager.get<T>(queryKey);
        if (cachedData !== null) {
          setData(options.select ? options.select(cachedData) : cachedData);
          setIsLoading(false);
          setIsFetching(false);
          return;
        }
      }

      // Execute the query
      const result = await queryFn();

      // Check if request was aborted
      if (abortControllerRef.current?.signal.aborted) {
        return;
      }

      // Transform data if select function provided
      const transformedData = options.select ? options.select(result) : result;

      // Update state
      setData(transformedData);
      setError(null);
      setRetryCount(0);
      setLastFetchedAt(new Date());
      setIsLoading(false);
      setIsFetching(false);

      // Cache the result
      if (options.useOfflineCache) {
        cacheManager.set(queryKey, result, { ttl: options.cacheTime || 5 * 60 * 1000 });
      }

      // Set up stale timeout
      if (options.staleTime) {
        staleTimeoutRef.current = setTimeout(() => {
          setIsStale(true);
        }, options.staleTime);
      }

      // Call success callback
      if (options.onSuccess) {
        options.onSuccess(transformedData);
      }

    } catch (err: any) {
      // Check if request was aborted
      if (abortControllerRef.current?.signal.aborted) {
        return;
      }

      const transformedError = options.transformError ? options.transformError(err) : err;
      setError(transformedError);
      setIsLoading(false);
      setIsFetching(false);

      // Call error callback
      if (options.onError) {
        options.onError(transformedError);
      }

      // Auto retry if configured
      if (retryCount < (options.maxRetries || 3)) {
        const delay = options.exponentialBackoff
          ? (options.retryDelay || 1000) * Math.pow(2, retryCount)
          : (options.retryDelay || 1000);

        retryTimeoutRef.current = setTimeout(() => {
          setRetryCount(prev => prev + 1);
          executeQuery(false);
        }, delay);
      }
    }
  }, [queryKey, queryFn, options, retryCount, data, isOffline]);

  // Manual refetch function
  const refetch = useCallback(async () => {
    setIsStale(false);
    return executeQuery(true);
  }, [executeQuery]);

  // Manual retry function
  const retry = useCallback(() => {
    if (canRetry) {
      executeQuery(false);
    }
  }, [canRetry, executeQuery]);

  // Invalidate query
  const invalidate = useCallback(() => {
    setIsStale(true);
    if (options.refetchOnMount !== false) {
      refetch();
    }
  }, [refetch, options.refetchOnMount]);

  // Reset query state
  const reset = useCallback(() => {
    setData(undefined);
    setError(null);
    setIsLoading(true);
    setIsFetching(false);
    setRetryCount(0);
    setLastFetchedAt(undefined);
    setIsStale(false);

    // Clear timeouts
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
    }
    if (staleTimeoutRef.current) {
      clearTimeout(staleTimeoutRef.current);
    }
  }, []);

  // Set up polling
  useEffect(() => {
    if (options.pollingInterval && options.pollingInterval > 0) {
      pollingIntervalRef.current = setInterval(() => {
        // Skip polling if document is hidden and skipPollingWhenHidden is true
        if (options.skipPollingWhenHidden && document.hidden) {
          return;
        }

        executeQuery(false);
      }, options.pollingInterval);
    }

    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
    };
  }, [options.pollingInterval, options.skipPollingWhenHidden, executeQuery]);

  // Set up real-time subscriptions
  useEffect(() => {
    if (options.subscribeToUpdates && options.updateEvents) {
      const unsubscribers = options.updateEvents.map(eventType =>
        wsConnection.subscribe(eventType, () => {
          refetch();
        })
      );

      return () => {
        unsubscribers.forEach(unsubscribe => unsubscribe());
      };
    }
  }, [options.subscribeToUpdates, options.updateEvents, wsConnection, refetch]);

  // Handle focus/visibility changes
  useEffect(() => {
    const handleFocus = () => {
      if (options.refetchOnFocus) {
        refetch();
      }
    };

    const handleVisibilityChange = () => {
      if (options.refetchOnVisibilityChange && !document.hidden) {
        refetch();
      }
    };

    const handleOnline = () => {
      isOnlineRef.current = true;
      if (options.refetchOnReconnect) {
        refetch();
      }
    };

    const handleOffline = () => {
      isOnlineRef.current = false;
    };

    if (options.refetchOnFocus) {
      window.addEventListener('focus', handleFocus);
    }

    if (options.refetchOnVisibilityChange) {
      document.addEventListener('visibilitychange', handleVisibilityChange);
    }

    if (options.refetchOnReconnect) {
      window.addEventListener('online', handleOnline);
      window.addEventListener('offline', handleOffline);
    }

    return () => {
      window.removeEventListener('focus', handleFocus);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [options.refetchOnFocus, options.refetchOnVisibilityChange, options.refetchOnReconnect, refetch]);

  // Initial fetch
  useEffect(() => {
    if (options.refetchOnMount !== false) {
      executeQuery();
    }
  }, []);

  // Cleanup
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
      }
      if (staleTimeoutRef.current) {
        clearTimeout(staleTimeoutRef.current);
      }
    };
  }, []);

  return {
    data,
    error,
    isLoading,
    isFetching,
    isSuccess,
    isError,
    isStale,
    isOffline,
    lastFetchedAt,
    refetch,
    invalidate,
    reset,
    hasData,
    isEmpty,
    retry,
    retryCount,
    canRetry,
  };
}

// Enhanced mutation hook
export function useEnhancedMutation<T, V>(
  mutationFn: (variables: V) => Promise<T>,
  options: EnhancedMutationOptions<T> = {}
): EnhancedMutationResult<T, V> {
  const dispatch = useDispatch();

  const [data, setData] = useState<T | undefined>(undefined);
  const [error, setError] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isIdle, setIsIdle] = useState(true);

  const isSuccess = !isLoading && !error && data !== undefined;
  const isError = !isLoading && error !== null;
  const hasData = data !== undefined;
  const canMutate = !isLoading;

  const mutate = useCallback(async (variables: V): Promise<T> => {
    setIsLoading(true);
    setIsIdle(false);
    setError(null);

    // Call onMutate callback
    if (options.onMutate) {
      options.onMutate(variables);
    }

    // Store original data for potential rollback
    const originalCacheState = new Map();

    try {
      // Apply optimistic updates
      if (options.optimisticUpdate && options.updatesCaches) {
        options.updatesCaches.forEach(({ queryKey, updater }) => {
          const currentData = cacheManager.get(queryKey);
          if (currentData !== null) {
            originalCacheState.set(queryKey, currentData);
            const optimisticData = options.optimisticUpdate!(currentData, variables);
            cacheManager.set(queryKey, optimisticData);
          }
        });
      }

      // Execute mutation
      const result = await mutationFn(variables);

      // Update state
      setData(result);
      setIsLoading(false);

      // Update caches with real data
      if (options.updatesCaches) {
        options.updatesCaches.forEach(({ queryKey, updater }) => {
          const currentData = cacheManager.get(queryKey);
          if (currentData !== null) {
            const updatedData = updater(currentData, result, variables);
            cacheManager.set(queryKey, updatedData);
          }
        });
      }

      // Invalidate cache tags
      if (options.invalidatesTags) {
        cacheManager.clearByTags(options.invalidatesTags);
      }

      // Call success callback
      if (options.onSuccess) {
        options.onSuccess(result, variables);
      }

      return result;

    } catch (err: any) {
      setError(err);
      setIsLoading(false);

      // Rollback optimistic updates
      if (options.rollbackOnError && originalCacheState.size > 0) {
        originalCacheState.forEach((originalData, queryKey) => {
          cacheManager.set(queryKey, originalData);
        });
      }

      // Call error callback
      if (options.onError) {
        options.onError(err, variables);
      }

      throw err;

    } finally {
      // Call settled callback
      if (options.onSettled) {
        options.onSettled(data, error, variables);
      }
    }
  }, [mutationFn, options, data, error]);

  const mutateAsync = useCallback((variables: V) => {
    return mutate(variables);
  }, [mutate]);

  const reset = useCallback(() => {
    setData(undefined);
    setError(null);
    setIsLoading(false);
    setIsIdle(true);
  }, []);

  return {
    mutate,
    mutateAsync,
    reset,
    data,
    error,
    isLoading,
    isSuccess,
    isError,
    isIdle,
    hasData,
    canMutate,
  };
}

// Utility hooks for common patterns
export const useInfiniteQuery = <T>(
  queryKeyFactory: (page: number) => string,
  queryFnFactory: (page: number) => () => Promise<T[]>,
  options: EnhancedQueryOptions & { pageSize?: number } = {}
) => {
  const [pages, setPages] = useState<T[][]>([]);
  const [hasNextPage, setHasNextPage] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const pageSize = options.pageSize || 20;

  // Load first page
  const firstPageQuery = useEnhancedQuery(
    queryKeyFactory(1),
    queryFnFactory(1),
    options
  );

  useEffect(() => {
    if (firstPageQuery.data) {
      setPages([firstPageQuery.data]);
      setHasNextPage(firstPageQuery.data.length === pageSize);
    }
  }, [firstPageQuery.data, pageSize]);

  const loadMore = useCallback(async () => {
    if (!hasNextPage || isLoadingMore) return;

    setIsLoadingMore(true);
    try {
      const nextPage = pages.length + 1;
      const nextPageData = await queryFnFactory(nextPage)();

      setPages(prev => [...prev, nextPageData]);
      setHasNextPage(nextPageData.length === pageSize);
    } catch (error) {
      console.error('Failed to load more data:', error);
    } finally {
      setIsLoadingMore(false);
    }
  }, [hasNextPage, isLoadingMore, pages.length, pageSize, queryFnFactory]);

  return {
    data: pages.flat(),
    pages,
    error: firstPageQuery.error,
    isLoading: firstPageQuery.isLoading,
    isFetching: firstPageQuery.isFetching,
    isLoadingMore,
    hasNextPage,
    loadMore,
    refetch: firstPageQuery.refetch,
    reset: () => {
      setPages([]);
      setHasNextPage(true);
      setIsLoadingMore(false);
      firstPageQuery.reset();
    },
  };
};

export default useEnhancedQuery;
