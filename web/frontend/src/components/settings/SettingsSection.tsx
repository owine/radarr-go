import React, { useState } from 'react';
import clsx from 'clsx';
import styles from './SettingsSection.module.css';

export interface SettingsSectionProps {
  title: string;
  description?: string;
  collapsible?: boolean;
  defaultExpanded?: boolean;
  children: React.ReactNode;
  icon?: React.ReactNode;
  badge?: React.ReactNode;
  className?: string;
}

export const SettingsSection: React.FC<SettingsSectionProps> = ({
  title,
  description,
  collapsible = false,
  defaultExpanded = true,
  children,
  icon,
  badge,
  className,
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  const handleToggle = () => {
    if (collapsible) {
      setIsExpanded(!isExpanded);
    }
  };

  return (
    <div className={clsx(styles.settingsSection, className)}>
      <div
        className={clsx(
          styles.header,
          collapsible && styles.collapsible,
          collapsible && !isExpanded && styles.collapsed
        )}
        onClick={handleToggle}
        role={collapsible ? 'button' : undefined}
        tabIndex={collapsible ? 0 : undefined}
        onKeyDown={(e) => {
          if (collapsible && (e.key === 'Enter' || e.key === ' ')) {
            e.preventDefault();
            handleToggle();
          }
        }}
      >
        <div className={styles.headerContent}>
          {icon && <div className={styles.icon}>{icon}</div>}

          <div className={styles.titleContainer}>
            <h3 className={styles.title}>{title}</h3>
            {description && (
              <p className={styles.description}>{description}</p>
            )}
          </div>

          {badge && (
            <div className={styles.badge}>{badge}</div>
          )}

          {collapsible && (
            <div className={styles.expandIcon}>
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                className={clsx(
                  styles.chevron,
                  isExpanded && styles.chevronExpanded
                )}
              >
                <polyline
                  points="6,9 12,15 18,9"
                  stroke="currentColor"
                  strokeWidth="2"
                />
              </svg>
            </div>
          )}
        </div>
      </div>

      {(!collapsible || isExpanded) && (
        <div
          className={clsx(
            styles.content,
            collapsible && styles.collapsibleContent
          )}
        >
          {children}
        </div>
      )}
    </div>
  );
};

SettingsSection.displayName = 'SettingsSection';
