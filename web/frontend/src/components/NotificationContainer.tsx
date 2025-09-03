import { useEffect } from 'react';
import { useAppSelector, useAppDispatch } from '../hooks/redux';
import { removeNotification } from '../store/slices/uiSlice';
import styles from './NotificationContainer.module.css';

export const NotificationContainer = () => {
  const notifications = useAppSelector(state => state.ui.notifications);
  const dispatch = useAppDispatch();

  useEffect(() => {
    notifications.forEach(notification => {
      if (notification.autoClose) {
        const timer = setTimeout(() => {
          dispatch(removeNotification(notification.id));
        }, notification.duration || 5000);

        return () => clearTimeout(timer);
      }
    });
  }, [notifications, dispatch]);

  const handleClose = (id: string) => {
    dispatch(removeNotification(id));
  };

  if (notifications.length === 0) return null;

  return (
    <div className={styles.container}>
      {notifications.map(notification => (
        <div
          key={notification.id}
          className={`${styles.notification} ${styles[notification.type]}`}
        >
          <div className={styles.content}>
            <div className={styles.title}>{notification.title}</div>
            {notification.message && (
              <div className={styles.message}>{notification.message}</div>
            )}
          </div>
          <button
            className={styles.closeButton}
            onClick={() => handleClose(notification.id)}
            aria-label="Close notification"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
};
