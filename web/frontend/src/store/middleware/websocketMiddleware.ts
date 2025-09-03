import { Middleware, isAnyOf } from '@reduxjs/toolkit';
import { radarrApi } from '../api/radarrApi';
import type { RootState } from '../index';

// WebSocket Event Types
export interface WebSocketEvent {
  type: string;
  data: any;
  timestamp: string;
}

export interface QueueUpdateEvent extends WebSocketEvent {
  type: 'QueueUpdate';
  data: {
    id: number;
    status: string;
    progress?: number;
    size?: number;
    sizeleft?: number;
    timeleft?: string;
    errorMessage?: string;
  };
}

export interface ActivityUpdateEvent extends WebSocketEvent {
  type: 'ActivityUpdate';
  data: {
    id: number;
    status: string;
    progress: number;
    message?: string;
    endTime?: string;
  };
}

export interface HealthUpdateEvent extends WebSocketEvent {
  type: 'HealthUpdate';
  data: {
    source: string;
    type: string;
    message: string;
    status: 'healthy' | 'warning' | 'error';
  };
}

export interface MovieUpdateEvent extends WebSocketEvent {
  type: 'MovieUpdate';
  data: {
    id: number;
    hasFile: boolean;
    monitored: boolean;
    lastInfoSync?: string;
  };
}

// WebSocket Connection States
export type WebSocketConnectionState =
  | 'disconnected'
  | 'connecting'
  | 'connected'
  | 'reconnecting'
  | 'error';

// WebSocket Configuration
interface WebSocketConfig {
  url: string;
  protocols?: string[];
  maxReconnectAttempts: number;
  reconnectDelay: number;
  maxReconnectDelay: number;
  reconnectBackoffMultiplier: number;
  heartbeatInterval: number;
  connectionTimeout: number;
}

// WebSocket State
interface WebSocketState {
  connection: WebSocket | null;
  connectionState: WebSocketConnectionState;
  reconnectAttempts: number;
  lastConnectedAt?: Date;
  lastDisconnectedAt?: Date;
  eventHistory: WebSocketEvent[];
  heartbeatTimer?: NodeJS.Timeout;
  reconnectTimer?: NodeJS.Timeout;
  connectionTimer?: NodeJS.Timeout;
}

// WebSocket Manager Class
class WebSocketManager {
  private state: WebSocketState = {
    connection: null,
    connectionState: 'disconnected',
    reconnectAttempts: 0,
    eventHistory: [],
  };

  private config: WebSocketConfig = {
    url: '',
    maxReconnectAttempts: 10,
    reconnectDelay: 1000,
    maxReconnectDelay: 30000,
    reconnectBackoffMultiplier: 2,
    heartbeatInterval: 30000,
    connectionTimeout: 10000,
  };

  private dispatch: any = null;
  private getState: (() => RootState) | null = null;
  private listeners: Map<string, ((event: WebSocketEvent) => void)[]> = new Map();

  initialize(dispatch: any, getState: () => RootState, apiKey?: string) {
    this.dispatch = dispatch;
    this.getState = getState;

    // Build WebSocket URL based on current location
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/v3/ws`;

    this.config.url = wsUrl;

    // Add API key to protocols if available
    if (apiKey) {
      this.config.protocols = [`radarr-api-key-${apiKey}`];
    }

    // Start connection
    this.connect();
  }

  connect() {
    if (this.state.connectionState === 'connecting' || this.state.connectionState === 'connected') {
      return;
    }

    this.setState('connecting');

    try {
      const ws = new WebSocket(this.config.url, this.config.protocols);

      // Set connection timeout
      this.state.connectionTimer = setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING) {
          ws.close();
          this.handleConnectionTimeout();
        }
      }, this.config.connectionTimeout);

      ws.onopen = this.handleOpen.bind(this);
      ws.onmessage = this.handleMessage.bind(this);
      ws.onclose = this.handleClose.bind(this);
      ws.onerror = this.handleError.bind(this);

      this.state.connection = ws;
    } catch (error) {
      console.error('WebSocket connection failed:', error);
      this.handleConnectionError(error as Error);
    }
  }

  disconnect() {
    this.clearTimers();

    if (this.state.connection) {
      this.state.connection.close(1000, 'Intentional disconnect');
    }

    this.setState('disconnected');
    this.state.connection = null;
    this.state.reconnectAttempts = 0;
  }

  send(data: any) {
    if (this.state.connectionState === 'connected' && this.state.connection) {
      try {
        this.state.connection.send(JSON.stringify(data));
        return true;
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        return false;
      }
    }
    return false;
  }

  subscribe(eventType: string, callback: (event: WebSocketEvent) => void) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, []);
    }
    this.listeners.get(eventType)!.push(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.listeners.get(eventType);
      if (callbacks) {
        const index = callbacks.indexOf(callback);
        if (index > -1) {
          callbacks.splice(index, 1);
        }
      }
    };
  }

  getConnectionState(): WebSocketConnectionState {
    return this.state.connectionState;
  }

  getEventHistory(): WebSocketEvent[] {
    return [...this.state.eventHistory];
  }

  private handleOpen() {
    this.clearTimers();
    this.setState('connected');
    this.state.lastConnectedAt = new Date();
    this.state.reconnectAttempts = 0;

    console.log('WebSocket connected successfully');

    // Start heartbeat
    this.startHeartbeat();

    // Dispatch connection event
    if (this.dispatch) {
      this.dispatch(radarrApi.util.invalidateTags(['Queue', 'Activity', 'Health']));
    }
  }

  private handleMessage(event: MessageEvent) {
    try {
      const wsEvent: WebSocketEvent = JSON.parse(event.data);
      this.addToEventHistory(wsEvent);
      this.processEvent(wsEvent);
      this.notifyListeners(wsEvent);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error, event.data);
    }
  }

  private handleClose(event: CloseEvent) {
    this.clearTimers();
    this.state.connection = null;
    this.state.lastDisconnectedAt = new Date();

    console.log('WebSocket closed:', event.code, event.reason);

    // Only attempt reconnection if not intentional disconnect
    if (event.code !== 1000 && this.state.reconnectAttempts < this.config.maxReconnectAttempts) {
      this.scheduleReconnect();
    } else {
      this.setState('disconnected');
    }
  }

  private handleError(error: Event) {
    console.error('WebSocket error:', error);
    this.handleConnectionError(new Error('WebSocket connection error'));
  }

  private handleConnectionTimeout() {
    console.error('WebSocket connection timeout');
    this.handleConnectionError(new Error('Connection timeout'));
  }

  private handleConnectionError(error: Error) {
    this.clearTimers();
    this.setState('error');

    if (this.state.reconnectAttempts < this.config.maxReconnectAttempts) {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect() {
    this.setState('reconnecting');
    this.state.reconnectAttempts++;

    const delay = Math.min(
      this.config.reconnectDelay * Math.pow(this.config.reconnectBackoffMultiplier, this.state.reconnectAttempts - 1),
      this.config.maxReconnectDelay
    );

    console.log(`Scheduling WebSocket reconnect in ${delay}ms (attempt ${this.state.reconnectAttempts})`);

    this.state.reconnectTimer = setTimeout(() => {
      this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.state.heartbeatTimer = setInterval(() => {
      if (this.state.connectionState === 'connected') {
        this.send({ type: 'ping', timestamp: new Date().toISOString() });
      }
    }, this.config.heartbeatInterval);
  }

  private clearTimers() {
    if (this.state.heartbeatTimer) {
      clearInterval(this.state.heartbeatTimer);
      this.state.heartbeatTimer = undefined;
    }

    if (this.state.reconnectTimer) {
      clearTimeout(this.state.reconnectTimer);
      this.state.reconnectTimer = undefined;
    }

    if (this.state.connectionTimer) {
      clearTimeout(this.state.connectionTimer);
      this.state.connectionTimer = undefined;
    }
  }

  private setState(newState: WebSocketConnectionState) {
    const oldState = this.state.connectionState;
    this.state.connectionState = newState;

    // Notify state change listeners
    this.notifyListeners({
      type: 'ConnectionStateChange',
      data: { oldState, newState },
      timestamp: new Date().toISOString(),
    });
  }

  private addToEventHistory(event: WebSocketEvent) {
    this.state.eventHistory.push(event);

    // Keep only last 1000 events
    if (this.state.eventHistory.length > 1000) {
      this.state.eventHistory = this.state.eventHistory.slice(-1000);
    }
  }

  private processEvent(event: WebSocketEvent) {
    if (!this.dispatch) return;

    // Handle different event types and update RTK Query cache
    switch (event.type) {
      case 'QueueUpdate':
        // Invalidate queue cache to trigger refetch
        this.dispatch(radarrApi.util.invalidateTags(['Queue']));
        break;

      case 'ActivityUpdate':
        // Invalidate activity cache
        this.dispatch(radarrApi.util.invalidateTags(['Activity']));
        break;

      case 'HealthUpdate':
        // Invalidate health cache
        this.dispatch(radarrApi.util.invalidateTags(['Health']));
        break;

      case 'MovieUpdate':
        // Invalidate specific movie and movies list
        const movieId = (event as MovieUpdateEvent).data.id;
        this.dispatch(radarrApi.util.invalidateTags([
          { type: 'Movie', id: movieId },
          { type: 'Movie', id: 'LIST' }
        ]));
        break;

      case 'DownloadComplete':
        // Invalidate multiple caches for download completion
        this.dispatch(radarrApi.util.invalidateTags([
          'Queue',
          'Movie',
          'Activity',
          'History',
          'WantedMovie'
        ]));
        break;

      case 'ImportComplete':
        // Similar to download complete
        this.dispatch(radarrApi.util.invalidateTags([
          'Movie',
          'Activity',
          'History'
        ]));
        break;
    }
  }

  private notifyListeners(event: WebSocketEvent) {
    // Notify specific event type listeners
    const callbacks = this.listeners.get(event.type);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(event);
        } catch (error) {
          console.error('Error in WebSocket event listener:', error);
        }
      });
    }

    // Notify all event listeners
    const allCallbacks = this.listeners.get('*');
    if (allCallbacks) {
      allCallbacks.forEach(callback => {
        try {
          callback(event);
        } catch (error) {
          console.error('Error in WebSocket all-event listener:', error);
        }
      });
    }
  }
}

// Global WebSocket manager instance
export const webSocketManager = new WebSocketManager();

// Redux middleware for WebSocket integration
export const websocketMiddleware: Middleware<{}, RootState> =
  (store) => (next) => (action) => {
    const { dispatch, getState } = store;

    // Initialize WebSocket connection when auth state changes
    if (action.type === 'auth/setAuthenticated' && action.payload === true) {
      const state = getState();
      const apiKey = state.auth.apiKey;

      // Initialize WebSocket with API key
      webSocketManager.initialize(dispatch, getState, apiKey || undefined);
    }

    // Disconnect WebSocket when logging out
    if (action.type === 'auth/logout' || action.type === 'auth/setAuthenticated' && action.payload === false) {
      webSocketManager.disconnect();
    }

    // Handle specific actions that should trigger WebSocket operations
    if (isAnyOf(
      radarrApi.endpoints.queueCommand.matchFulfilled,
      radarrApi.endpoints.grabRelease.matchFulfilled,
      radarrApi.endpoints.removeQueueItem.matchFulfilled
    )(action)) {
      // Force refresh relevant data after important actions
      setTimeout(() => {
        dispatch(radarrApi.util.invalidateTags(['Queue', 'Activity']));
      }, 1000);
    }

    return next(action);
  };

// Utility hook for WebSocket connection state
export const useWebSocketConnection = () => {
  return {
    connectionState: webSocketManager.getConnectionState(),
    subscribe: webSocketManager.subscribe.bind(webSocketManager),
    send: webSocketManager.send.bind(webSocketManager),
    disconnect: webSocketManager.disconnect.bind(webSocketManager),
    getEventHistory: webSocketManager.getEventHistory.bind(webSocketManager),
  };
};
