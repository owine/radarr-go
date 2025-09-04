import React from 'react';
import { SettingsForm } from './SettingsForm';
import { SettingsSection } from './SettingsSection';
import { SettingsField } from './SettingsField';
import { useGetMediaManagementConfigQuery, useUpdateMediaManagementConfigMutation } from '../../store/api/radarrApi';
import { useNotification } from '../../hooks/useNotification';
import type { MediaManagementConfig } from '../../types/api';

export const MediaManagementSettings: React.FC = () => {
  const { data: mediaConfig, isLoading, error } = useGetMediaManagementConfigQuery();
  const [updateMediaConfig] = useUpdateMediaManagementConfigMutation();
  const { showSuccess, showError } = useNotification();

  const handleSave = async (data: MediaManagementConfig) => {
    try {
      await updateMediaConfig(data).unwrap();
      showSuccess('Media management settings saved', 'Your media management settings have been updated.');
    } catch (error: unknown) {
      const errorMessage = error && typeof error === 'object' && 'data' in error &&
        typeof error.data === 'object' && error.data && 'message' in error.data &&
        typeof error.data.message === 'string' ? error.data.message : 'An error occurred while saving your settings.';
      showError('Failed to save media management settings', errorMessage);
      throw error;
    }
  };

  if (isLoading) {
    return <div>Loading media management settings...</div>;
  }

  if (error) {
    return <div>Error loading media management settings. Please try again.</div>;
  }

  if (!mediaConfig) {
    return <div>No media management configuration data available.</div>;
  }

  const downloadPropersOptions = [
    { value: 'preferAndUpgrade', label: 'Prefer and Upgrade' },
    { value: 'doNotPrefer', label: 'Do Not Prefer' },
    { value: 'doNotUpgrade', label: 'Do Not Upgrade' },
  ];

  const fileDateOptions = [
    { value: 'none', label: 'None' },
    { value: 'cinemaRelease', label: 'Cinema Release' },
    { value: 'physicalRelease', label: 'Physical Release' },
  ];

  const rescanAfterRefreshOptions = [
    { value: 'always', label: 'Always' },
    { value: 'afterManual', label: 'After Manual' },
    { value: 'never', label: 'Never' },
  ];

  return (
    <SettingsForm
      data={mediaConfig}
      originalData={mediaConfig}
      onSave={handleSave}
    >
      <SettingsSection
        title="Movie Monitoring"
        description="Configure how movies are monitored and handled"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
            <polyline points="10,8 16,12 10,16 10,8" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="autoUnmonitorPreviouslyDownloadedMovies"
          label="Auto Unmonitor Previously Downloaded Movies"
          description="Automatically unmonitor movies when they are downloaded"
          type="checkbox"
        />

        <SettingsField
          name="downloadPropersAndRepacks"
          label="Download Propers and Repacks"
          description="How to handle proper and repack releases"
          type="select"
          options={downloadPropersOptions}
        />
      </SettingsSection>

      <SettingsSection
        title="File Management"
        description="Configure file handling and organization"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" stroke="currentColor" strokeWidth="2"/>
            <polyline points="13,2 13,9 20,9" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="createEmptyMovieFolders"
          label="Create Empty Movie Folders"
          description="Create folders for movies even when not downloaded"
          type="checkbox"
        />

        <SettingsField
          name="deleteEmptyFolders"
          label="Delete Empty Folders"
          description="Automatically delete empty folders"
          type="checkbox"
        />

        <SettingsField
          name="autoRenameFolders"
          label="Auto Rename Folders"
          description="Automatically rename movie folders"
          type="checkbox"
        />

        <SettingsField
          name="fileDate"
          label="File Date"
          description="Which date to use for the file date"
          type="select"
          options={fileDateOptions}
        />
      </SettingsSection>

      <SettingsSection
        title="Importing"
        description="Configure how files are imported and processed"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" stroke="currentColor" strokeWidth="2"/>
            <polyline points="7,10 12,15 17,10" stroke="currentColor" strokeWidth="2"/>
            <line x1="12" y1="15" x2="12" y2="3" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="rescanAfterRefresh"
          label="Rescan After Refresh"
          description="When to rescan the movie folder"
          type="select"
          options={rescanAfterRefreshOptions}
        />

        <SettingsField
          name="skipFreeSpaceCheckWhenImporting"
          label="Skip Free Space Check When Importing"
          description="Skip checking free space when importing"
          type="checkbox"
        />

        <SettingsField
          name="minimumFreeSpaceWhenImporting"
          label="Minimum Free Space When Importing (MB)"
          description="Minimum free space required when importing (in MB)"
          type="number"
          min={0}
        />

        <SettingsField
          name="copyUsingHardlinks"
          label="Use Hardlinks Instead of Copy"
          description="Use hardlinks when possible instead of copying"
          type="checkbox"
        />

        <SettingsField
          name="importExtraFiles"
          label="Import Extra Files"
          description="Import extra files (subtitles, info files, etc.)"
          type="checkbox"
        />

        <SettingsField
          name="extraFileExtensions"
          label="Extra File Extensions"
          description="Comma-separated list of extra file extensions to import"
          placeholder="srt,nfo,jpg,png"
          helpText="Examples: srt,nfo,jpg,png"
        />
      </SettingsSection>

      <SettingsSection
        title="Recycle Bin"
        description="Configure the recycle bin for deleted files"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <polyline points="3,6 5,6 21,6" stroke="currentColor" strokeWidth="2"/>
            <path d="M19,6v14a2 2 0 0 1-2,2H7a2 2 0 0 1-2-2V6m3,0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2,2v2" stroke="currentColor" strokeWidth="2"/>
            <line x1="10" y1="11" x2="10" y2="17" stroke="currentColor" strokeWidth="2"/>
            <line x1="14" y1="11" x2="14" y2="17" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="recycleBin"
          label="Recycle Bin Path"
          description="Path to the recycle bin folder (leave empty to delete permanently)"
          placeholder="/path/to/recycle/bin"
          helpText="Files will be moved here instead of being permanently deleted"
        />

        <SettingsField
          name="recycleBinCleanupDays"
          label="Recycle Bin Cleanup (Days)"
          description="Automatically delete files older than this many days (0 to disable)"
          type="number"
          min={0}
        />
      </SettingsSection>

      <SettingsSection
        title="Permissions (Linux)"
        description="Configure file and folder permissions on Linux systems"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M9 12l2 2 4-4" stroke="currentColor" strokeWidth="2"/>
            <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
        collapsible
        defaultExpanded={false}
      >
        <SettingsField
          name="setPermissionsLinux"
          label="Set Permissions"
          description="Set file and folder permissions on Linux"
          type="checkbox"
        />

        <SettingsField
          name="chmodFolder"
          label="Folder chmod"
          description="Octal permissions for folders (e.g. 755)"
          placeholder="755"
        />

        <SettingsField
          name="chownGroup"
          label="chown Group"
          description="Group name or gid to set ownership"
          placeholder="media"
        />
      </SettingsSection>
    </SettingsForm>
  );
};
