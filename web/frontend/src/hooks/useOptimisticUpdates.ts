import { useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { radarrApi } from '../store/api/radarrApi';
import type { Movie, QueueItem, WantedMovie, Activity, Tag } from '../types/api';

// Optimistic update utilities for common operations
export const useOptimisticUpdates = () => {
  const dispatch = useDispatch();

  // Optimistic movie updates
  const updateMovieOptimistically = useCallback(
    (movieId: number, updates: Partial<Movie>) => {
      // Update the specific movie cache
      dispatch(
        radarrApi.util.updateQueryData('getMovie', movieId, (draft) => {
          Object.assign(draft, updates);
        })
      );

      // Update the movies list cache
      dispatch(
        radarrApi.util.updateQueryData('getMovies', undefined, (draft) => {
          const movieIndex = draft.findIndex((movie: Movie) => movie.id === movieId);
          if (movieIndex !== -1) {
            Object.assign(draft[movieIndex], updates);
          }
        })
      );

      // Update any paginated movies queries
      // This would need to be implemented for each possible query parameter combination
    },
    [dispatch]
  );

  // Optimistic queue item removal
  const removeQueueItemOptimistically = useCallback(
    (queueItemId: number) => {
      // Remove from queue list
      dispatch(
        radarrApi.util.updateQueryData('getQueue', undefined, (draft) => {
          if (draft.records) {
            draft.records = draft.records.filter((item: QueueItem) => item.id !== queueItemId);
            draft.totalRecords = Math.max(0, draft.totalRecords - 1);
          }
        })
      );

      // Update queue stats
      dispatch(
        radarrApi.util.updateQueryData('getQueueStats', undefined, (draft) => {
          if (draft) {
            draft.totalItems = Math.max(0, draft.totalItems - 1);
          }
        })
      );
    },
    [dispatch]
  );

  // Optimistic movie monitoring toggle
  const toggleMovieMonitoringOptimistically = useCallback(
    (movieId: number, monitored: boolean) => {
      const updates = { monitored };
      updateMovieOptimistically(movieId, updates);

      // Also update wanted movies if this affects monitoring
      if (!monitored) {
        // Remove from missing movies if unmonitored
        dispatch(
          radarrApi.util.updateQueryData('getMissingMovies', undefined, (draft) => {
            if (draft.records) {
              draft.records = draft.records.filter(
                (wantedMovie: WantedMovie) => wantedMovie.movieId !== movieId
              );
              draft.totalRecords = Math.max(0, draft.totalRecords - 1);
            }
          })
        );
      }
    },
    [dispatch, updateMovieOptimistically]
  );

  // Optimistic queue item status update
  const updateQueueItemStatusOptimistically = useCallback(
    (queueItemId: number, status: string, progress?: number) => {
      dispatch(
        radarrApi.util.updateQueryData('getQueueItem', queueItemId, (draft) => {
          draft.status = status;
          if (progress !== undefined) {
            draft.sizeleft = Math.max(0, draft.size * (1 - progress / 100));
          }
        })
      );

      dispatch(
        radarrApi.util.updateQueryData('getQueue', undefined, (draft) => {
          if (draft.records) {
            const itemIndex = draft.records.findIndex((item: QueueItem) => item.id === queueItemId);
            if (itemIndex !== -1) {
              draft.records[itemIndex].status = status;
              if (progress !== undefined) {
                draft.records[itemIndex].sizeleft = Math.max(
                  0,
                  draft.records[itemIndex].size * (1 - progress / 100)
                );
              }
            }
          }
        })
      );
    },
    [dispatch]
  );

  // Optimistic activity addition
  const addActivityOptimistically = useCallback(
    (activity: Partial<Activity>) => {
      const newActivity: Activity = {
        id: Date.now(), // Temporary ID
        type: 'Unknown',
        status: 'running',
        progress: 0,
        startTime: new Date().toISOString(),
        message: 'Operation started...',
        data: {},
        ...activity,
      };

      dispatch(
        radarrApi.util.updateQueryData('getActivity', undefined, (draft) => {
          draft.unshift(newActivity);
        })
      );

      dispatch(
        radarrApi.util.updateQueryData('getRunningActivities', undefined, (draft) => {
          draft.unshift(newActivity);
        })
      );
    },
    [dispatch]
  );

  // Optimistic tag operations
  const addTagOptimistically = useCallback(
    (label: string) => {
      const newTag = {
        id: Date.now(), // Temporary ID
        label,
      };

      dispatch(
        radarrApi.util.updateQueryData('getTags', undefined, (draft) => {
          draft.push(newTag);
        })
      );

      return newTag;
    },
    [dispatch]
  );

  const removeTagOptimistically = useCallback(
    (tagId: number) => {
      dispatch(
        radarrApi.util.updateQueryData('getTags', undefined, (draft) => {
          return draft.filter((tag: Tag) => tag.id !== tagId);
        })
      );
    },
    [dispatch]
  );

  // Rollback functions for when optimistic updates fail
  const rollbackOptimisticUpdate = useCallback(
    (queryKey: string) => {
      // Force refetch the data to get the real state
      dispatch(radarrApi.util.invalidateTags([queryKey]));
    },
    [dispatch]
  );

  return {
    updateMovieOptimistically,
    removeQueueItemOptimistically,
    toggleMovieMonitoringOptimistically,
    updateQueueItemStatusOptimistically,
    addActivityOptimistically,
    addTagOptimistically,
    removeTagOptimistically,
    rollbackOptimisticUpdate,
  };
};

// Enhanced hooks that include optimistic updates
export const useOptimisticMovieUpdate = () => {
  const { updateMovieOptimistically, rollbackOptimisticUpdate } = useOptimisticUpdates();
  const [updateMovie] = radarrApi.useUpdateMovieMutation();

  return useCallback(
    async (movieId: number, updates: Partial<Movie>) => {
      // Apply optimistic update
      updateMovieOptimistically(movieId, updates);

      try {
        // Perform actual update
        const result = await updateMovie({ id: movieId, ...updates }).unwrap();
        return result;
      } catch (error) {
        // Rollback on error
        rollbackOptimisticUpdate('Movie');
        throw error;
      }
    },
    [updateMovieOptimistically, rollbackOptimisticUpdate, updateMovie]
  );
};

export const useOptimisticQueueRemoval = () => {
  const { removeQueueItemOptimistically, rollbackOptimisticUpdate } = useOptimisticUpdates();
  const [removeQueueItem] = radarrApi.useRemoveQueueItemMutation();

  return useCallback(
    async (queueItemId: number, options: { removeFromClient?: boolean; blocklist?: boolean } = {}) => {
      // Apply optimistic update
      removeQueueItemOptimistically(queueItemId);

      try {
        // Perform actual removal
        await removeQueueItem({ id: queueItemId, ...options }).unwrap();
      } catch (error) {
        // Rollback on error
        rollbackOptimisticUpdate('Queue');
        throw error;
      }
    },
    [removeQueueItemOptimistically, rollbackOptimisticUpdate, removeQueueItem]
  );
};

export const useOptimisticMovieMonitoring = () => {
  const { toggleMovieMonitoringOptimistically, rollbackOptimisticUpdate } = useOptimisticUpdates();
  const [updateMovie] = radarrApi.useUpdateMovieMutation();

  return useCallback(
    async (movieId: number, monitored: boolean) => {
      // Apply optimistic update
      toggleMovieMonitoringOptimistically(movieId, monitored);

      try {
        // Perform actual update
        const result = await updateMovie({ id: movieId, monitored }).unwrap();
        return result;
      } catch (error) {
        // Rollback on error
        rollbackOptimisticUpdate('Movie');
        rollbackOptimisticUpdate('WantedMovie');
        throw error;
      }
    },
    [toggleMovieMonitoringOptimistically, rollbackOptimisticUpdate, updateMovie]
  );
};

export const useOptimisticTagManagement = () => {
  const dispatch = useDispatch();
  const { addTagOptimistically, removeTagOptimistically, rollbackOptimisticUpdate } = useOptimisticUpdates();
  const [createTag] = radarrApi.useCreateTagMutation();
  const [deleteTag] = radarrApi.useDeleteTagMutation();

  const createTagOptimistically = useCallback(
    async (label: string) => {
      // Apply optimistic update
      const tempTag = addTagOptimistically(label);

      try {
        // Perform actual creation
        const result = await createTag({ label }).unwrap();

        // Update with real ID
        dispatch(
          radarrApi.util.updateQueryData('getTags', undefined, (draft) => {
            const tempIndex = draft.findIndex((tag: Tag) => tag.id === tempTag.id);
            if (tempIndex !== -1) {
              draft[tempIndex] = result;
            }
          })
        );

        return result;
      } catch (error) {
        // Rollback on error
        rollbackOptimisticUpdate('Tag');
        throw error;
      }
    },
    [dispatch, addTagOptimistically, rollbackOptimisticUpdate, createTag]
  );

  const deleteTagOptimistically = useCallback(
    async (tagId: number) => {
      // Apply optimistic update
      removeTagOptimistically(tagId);

      try {
        // Perform actual deletion
        await deleteTag(tagId).unwrap();
      } catch (error) {
        // Rollback on error
        rollbackOptimisticUpdate('Tag');
        throw error;
      }
    },
    [removeTagOptimistically, rollbackOptimisticUpdate, deleteTag]
  );

  return {
    createTagOptimistically,
    deleteTagOptimistically,
  };
};

export default useOptimisticUpdates;
