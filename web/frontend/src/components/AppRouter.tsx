import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAppSelector } from '../hooks/redux';
import { Layout } from './layout/Layout';
import { ProtectedRoute } from './ProtectedRoute';
import { ErrorBoundary } from './ErrorBoundary';
import { OfflineDetector } from './OfflineDetector';
import { LoginPage } from '../pages/LoginPage';
import { NotFoundPage } from '../pages/NotFoundPage';
import { SkeletonCard, SkeletonMovieGrid, SkeletonTable } from './common';
import styles from './AppRouter.module.css';

// Lazy load pages for code splitting
const DashboardPage = lazy(() => import('../pages/DashboardPage').then(module => ({ default: module.DashboardPage })));
const MoviesPage = lazy(() => import('../pages/MoviesPage').then(module => ({ default: module.MoviesPage })));
const SystemPage = lazy(() => import('../pages/SystemPage').then(module => ({ default: module.SystemPage })));

// Lazy load additional pages
const CalendarPage = lazy(() => import('../pages/CalendarPage').then(module => ({ default: module.CalendarPage || (() => <div>Calendar Page (Coming Soon)</div>) })));
const QueuePage = lazy(() => import('../pages/QueuePage').then(module => ({ default: module.QueuePage || (() => <div>Queue Page (Coming Soon)</div>) })));
const HistoryPage = lazy(() => import('../pages/HistoryPage').then(module => ({ default: module.HistoryPage || (() => <div>History Page (Coming Soon)</div>) })));
const WantedPage = lazy(() => import('../pages/WantedPage').then(module => ({ default: module.WantedPage || (() => <div>Wanted Page (Coming Soon)</div>) })));
const SettingsPage = lazy(() => import('../pages/SettingsPage').then(module => ({ default: module.SettingsPage })));

// Loading components for different page types
const DashboardLoading = () => (
  <div className={styles.loadingContainer}>
    <div className={styles.dashboardLoading}>
      <SkeletonCard />
      <SkeletonCard />
      <SkeletonCard />
    </div>
  </div>
);

const MoviesLoading = () => (
  <div className={styles.loadingContainer}>
    <SkeletonMovieGrid items={12} />
  </div>
);

const TableLoading = () => (
  <div className={styles.loadingContainer}>
    <SkeletonTable rows={10} columns={5} />
  </div>
);

const GenericLoading = () => (
  <div className={styles.loadingContainer}>
    <SkeletonCard />
  </div>
);

// Unauthorized page component
const UnauthorizedPage = () => (
  <div className={styles.unauthorizedContainer}>
    <div className={styles.unauthorizedContent}>
      <h1>Access Denied</h1>
      <p>You don't have permission to access this resource.</p>
      <button onClick={() => window.history.back()}>Go Back</button>
    </div>
  </div>
);

export const AppRouter: React.FC = () => {
  const isAuthenticated = useAppSelector(state => state.auth.isAuthenticated);

  if (!isAuthenticated) {
    return (
      <ErrorBoundary>
        <OfflineDetector>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="*" element={<Navigate to="/login" replace />} />
          </Routes>
        </OfflineDetector>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <OfflineDetector>
        <Layout>
          <Routes>
            {/* Dashboard */}
            <Route
              path="/"
              element={<Navigate to="/dashboard" replace />}
            />
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute fallback={<DashboardLoading />}>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />

            {/* Movies */}
            <Route
              path="/movies"
              element={
                <ProtectedRoute fallback={<MoviesLoading />}>
                  <MoviesPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/movies/:id"
              element={
                <ProtectedRoute fallback={<GenericLoading />}>
                  <div>Movie Detail Page (Coming Soon)</div>
                </ProtectedRoute>
              }
            />

            {/* Calendar */}
            <Route
              path="/calendar"
              element={
                <ProtectedRoute fallback={<TableLoading />}>
                  <CalendarPage />
                </ProtectedRoute>
              }
            />

            {/* Activity */}
            <Route path="/activity">
              <Route
                index
                element={<Navigate to="/activity/queue" replace />}
              />
              <Route
                path="queue"
                element={
                  <ProtectedRoute fallback={<TableLoading />}>
                    <QueuePage />
                  </ProtectedRoute>
                }
              />
              <Route
                path="history"
                element={
                  <ProtectedRoute fallback={<TableLoading />}>
                    <HistoryPage />
                  </ProtectedRoute>
                }
              />
            </Route>

            {/* Wanted */}
            <Route
              path="/wanted"
              element={
                <ProtectedRoute fallback={<TableLoading />}>
                  <WantedPage />
                </ProtectedRoute>
              }
            />

            {/* Settings */}
            <Route
              path="/settings"
              element={
                <ProtectedRoute fallback={<GenericLoading />}>
                  <SettingsPage />
                </ProtectedRoute>
              }
            />

            {/* System */}
            <Route
              path="/system"
              element={
                <ProtectedRoute fallback={<GenericLoading />}>
                  <SystemPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/system/status"
              element={
                <ProtectedRoute fallback={<GenericLoading />}>
                  <div>System Status Page (Coming Soon)</div>
                </ProtectedRoute>
              }
            />
            <Route
              path="/system/settings"
              element={
                <ProtectedRoute fallback={<GenericLoading />}>
                  <div>System Settings Page (Coming Soon)</div>
                </ProtectedRoute>
              }
            />

            {/* Special routes */}
            <Route path="/unauthorized" element={<UnauthorizedPage />} />
            <Route path="/login" element={<Navigate to="/dashboard" replace />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Layout>
      </OfflineDetector>
    </ErrorBoundary>
  );
};
