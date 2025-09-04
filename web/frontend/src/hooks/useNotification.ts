import { useCallback } from 'react';
import { useAppDispatch } from './redux';
import { addNotification } from '../store/slices/uiSlice';
import type { Notification } from '../store/slices/uiSlice';

export type NotificationInput = Omit<Notification, 'id' | 'timestamp'>;

export const useNotification = () => {
  const dispatch = useAppDispatch();

  const showNotification = useCallback(
    (notification: NotificationInput) => {
      dispatch(addNotification(notification));
    },
    [dispatch]
  );

  // Convenience methods for different notification types
  const showSuccess = useCallback(
    (title: string, message?: string, options?: Partial<NotificationInput>) => {
      showNotification({
        type: 'success',
        title,
        message,
        ...options,
      });
    },
    [showNotification]
  );

  const showError = useCallback(
    (title: string, message?: string, options?: Partial<NotificationInput>) => {
      showNotification({
        type: 'error',
        title,
        message,
        autoClose: false, // Errors shouldn't auto-close by default
        ...options,
      });
    },
    [showNotification]
  );

  const showWarning = useCallback(
    (title: string, message?: string, options?: Partial<NotificationInput>) => {
      showNotification({
        type: 'warning',
        title,
        message,
        autoClose: false, // Warnings shouldn't auto-close by default
        ...options,
      });
    },
    [showNotification]
  );

  const showInfo = useCallback(
    (title: string, message?: string, options?: Partial<NotificationInput>) => {
      showNotification({
        type: 'info',
        title,
        message,
        ...options,
      });
    },
    [showNotification]
  );

  return {
    showNotification,
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };
};
