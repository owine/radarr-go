import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAppSelector } from '../../hooks/redux';
import { useGetQueueQuery, useGetHealthQuery } from '../../store/api/radarrApi';
import styles from './Sidebar.module.css';

interface NavigationItem {
  path: string;
  label: string;
  icon: React.ReactNode;
  badge?: 'queue' | 'health' | 'warning';
  subItems?: NavigationItem[];
}

const getNavigationItems = (): NavigationItem[] => [
  {
    path: '/dashboard',
    label: 'Dashboard',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="3" y="3" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
        <rect x="14" y="3" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
        <rect x="14" y="14" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
        <rect x="3" y="14" width="7" height="7" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
  },
  {
    path: '/movies',
    label: 'Movies',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
        <line x1="8" y1="21" x2="16" y2="21" stroke="currentColor" strokeWidth="2"/>
        <line x1="12" y1="17" x2="12" y2="21" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
  },
  {
    path: '/calendar',
    label: 'Calendar',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="3" y="4" width="18" height="18" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
        <line x1="16" y1="2" x2="16" y2="6" stroke="currentColor" strokeWidth="2"/>
        <line x1="8" y1="2" x2="8" y2="6" stroke="currentColor" strokeWidth="2"/>
        <line x1="3" y1="10" x2="21" y2="10" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
  },
  {
    path: '/activity',
    label: 'Activity',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <polyline points="22,12 18,12 15,21 9,3 6,12 2,12" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    badge: 'queue',
    subItems: [
      {
        path: '/activity/queue',
        label: 'Queue',
        icon: (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <line x1="8" y1="6" x2="21" y2="6" stroke="currentColor" strokeWidth="2"/>
            <line x1="8" y1="12" x2="21" y2="12" stroke="currentColor" strokeWidth="2"/>
            <line x1="8" y1="18" x2="21" y2="18" stroke="currentColor" strokeWidth="2"/>
            <line x1="3" y1="6" x2="3.01" y2="6" stroke="currentColor" strokeWidth="2"/>
            <line x1="3" y1="12" x2="3.01" y2="12" stroke="currentColor" strokeWidth="2"/>
            <line x1="3" y1="18" x2="3.01" y2="18" stroke="currentColor" strokeWidth="2"/>
          </svg>
        ),
        badge: 'queue',
      },
      {
        path: '/activity/history',
        label: 'History',
        icon: (
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
            <polyline points="12,6 12,12 16,14" stroke="currentColor" strokeWidth="2"/>
          </svg>
        ),
      },
    ],
  },
  {
    path: '/wanted',
    label: 'Wanted',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
        <line x1="12" y1="8" x2="12" y2="12" stroke="currentColor" strokeWidth="2"/>
        <line x1="12" y1="16" x2="12.01" y2="16" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
  },
  {
    path: '/settings',
    label: 'Settings',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
  },
  {
    path: '/system',
    label: 'System',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
        <line x1="8" y1="21" x2="16" y2="21" stroke="currentColor" strokeWidth="2"/>
        <line x1="12" y1="17" x2="12" y2="21" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    badge: 'health',
  },
];

export const Sidebar = () => {
  const location = useLocation();
  const sidebarCollapsed = useAppSelector(state => state.ui.sidebarCollapsed);

  // Get data for activity indicators
  const { data: queueData } = useGetQueueQuery({}, {
    pollingInterval: 10000, // Poll every 10 seconds
  });
  const { data: healthData } = useGetHealthQuery(undefined, {
    pollingInterval: 30000, // Poll every 30 seconds
  });

  // Calculate badge counts
  const getBadgeCount = (badge: string | undefined) => {
    switch (badge) {
      case 'queue':
        return queueData?.length || 0;
      case 'health':
        return healthData?.filter(check => check.type === 'error' || check.type === 'warning').length || 0;
      default:
        return 0;
    }
  };

  const getBadgeColor = (badge: string | undefined, count: number) => {
    if (count === 0) return 'default';

    switch (badge) {
      case 'queue':
        return 'info';
      case 'health':
        return healthData?.some(check => check.type === 'error') ? 'error' : 'warning';
      default:
        return 'default';
    }
  };

  const isPathActive = (itemPath: string) => {
    if (itemPath === '/dashboard') {
      return location.pathname === '/' || location.pathname === '/dashboard';
    }
    return location.pathname.startsWith(itemPath);
  };

  const navigationItems = getNavigationItems();

  return (
    <aside className={`${styles.sidebar} ${sidebarCollapsed ? styles.collapsed : ''}`}>
      <nav className={styles.navigation}>
        {navigationItems.map((item) => {
          const badgeCount = getBadgeCount(item.badge);
          const badgeColor = getBadgeColor(item.badge, badgeCount);
          const isActive = isPathActive(item.path);

          return (
            <div key={item.path} className={styles.navGroup}>
              <NavLink
                to={item.path}
                className={`${styles.navItem} ${isActive ? styles.active : ''}`}
              >
                <span className={styles.iconWrapper}>
                  <span className={styles.icon}>{item.icon}</span>
                  {badgeCount > 0 && (
                    <span className={`${styles.badge} ${styles[badgeColor]}`}>
                      {badgeCount > 99 ? '99+' : badgeCount}
                    </span>
                  )}
                </span>
                {!sidebarCollapsed && (
                  <span className={styles.label}>{item.label}</span>
                )}
                {!sidebarCollapsed && item.subItems && (
                  <svg
                    className={`${styles.expandIcon} ${isActive ? styles.expanded : ''}`}
                    width="16" height="16" viewBox="0 0 24 24" fill="none"
                  >
                    <polyline points="9,18 15,12 9,6" stroke="currentColor" strokeWidth="2"/>
                  </svg>
                )}
              </NavLink>

              {!sidebarCollapsed && item.subItems && isActive && (
                <div className={styles.subNavigation}>
                  {item.subItems.map((subItem) => {
                    const subBadgeCount = getBadgeCount(subItem.badge);
                    const subBadgeColor = getBadgeColor(subItem.badge, subBadgeCount);

                    return (
                      <NavLink
                        key={subItem.path}
                        to={subItem.path}
                        className={({ isActive }) =>
                          `${styles.subNavItem} ${isActive ? styles.active : ''}`
                        }
                      >
                        <span className={styles.iconWrapper}>
                          <span className={styles.icon}>{subItem.icon}</span>
                          {subBadgeCount > 0 && (
                            <span className={`${styles.badge} ${styles[subBadgeColor]}`}>
                              {subBadgeCount > 99 ? '99+' : subBadgeCount}
                            </span>
                          )}
                        </span>
                        <span className={styles.label}>{subItem.label}</span>
                      </NavLink>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </nav>

      {/* Footer with system version */}
      {!sidebarCollapsed && (
        <div className={styles.footer}>
          <div className={styles.version}>
            <div className={styles.versionText}>Radarr Go</div>
            <div className={styles.versionNumber}>v0.9.0-alpha</div>
          </div>
        </div>
      )}
    </aside>
  );
};
