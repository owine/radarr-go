import { useGetSystemStatusQuery, useGetHealthQuery, useGetMoviesQuery } from '../store/api/radarrApi';
import { Card } from '../components/common';
import styles from './DashboardPage.module.css';

export const DashboardPage = () => {
  const { data: systemStatus, isLoading: statusLoading } = useGetSystemStatusQuery();
  const { data: healthChecks, isLoading: healthLoading } = useGetHealthQuery();
  const { data: movies, isLoading: moviesLoading } = useGetMoviesQuery({ pageSize: 5 });

  const errorChecks = healthChecks?.filter(check => check.type === 'error') || [];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1>Dashboard</h1>
        <p>Welcome back! Here's what's happening with your movies.</p>
      </div>

      <div className={styles.statsGrid}>
        <Card title="System Status" loading={statusLoading}>
          {systemStatus && (
            <div className={styles.statusInfo}>
              <div className={styles.statusItem}>
                <span className={styles.label}>Version:</span>
                <span className={styles.value}>{systemStatus.version}</span>
              </div>
              <div className={styles.statusItem}>
                <span className={styles.label}>Uptime:</span>
                <span className={styles.value}>
                  {systemStatus.startTime ?
                    new Date(systemStatus.startTime).toLocaleDateString() :
                    'Unknown'
                  }
                </span>
              </div>
              <div className={styles.statusItem}>
                <span className={styles.label}>Database:</span>
                <span className={styles.value}>
                  {systemStatus.sqliteVersion ? 'SQLite' : 'External'}
                </span>
              </div>
            </div>
          )}
        </Card>

        <Card title="Health" loading={healthLoading}>
          <div className={styles.healthStatus}>
            {errorChecks.length > 0 ? (
              <div className={`${styles.healthIndicator} ${styles.error}`}>
                <span className={styles.healthIcon}>⚠️</span>
                <div>
                  <div className={styles.healthTitle}>
                    {errorChecks.length} Issue{errorChecks.length !== 1 ? 's' : ''}
                  </div>
                  <div className={styles.healthSubtitle}>Requires attention</div>
                </div>
              </div>
            ) : (
              <div className={`${styles.healthIndicator} ${styles.healthy}`}>
                <span className={styles.healthIcon}>✅</span>
                <div>
                  <div className={styles.healthTitle}>All Systems Healthy</div>
                  <div className={styles.healthSubtitle}>No issues detected</div>
                </div>
              </div>
            )}
          </div>
        </Card>

        <Card title="Movies" loading={moviesLoading}>
          <div className={styles.movieStats}>
            <div className={styles.statItem}>
              <div className={styles.statNumber}>{movies?.length || 0}</div>
              <div className={styles.statLabel}>Total Movies</div>
            </div>
            <div className={styles.statItem}>
              <div className={styles.statNumber}>
                {movies?.filter(movie => movie.hasFile).length || 0}
              </div>
              <div className={styles.statLabel}>Downloaded</div>
            </div>
            <div className={styles.statItem}>
              <div className={styles.statNumber}>
                {movies?.filter(movie => movie.monitored).length || 0}
              </div>
              <div className={styles.statLabel}>Monitored</div>
            </div>
          </div>
        </Card>

        <Card title="Recent Activity">
          <div className={styles.activityPlaceholder}>
            <p>Activity tracking will be implemented in future updates.</p>
          </div>
        </Card>
      </div>

      {errorChecks.length > 0 && (
        <div className={styles.healthIssues}>
          <Card title="Health Issues" variant="outlined">
            <div className={styles.issuesList}>
              {errorChecks.map((check, index) => (
                <div key={index} className={styles.issue}>
                  <div className={styles.issueHeader}>
                    <span className={styles.issueType}>{check.source}</span>
                    <span className={styles.issueStatus}>Error</span>
                  </div>
                  <div className={styles.issueMessage}>{check.message}</div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}
    </div>
  );
};
