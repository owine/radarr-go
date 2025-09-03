import { createSlice } from '@reduxjs/toolkit';
import type { PayloadAction } from '@reduxjs/toolkit';

export interface UiState {
  theme: 'light' | 'dark' | 'auto';
  sidebarCollapsed: boolean;
  currentPage: string;
  breadcrumbs: BreadcrumbItem[];
  notifications: Notification[];
  isLoading: boolean;
  modals: {
    isAddMovieOpen: boolean;
    isEditMovieOpen: boolean;
    isDeleteConfirmOpen: boolean;
    isSettingsOpen: boolean;
  };
  filters: {
    movies: MovieFilters;
  };
  viewSettings: {
    moviesView: 'grid' | 'list' | 'table';
    moviesSort: string;
    moviesSortDirection: 'asc' | 'desc';
    itemsPerPage: number;
  };
}

export interface BreadcrumbItem {
  label: string;
  path?: string;
}

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  timestamp: number;
  autoClose?: boolean;
  duration?: number;
}

export interface MovieFilters {
  monitored?: boolean;
  status?: string[];
  genres?: string[];
  year?: { min?: number; max?: number };
  rating?: { min?: number; max?: number };
  qualityProfile?: number[];
  tags?: number[];
  search?: string;
}

const initialState: UiState = {
  theme: (localStorage.getItem('radarr_theme') as 'light' | 'dark' | 'auto') || 'auto',
  sidebarCollapsed: localStorage.getItem('radarr_sidebar_collapsed') === 'true',
  currentPage: '',
  breadcrumbs: [],
  notifications: [],
  isLoading: false,
  modals: {
    isAddMovieOpen: false,
    isEditMovieOpen: false,
    isDeleteConfirmOpen: false,
    isSettingsOpen: false,
  },
  filters: {
    movies: {},
  },
  viewSettings: {
    moviesView: (localStorage.getItem('radarr_movies_view') as 'grid' | 'list' | 'table') || 'grid',
    moviesSort: localStorage.getItem('radarr_movies_sort') || 'title',
    moviesSortDirection: (localStorage.getItem('radarr_movies_sort_direction') as 'asc' | 'desc') || 'asc',
    itemsPerPage: parseInt(localStorage.getItem('radarr_items_per_page') || '20'),
  },
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'auto'>) => {
      state.theme = action.payload;
      localStorage.setItem('radarr_theme', action.payload);
    },
    toggleSidebar: (state) => {
      state.sidebarCollapsed = !state.sidebarCollapsed;
      localStorage.setItem('radarr_sidebar_collapsed', state.sidebarCollapsed.toString());
    },
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload;
      localStorage.setItem('radarr_sidebar_collapsed', action.payload.toString());
    },
    setCurrentPage: (state, action: PayloadAction<string>) => {
      state.currentPage = action.payload;
    },
    setBreadcrumbs: (state, action: PayloadAction<BreadcrumbItem[]>) => {
      state.breadcrumbs = action.payload;
    },
    addNotification: (state, action: PayloadAction<Omit<Notification, 'id' | 'timestamp'>>) => {
      const notification: Notification = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: Date.now(),
        autoClose: action.payload.autoClose ?? (action.payload.type === 'success' || action.payload.type === 'info'),
        duration: action.payload.duration ?? 5000,
      };
      state.notifications.push(notification);
    },
    removeNotification: (state, action: PayloadAction<string>) => {
      state.notifications = state.notifications.filter(n => n.id !== action.payload);
    },
    clearNotifications: (state) => {
      state.notifications = [];
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    openModal: (state, action: PayloadAction<keyof UiState['modals']>) => {
      state.modals[action.payload] = true;
    },
    closeModal: (state, action: PayloadAction<keyof UiState['modals']>) => {
      state.modals[action.payload] = false;
    },
    closeAllModals: (state) => {
      Object.keys(state.modals).forEach(key => {
        state.modals[key as keyof UiState['modals']] = false;
      });
    },
    setMovieFilters: (state, action: PayloadAction<Partial<MovieFilters>>) => {
      state.filters.movies = { ...state.filters.movies, ...action.payload };
    },
    clearMovieFilters: (state) => {
      state.filters.movies = {};
    },
    setMoviesView: (state, action: PayloadAction<'grid' | 'list' | 'table'>) => {
      state.viewSettings.moviesView = action.payload;
      localStorage.setItem('radarr_movies_view', action.payload);
    },
    setMoviesSort: (state, action: PayloadAction<{ sort: string; direction: 'asc' | 'desc' }>) => {
      state.viewSettings.moviesSort = action.payload.sort;
      state.viewSettings.moviesSortDirection = action.payload.direction;
      localStorage.setItem('radarr_movies_sort', action.payload.sort);
      localStorage.setItem('radarr_movies_sort_direction', action.payload.direction);
    },
    setItemsPerPage: (state, action: PayloadAction<number>) => {
      state.viewSettings.itemsPerPage = action.payload;
      localStorage.setItem('radarr_items_per_page', action.payload.toString());
    },
  },
});

export const {
  setTheme,
  toggleSidebar,
  setSidebarCollapsed,
  setCurrentPage,
  setBreadcrumbs,
  addNotification,
  removeNotification,
  clearNotifications,
  setLoading,
  openModal,
  closeModal,
  closeAllModals,
  setMovieFilters,
  clearMovieFilters,
  setMoviesView,
  setMoviesSort,
  setItemsPerPage,
} = uiSlice.actions;

export default uiSlice.reducer;
