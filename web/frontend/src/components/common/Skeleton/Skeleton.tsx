import React from 'react';
import styles from './Skeleton.module.css';

interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  borderRadius?: string | number;
  variant?: 'rectangular' | 'circular' | 'rounded';
  animation?: 'pulse' | 'wave' | 'none';
  className?: string;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = '1rem',
  borderRadius,
  variant = 'rectangular',
  animation = 'pulse',
  className = '',
}) => {
  const getVariantStyles = () => {
    switch (variant) {
      case 'circular':
        return {
          borderRadius: '50%',
          width: height, // Make it square for circular
        };
      case 'rounded':
        return {
          borderRadius: 'var(--radius-md)',
        };
      default:
        return {};
    }
  };

  const style = {
    width: typeof width === 'number' ? `${width}px` : width,
    height: typeof height === 'number' ? `${height}px` : height,
    borderRadius: typeof borderRadius === 'number' ? `${borderRadius}px` : borderRadius,
    ...getVariantStyles(),
  };

  return (
    <div
      className={`${styles.skeleton} ${styles[animation]} ${className}`}
      style={style}
      aria-label="Loading..."
      role="status"
    />
  );
};

// Predefined skeleton components for common use cases
export const SkeletonText: React.FC<{
  lines?: number;
  spacing?: string;
  className?: string;
}> = ({ lines = 1, spacing = 'var(--space-2)', className = '' }) => (
  <div className={className} style={{ display: 'flex', flexDirection: 'column', gap: spacing }}>
    {Array.from({ length: lines }, (_, index) => (
      <Skeleton
        key={index}
        height="1rem"
        width={index === lines - 1 && lines > 1 ? '60%' : '100%'}
      />
    ))}
  </div>
);

export const SkeletonAvatar: React.FC<{
  size?: number;
  className?: string;
}> = ({ size = 40, className = '' }) => (
  <Skeleton
    variant="circular"
    width={size}
    height={size}
    className={className}
  />
);

export const SkeletonButton: React.FC<{
  width?: string | number;
  className?: string;
}> = ({ width = 120, className = '' }) => (
  <Skeleton
    width={width}
    height={36}
    variant="rounded"
    className={className}
  />
);

export const SkeletonCard: React.FC<{
  className?: string;
}> = ({ className = '' }) => (
  <div className={`${styles.skeletonCard} ${className}`}>
    <Skeleton height={200} variant="rounded" />
    <div className={styles.skeletonCardContent}>
      <SkeletonText lines={2} />
      <div className={styles.skeletonCardActions}>
        <SkeletonButton width={80} />
        <SkeletonButton width={100} />
      </div>
    </div>
  </div>
);

export const SkeletonTable: React.FC<{
  rows?: number;
  columns?: number;
  className?: string;
}> = ({ rows = 5, columns = 4, className = '' }) => (
  <div className={`${styles.skeletonTable} ${className}`}>
    {/* Header */}
    <div className={styles.skeletonTableRow}>
      {Array.from({ length: columns }, (_, index) => (
        <Skeleton key={`header-${index}`} height={20} variant="rounded" />
      ))}
    </div>

    {/* Rows */}
    {Array.from({ length: rows }, (_, rowIndex) => (
      <div key={`row-${rowIndex}`} className={styles.skeletonTableRow}>
        {Array.from({ length: columns }, (_, colIndex) => (
          <Skeleton key={`cell-${rowIndex}-${colIndex}`} height={16} />
        ))}
      </div>
    ))}
  </div>
);

export const SkeletonMovieGrid: React.FC<{
  items?: number;
  className?: string;
}> = ({ items = 12, className = '' }) => (
  <div className={`${styles.skeletonMovieGrid} ${className}`}>
    {Array.from({ length: items }, (_, index) => (
      <div key={index} className={styles.skeletonMovieItem}>
        <Skeleton height={300} variant="rounded" />
        <SkeletonText lines={2} />
        <div className={styles.skeletonMovieInfo}>
          <Skeleton width={60} height={16} />
          <Skeleton width={40} height={16} />
        </div>
      </div>
    ))}
  </div>
);
