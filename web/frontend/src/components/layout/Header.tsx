import React, { useState, useRef, useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../hooks/redux';
import { toggleSidebar, setTheme } from '../../store/slices/uiSlice';
import { logout } from '../../store/slices/authSlice';
import { useGetSystemStatusQuery, useGetHealthQuery } from '../../store/api/radarrApi';
import { Button, Input } from '../common';
import styles from './Header.module.css';

export const Header = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useAppSelector(state => state.ui.theme);
  const { user } = useAppSelector(state => state.auth);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchFocused, setSearchFocused] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const userMenuRef = useRef<HTMLDivElement>(null);

  // Get system status and health for indicators
  const { data: systemStatus } = useGetSystemStatusQuery(undefined, {
    pollingInterval: 30000, // Poll every 30 seconds
  });
  const { data: healthChecks } = useGetHealthQuery(undefined, {
    pollingInterval: 60000, // Poll every minute
  });

  // Generate breadcrumbs based on current route
  const getBreadcrumbs = () => {
    const path = location.pathname;
    const segments = path.split('/').filter(Boolean);

    const breadcrumbs = [{ label: 'Home', path: '/dashboard' }];

    let currentPath = '';
    segments.forEach(segment => {
      currentPath += `/${segment}`;
      const label = segment.charAt(0).toUpperCase() + segment.slice(1);
      breadcrumbs.push({ label, path: currentPath });
    });

    return breadcrumbs.slice(0, -1); // Remove last item as it's the current page
  };

  // Get health status summary
  const getHealthStatus = () => {
    if (!healthChecks) return { status: 'unknown', count: 0 };

    const errorCount = healthChecks.filter(check => check.type === 'error').length;
    const warningCount = healthChecks.filter(check => check.type === 'warning').length;

    if (errorCount > 0) return { status: 'error', count: errorCount };
    if (warningCount > 0) return { status: 'warning', count: warningCount };
    return { status: 'ok', count: 0 };
  };

  const handleToggleSidebar = () => {
    dispatch(toggleSidebar());
  };

  const handleToggleTheme = () => {
    const nextTheme = theme === 'light' ? 'dark' : theme === 'dark' ? 'auto' : 'light';
    dispatch(setTheme(nextTheme));
  };

  const handleLogout = () => {
    dispatch(logout());
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      navigate(`/movies?search=${encodeURIComponent(searchQuery.trim())}`);
      setSearchQuery('');
      searchInputRef.current?.blur();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === '/' && !searchFocused) {
      e.preventDefault();
      searchInputRef.current?.focus();
    }
    if (e.key === 'Escape') {
      searchInputRef.current?.blur();
      setUserMenuOpen(false);
    }
  };

  // Click outside handler for user menu
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setUserMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [searchFocused]);

  const breadcrumbs = getBreadcrumbs();
  const healthStatus = getHealthStatus();

  return (
    <header className={styles.header}>
      <div className={styles.left}>
        <Button
          variant="ghost"
          iconOnly
          onClick={handleToggleSidebar}
          aria-label="Toggle sidebar"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path
              d="M3 12H21M3 6H21M3 18H21"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </Button>

        <div className={styles.logo}>
          <h1>Radarr</h1>
        </div>
      </div>

      <div className={styles.center}>
        <div className={styles.breadcrumbs}>
          {breadcrumbs.map((crumb, index) => (
            <React.Fragment key={crumb.path}>
              <button
                className={styles.breadcrumb}
                onClick={() => navigate(crumb.path)}
              >
                {crumb.label}
              </button>
              {index < breadcrumbs.length - 1 && (
                <span className={styles.breadcrumbSeparator}>/</span>
              )}
            </React.Fragment>
          ))}
        </div>

        <div className={styles.search}>
          <form onSubmit={handleSearch} className={styles.searchForm}>
            <div className={styles.searchWrapper}>
              <svg className={styles.searchIcon} width="16" height="16" viewBox="0 0 24 24" fill="none">
                <circle cx="11" cy="11" r="8" stroke="currentColor" strokeWidth="2"/>
                <path d="m21 21-4.35-4.35" stroke="currentColor" strokeWidth="2"/>
              </svg>
              <input
                ref={searchInputRef}
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onFocus={() => setSearchFocused(true)}
                onBlur={() => setSearchFocused(false)}
                placeholder="Search movies... (Press / to focus)"
                className={styles.searchInput}
              />
              {!searchFocused && (
                <kbd className={styles.searchShortcut}>/</kbd>
              )}
            </div>
          </form>
        </div>
      </div>

      <div className={styles.right}>
        {/* Health status indicator */}
        <div className={styles.statusIndicator}>
          <div
            className={`${styles.healthBadge} ${styles[healthStatus.status]}`}
            title={`System health: ${healthStatus.status}${healthStatus.count > 0 ? ` (${healthStatus.count} issues)` : ''}`}
          >
            {healthStatus.status === 'error' && (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
                <line x1="15" y1="9" x2="9" y2="15" stroke="currentColor" strokeWidth="2"/>
                <line x1="9" y1="9" x2="15" y2="15" stroke="currentColor" strokeWidth="2"/>
              </svg>
            )}
            {healthStatus.status === 'warning' && (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" stroke="currentColor" strokeWidth="2"/>
                <line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="17" r="1" fill="currentColor"/>
              </svg>
            )}
            {healthStatus.status === 'ok' && (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" stroke="currentColor" strokeWidth="2"/>
                <polyline points="22,4 12,14.01 9,11.01" stroke="currentColor" strokeWidth="2"/>
              </svg>
            )}
            {healthStatus.count > 0 && (
              <span className={styles.healthCount}>{healthStatus.count}</span>
            )}
          </div>
        </div>

        <Button
          variant="ghost"
          iconOnly
          onClick={handleToggleTheme}
          aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
        >
          {theme === 'dark' ? (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <circle cx="12" cy="12" r="5" stroke="currentColor" strokeWidth="2"/>
              <path d="M12 1v2m0 18v2M4.22 4.22l1.42 1.42m12.72 12.72 1.42 1.42M1 12h2m18 0h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" stroke="currentColor" strokeWidth="2"/>
            </svg>
          ) : (
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke="currentColor" strokeWidth="2"/>
            </svg>
          )}
        </Button>

        {/* User menu */}
        <div className={styles.userMenu} ref={userMenuRef}>
          <button
            className={styles.userButton}
            onClick={() => setUserMenuOpen(!userMenuOpen)}
            aria-label="User menu"
          >
            <div className={styles.userAvatar}>
              {user?.username ? user.username.charAt(0).toUpperCase() : 'U'}
            </div>
            <span className={styles.userName}>
              {user?.username || 'User'}
            </span>
            <svg
              className={`${styles.userMenuIcon} ${userMenuOpen ? styles.rotated : ''}`}
              width="16" height="16" viewBox="0 0 24 24" fill="none"
            >
              <polyline points="6,9 12,15 18,9" stroke="currentColor" strokeWidth="2"/>
            </svg>
          </button>

          {userMenuOpen && (
            <div className={styles.userDropdown}>
              <div className={styles.userInfo}>
                <div className={styles.userAvatarLarge}>
                  {user?.username ? user.username.charAt(0).toUpperCase() : 'U'}
                </div>
                <div className={styles.userDetails}>
                  <div className={styles.userDisplayName}>
                    {user?.username || 'User'}
                  </div>
                  <div className={styles.userRole}>
                    {systemStatus?.authentication?.method || 'API Key'} User
                  </div>
                </div>
              </div>

              <div className={styles.userMenuDivider}></div>

              <button
                className={styles.userMenuItem}
                onClick={() => {
                  navigate('/system');
                  setUserMenuOpen(false);
                }}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" stroke="currentColor" strokeWidth="2"/>
                </svg>
                Settings
              </button>

              <div className={styles.userMenuDivider}></div>

              <button
                className={`${styles.userMenuItem} ${styles.danger}`}
                onClick={() => {
                  handleLogout();
                  setUserMenuOpen(false);
                }}
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" stroke="currentColor" strokeWidth="2"/>
                  <polyline points="16,17 21,12 16,7" stroke="currentColor" strokeWidth="2"/>
                  <line x1="21" y1="12" x2="9" y2="12" stroke="currentColor" strokeWidth="2"/>
                </svg>
                Logout
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
