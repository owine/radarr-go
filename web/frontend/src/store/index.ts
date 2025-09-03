import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { radarrApi } from './api/radarrApi';
import { websocketMiddleware } from './middleware/websocketMiddleware';
import authReducer from './slices/authSlice';
import uiReducer from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    ui: uiReducer,
    [radarrApi.reducerPath]: radarrApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: [
          radarrApi.util.resetApiState.type,
          // Add WebSocket action types that may contain non-serializable data
          'auth/setAuthenticated',
          'auth/logout',
        ],
        ignoredActionsPaths: ['payload.timestamp', 'payload.date'],
        ignoredPaths: ['websocket.connection', 'websocket.heartbeatTimer'],
      },
    })
    .concat(radarrApi.middleware)
    .concat(websocketMiddleware),
  devTools: import.meta.env.DEV,
});

// Enable refetch on focus/reconnect behaviors
setupListeners(store.dispatch);

// Initialize cache invalidation manager
import { initializeCacheInvalidation } from '../utils/cacheInvalidation';
initializeCacheInvalidation(store.dispatch);

// Initialize persistence layer
import { initializePersistenceLayer } from '../utils/persistenceLayer';
initializePersistenceLayer(store.dispatch, () => store.getState());

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;