import { useGetSystemStatusQuery, useGetHealthQuery } from '../store/api/radarrApi';
import { Card } from '../components/common';
import styles from './SystemPage.module.css';

export const SystemPage = () => {
  const { data: systemStatus, isLoading: statusLoading } = useGetSystemStatusQuery();
  const { data: healthChecks, isLoading: healthLoading } = useGetHealthQuery();

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1>System</h1>
        <p>System information and health status</p>
      </div>

      <div className={styles.systemGrid}>
        <Card title="System Information" loading={statusLoading}>
          {systemStatus && (
            <div className={styles.systemInfo}>
              <div className={styles.infoSection}>
                <h3>Application</h3>
                <div className={styles.infoGrid}>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Name:</span>
                    <span className={styles.infoValue}>{systemStatus.appName}</span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Version:</span>
                    <span className={styles.infoValue}>{systemStatus.version}</span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Build Time:</span>
                    <span className={styles.infoValue}>
                      {new Date(systemStatus.buildTime).toLocaleString()}
                    </span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Start Time:</span>
                    <span className={styles.infoValue}>
                      {new Date(systemStatus.startTime).toLocaleString()}
                    </span>
                  </div>
                </div>
              </div>

              <div className={styles.infoSection}>
                <h3>System</h3>
                <div className={styles.infoGrid}>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>OS:</span>
                    <span className={styles.infoValue}>
                      {systemStatus.osName} {systemStatus.osVersion}
                    </span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Runtime:</span>
                    <span className={styles.infoValue}>
                      {systemStatus.runtimeName} {systemStatus.runtimeVersion}
                    </span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Mode:</span>
                    <span className={styles.infoValue}>
                      {systemStatus.isProduction ? 'Production' : 'Development'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </Card>

        <Card title="Health Checks" loading={healthLoading}>
          {healthChecks && (
            <div className={styles.healthList}>
              {healthChecks.length === 0 ? (
                <div className={styles.noIssues}>
                  <span className={styles.healthIcon}>✅</span>
                  <span>All systems healthy</span>
                </div>
              ) : (
                healthChecks.map((check, index) => (
                  <div
                    key={index}
                    className={`${styles.healthCheck} ${styles[check.type]}`}
                  >
                    <div className={styles.healthHeader}>
                      <span className={styles.healthSource}>{check.source}</span>
                      <span className={`${styles.healthType} ${styles[check.type]}`}>
                        {check.type}
                      </span>
                    </div>
                    <div className={styles.healthMessage}>{check.message}</div>
                    {check.wikiUrl && (
                      <a
                        href={check.wikiUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.wikiLink}
                      >
                        Learn more →
                      </a>
                    )}
                  </div>
                ))
              )}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};
