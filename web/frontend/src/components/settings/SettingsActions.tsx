import React from 'react';
import { Button } from '../common/Button/Button';
import clsx from 'clsx';
import styles from './SettingsActions.module.css';

export interface SettingsActionsProps {
  hasChanges: boolean;
  onSave: () => void;
  onReset: () => void;
  saveLoading?: boolean;
  resetLoading?: boolean;
  testButton?: {
    label: string;
    onTest: () => Promise<void>;
    loading?: boolean;
  };
  className?: string;
}

export const SettingsActions: React.FC<SettingsActionsProps> = ({
  hasChanges,
  onSave,
  onReset,
  saveLoading = false,
  resetLoading = false,
  testButton,
  className,
}) => {
  const [testLoading, setTestLoading] = React.useState(false);

  const handleTest = async () => {
    if (!testButton?.onTest) return;

    try {
      setTestLoading(true);
      await testButton.onTest();
    } catch (error) {
      console.error('Test failed:', error);
    } finally {
      setTestLoading(false);
    }
  };

  return (
    <div className={clsx(styles.settingsActions, className)}>
      <div className={styles.leftActions}>
        {testButton && (
          <Button
            variant="outline"
            onClick={handleTest}
            loading={testLoading || testButton.loading}
            disabled={saveLoading || resetLoading}
            leftIcon={
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M9 12l2 2 4-4" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
              </svg>
            }
          >
            {testButton.label}
          </Button>
        )}
      </div>

      <div className={styles.rightActions}>
        <Button
          variant="secondary"
          onClick={onReset}
          disabled={!hasChanges || saveLoading || resetLoading}
          loading={resetLoading}
        >
          Reset
        </Button>

        <Button
          variant="primary"
          onClick={onSave}
          disabled={!hasChanges || resetLoading}
          loading={saveLoading}
          leftIcon={
            hasChanges && !saveLoading ? (
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
                <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" stroke="currentColor" strokeWidth="2"/>
                <polyline points="17,21 17,13 7,13 7,21" stroke="currentColor" strokeWidth="2"/>
                <polyline points="7,3 7,8 15,8" stroke="currentColor" strokeWidth="2"/>
              </svg>
            ) : undefined
          }
        >
          Save Changes
        </Button>
      </div>

      {hasChanges && (
        <div className={styles.changeIndicator}>
          <span className={styles.changeIcon}>●</span>
          <span className={styles.changeText}>You have unsaved changes</span>
        </div>
      )}
    </div>
  );
};

SettingsActions.displayName = 'SettingsActions';
