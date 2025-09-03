import React, { Suspense } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAppSelector } from '../hooks/redux';
import { SkeletonCard } from './common';
import { ErrorBoundary } from './ErrorBoundary';
import styles from './ProtectedRoute.module.css';

interface ProtectedRouteProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  requireAuth?: boolean;
  requirePermissions?: string[];
}

const DefaultLoadingFallback = () => (
  <div className={styles.loadingContainer}>
    <div className={styles.loadingGrid}>
      <SkeletonCard />
      <SkeletonCard />
      <SkeletonCard />
    </div>
  </div>
);

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  fallback = <DefaultLoadingFallback />,
  requireAuth = true,
  requirePermissions = [],
}) => {
  const location = useLocation();
  const { isAuthenticated, user } = useAppSelector(state => state.auth);

  // Check authentication
  if (requireAuth && !isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Check permissions (if user has permissions system)
  if (requirePermissions.length > 0 && user?.permissions) {
    const hasRequiredPermissions = requirePermissions.every(permission =>
      user.permissions.includes(permission)
    );

    if (!hasRequiredPermissions) {
      return <Navigate to="/unauthorized" replace />;
    }
  }

  return (
    <ErrorBoundary>
      <Suspense fallback={fallback}>
        {children}
      </Suspense>
    </ErrorBoundary>
  );
};
