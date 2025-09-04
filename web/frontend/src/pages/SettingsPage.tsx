import React, { useState } from 'react';
import { Card } from '../components/common/Card/Card';
// import { useNotification } from '../hooks/useNotification';
import { GeneralSettings } from '../components/settings/GeneralSettings';
import { MediaManagementSettings } from '../components/settings/MediaManagementSettings';
import styles from './SettingsPage.module.css';

export interface SettingsTab {
  id: string;
  label: string;
  icon: React.ReactNode;
  component: React.ComponentType;
  badge?: number;
}

const settingsTabs: SettingsTab[] = [
  {
    id: 'general',
    label: 'General',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth="2"/>
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: GeneralSettings,
  },
  {
    id: 'media-management',
    label: 'Media Management',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
        <polyline points="14,2 14,8 20,8" stroke="currentColor" strokeWidth="2"/>
        <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" strokeWidth="2"/>
        <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" strokeWidth="2"/>
        <polyline points="10,9 9,9 8,9" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: MediaManagementSettings,
  },
  {
    id: 'quality',
    label: 'Quality',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: QualitySettings,
  },
  {
    id: 'indexers',
    label: 'Indexers',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <circle cx="11" cy="11" r="8" stroke="currentColor" strokeWidth="2"/>
        <path d="21 21l-4.35-4.35" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: IndexerSettings,
  },
  {
    id: 'download-clients',
    label: 'Download Clients',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" stroke="currentColor" strokeWidth="2"/>
        <polyline points="7,10 12,15 17,10" stroke="currentColor" strokeWidth="2"/>
        <line x1="12" y1="15" x2="12" y2="3" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: DownloadClientSettings,
  },
  {
    id: 'connect',
    label: 'Connect',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" stroke="currentColor" strokeWidth="2"/>
        <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: ConnectSettings,
  },
  {
    id: 'metadata',
    label: 'Metadata',
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
        <circle cx="8.5" cy="8.5" r="1.5" stroke="currentColor" strokeWidth="2"/>
        <polyline points="21,15 16,10 5,21" stroke="currentColor" strokeWidth="2"/>
      </svg>
    ),
    component: MetadataSettings,
  },
];

// Placeholder components - these will be implemented in subsequent tasks


function QualitySettings() {
  return <div className={styles.placeholder}>Quality Settings coming soon...</div>;
}

function IndexerSettings() {
  return <div className={styles.placeholder}>Indexer Settings coming soon...</div>;
}

function DownloadClientSettings() {
  return <div className={styles.placeholder}>Download Client Settings coming soon...</div>;
}

function ConnectSettings() {
  return <div className={styles.placeholder}>Connect Settings coming soon...</div>;
}

function MetadataSettings() {
  return <div className={styles.placeholder}>Metadata Settings coming soon...</div>;
}

export const SettingsPage = () => {
  const [activeTab, setActiveTab] = useState('general');
  // Notification hooks available for future use
  // const { showNotification, showSuccess, showError } = useNotification();

  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId);
  };

  const activeTabData = settingsTabs.find(tab => tab.id === activeTab);
  const ActiveComponent = activeTabData?.component || GeneralSettings;

  return (
    <div className={styles.settingsPage}>
      <div className={styles.header}>
        <h1 className={styles.title}>Settings</h1>
        <p className={styles.subtitle}>
          Configure your Radarr instance settings and preferences
        </p>
      </div>

      <div className={styles.container}>
        {/* Sidebar Navigation */}
        <aside className={styles.sidebar}>
          <nav className={styles.navigation}>
            {settingsTabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={`${styles.navItem} ${
                  activeTab === tab.id ? styles.active : ''
                }`}
                type="button"
              >
                <span className={styles.navIcon}>{tab.icon}</span>
                <span className={styles.navLabel}>{tab.label}</span>
                {tab.badge && tab.badge > 0 && (
                  <span className={styles.navBadge}>{tab.badge}</span>
                )}
              </button>
            ))}
          </nav>
        </aside>

        {/* Main Content */}
        <main className={styles.content}>
          <Card
            title={activeTabData?.label}
            variant="flat"
            size="spacious"
            className={styles.settingsCard}
          >
            <ActiveComponent />
          </Card>
        </main>
      </div>
    </div>
  );
};

SettingsPage.displayName = 'SettingsPage';
