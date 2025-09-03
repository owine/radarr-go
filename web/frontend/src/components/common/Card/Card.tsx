import type { HTMLAttributes, ReactNode } from 'react';
import clsx from 'clsx';
import styles from './Card.module.css';

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  subtitle?: string;
  variant?: 'default' | 'elevated' | 'flat' | 'outlined';
  size?: 'compact' | 'default' | 'spacious';
  interactive?: boolean;
  loading?: boolean;
  footer?: ReactNode;
  children?: ReactNode;
}

export const Card = ({
  title,
  subtitle,
  variant = 'default',
  size = 'default',
  interactive = false,
  loading = false,
  footer,
  children,
  className,
  ...props
}: CardProps) => {
  const cardClasses = clsx(
    styles.card,
    styles[variant],
    size !== 'default' && styles[size],
    {
      [styles.interactive]: interactive,
      [styles.loading]: loading,
      [styles.noHeader]: !title && !subtitle,
      [styles.noFooter]: !footer,
      [styles.contentOnly]: !title && !subtitle && !footer,
    },
    className
  );

  const hasHeader = title || subtitle;

  return (
    <div
      className={cardClasses}
      tabIndex={interactive ? 0 : undefined}
      role={interactive ? 'button' : undefined}
      {...props}
    >
      {hasHeader && (
        <div className={styles.header}>
          {title && <h3 className={styles.title}>{title}</h3>}
          {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
        </div>
      )}
      
      <div className={styles.content}>
        {loading ? (
          <div>Loading...</div>
        ) : (
          children
        )}
      </div>
      
      {footer && (
        <div className={styles.footer}>
          {footer}
        </div>
      )}
    </div>
  );
};

Card.displayName = 'Card';