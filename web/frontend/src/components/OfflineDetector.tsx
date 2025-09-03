import React, { useState, useEffect } from 'react';
import { useAppDispatch } from '../hooks/redux';
import { addNotification } from '../store/slices/uiSlice';
import styles from './OfflineDetector.module.css';

interface OfflineDetectorProps {
  children: React.ReactNode;
}

export const OfflineDetector: React.FC<OfflineDetectorProps> = ({ children }) => {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [showOfflineBanner, setShowOfflineBanner] = useState(false);
  const dispatch = useAppDispatch();

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      setShowOfflineBanner(false);

      dispatch(addNotification({
        type: 'success',
        title: 'Connection Restored',
        message: 'You are back online. The app will sync any pending changes.',
        duration: 3000,
      }));
    };

    const handleOffline = () => {
      setIsOnline(false);
      setShowOfflineBanner(true);

      dispatch(addNotification({
        type: 'warning',
        title: 'Connection Lost',
        message: 'You are currently offline. Some features may not work properly.',
        duration: 5000,
      }));
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Initial check
    if (!navigator.onLine) {
      setShowOfflineBanner(true);
    }

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [dispatch]);

  const dismissBanner = () => {
    setShowOfflineBanner(false);
  };

  const retryConnection = () => {
    // Force a network check by trying to fetch a small resource
    fetch('/api/v3/system/status', {
      method: 'HEAD',
      cache: 'no-cache',
    })
      .then(() => {
        if (!navigator.onLine) {
          // If fetch succeeds but navigator.onLine is false, manually trigger online event
          window.dispatchEvent(new Event('online'));
        }
      })
      .catch(() => {
        dispatch(addNotification({
          type: 'error',
          title: 'Still Offline',
          message: 'Unable to establish connection. Please check your network.',
          duration: 3000,
        }));
      });
  };

  return (
    <>
      {showOfflineBanner && (
        <div className={styles.offlineBanner}>
          <div className={styles.bannerContent}>
            <div className={styles.bannerIcon}>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <path
                  d="M1 9l2 2c4.97-4.97 13.03-4.97 18 0l2-2C16.93 2.93 7.07 2.93 1 9z"
                  stroke="currentColor"
                  strokeWidth="2"
                  fill="none"
                />
                <path
                  d="M8.5 16.5l2 2 2-2"
                  stroke="currentColor"
                  strokeWidth="2"
                  fill="none"
                />
                <line
                  x1="1"
                  y1="1"
                  x2="23"
                  y2="23"
                  stroke="currentColor"
                  strokeWidth="2"
                />
              </svg>
            </div>

            <div className={styles.bannerMessage}>
              <span className={styles.bannerTitle}>You're offline</span>
              <span className={styles.bannerDescription}>
                Check your connection and try again
              </span>
            </div>

            <div className={styles.bannerActions}>
              <button
                className={styles.retryButton}
                onClick={retryConnection}
                type="button"
              >
                Retry
              </button>

              <button
                className={styles.dismissButton}
                onClick={dismissBanner}
                type="button"
                aria-label="Dismiss offline banner"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                  <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" strokeWidth="2"/>
                  <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" strokeWidth="2"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add offline indicator to app when offline but banner is dismissed */}
      {!isOnline && !showOfflineBanner && (
        <div className={styles.offlineIndicator} onClick={() => setShowOfflineBanner(true)}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <path
              d="M1 9l2 2c4.97-4.97 13.03-4.97 18 0l2-2C16.93 2.93 7.07 2.93 1 9z"
              stroke="currentColor"
              strokeWidth="2"
              fill="none"
            />
            <line
              x1="1"
              y1="1"
              x2="23"
              y2="23"
              stroke="currentColor"
              strokeWidth="2"
            />
          </svg>
        </div>
      )}

      {children}
    </>
  );
};
