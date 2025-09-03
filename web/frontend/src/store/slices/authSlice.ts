import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';

export interface AuthState {
  isAuthenticated: boolean;
  apiKey: string | null;
  error: string | null;
  isLoading: boolean;
  rememberMe: boolean;
  lastValidated: number | null;
  user: {
    username?: string;
    permissions?: string[];
  } | null;
}

const getStoredApiKey = (): string | null => {
  const rememberMe = localStorage.getItem('radarr_remember_me') === 'true';
  if (rememberMe) {
    return localStorage.getItem('radarr_api_key') || null;
  }
  return sessionStorage.getItem('radarr_api_key') || null;
};

const initialState: AuthState = {
  isAuthenticated: false,
  apiKey: getStoredApiKey(),
  error: null,
  isLoading: false,
  rememberMe: localStorage.getItem('radarr_remember_me') === 'true',
  lastValidated: parseInt(localStorage.getItem('radarr_last_validated') || '0') || null,
  user: null,
};

// Set initial authentication state based on stored API key or development mode
if (initialState.apiKey) {
  initialState.isAuthenticated = true;
} else if (import.meta.env.DEV) {
  // In development mode, auto-authenticate if backend doesn't require auth
  initialState.isAuthenticated = true;
  initialState.apiKey = 'dev-mode-bypass';
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    loginStart: (state) => {
      state.isLoading = true;
      state.error = null;
    },
    loginSuccess: (state, action: PayloadAction<{ apiKey: string; user?: any; rememberMe?: boolean }>) => {
      state.isAuthenticated = true;
      state.apiKey = action.payload.apiKey;
      state.error = null;
      state.isLoading = false;
      state.user = action.payload.user || null;
      state.lastValidated = Date.now();

      if (action.payload.rememberMe !== undefined) {
        state.rememberMe = action.payload.rememberMe;
      }

      const storage = state.rememberMe ? localStorage : sessionStorage;
      storage.setItem('radarr_api_key', action.payload.apiKey);
      localStorage.setItem('radarr_remember_me', state.rememberMe.toString());
      localStorage.setItem('radarr_last_validated', state.lastValidated.toString());
    },
    loginFailure: (state, action: PayloadAction<string>) => {
      state.isAuthenticated = false;
      state.apiKey = null;
      state.error = action.payload;
      state.isLoading = false;
      state.user = null;
      state.lastValidated = null;
      localStorage.removeItem('radarr_api_key');
      sessionStorage.removeItem('radarr_api_key');
      localStorage.removeItem('radarr_last_validated');
    },
    logout: (state) => {
      state.isAuthenticated = false;
      state.apiKey = null;
      state.error = null;
      state.isLoading = false;
      state.user = null;
      state.lastValidated = null;
      localStorage.removeItem('radarr_api_key');
      sessionStorage.removeItem('radarr_api_key');
      localStorage.removeItem('radarr_last_validated');
      // Keep rememberMe preference for next login
    },
    clearError: (state) => {
      state.error = null;
    },
    setApiKey: (state, action: PayloadAction<string>) => {
      state.apiKey = action.payload;
      state.isAuthenticated = !!action.payload;
      if (action.payload) {
        const storage = state.rememberMe ? localStorage : sessionStorage;
        storage.setItem('radarr_api_key', action.payload);
      } else {
        localStorage.removeItem('radarr_api_key');
        sessionStorage.removeItem('radarr_api_key');
      }
    },
    setRememberMe: (state, action: PayloadAction<boolean>) => {
      state.rememberMe = action.payload;
      localStorage.setItem('radarr_remember_me', action.payload.toString());
    },
    updateUser: (state, action: PayloadAction<any>) => {
      state.user = action.payload;
    },
  },
});

export const {
  loginStart,
  loginSuccess,
  loginFailure,
  logout,
  clearError,
  setApiKey,
  setRememberMe,
  updateUser,
} = authSlice.actions;

export default authSlice.reducer;
