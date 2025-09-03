import { forwardRef } from 'react';
import type { InputHTMLAttributes, TextareaHTMLAttributes, ReactNode } from 'react';
import clsx from 'clsx';
import styles from './Input.module.css';

interface BaseProps {
  label?: string;
  required?: boolean;
  size?: 'small' | 'medium' | 'large';
  error?: boolean;
  success?: boolean;
  helpText?: string;
  errorText?: string;
  successText?: string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
}

export interface InputProps extends BaseProps, Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  multiline?: false;
}

export interface TextareaProps extends BaseProps, Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {
  multiline: true;
}

export type CombinedInputProps = InputProps | TextareaProps;

export const Input = forwardRef<HTMLInputElement | HTMLTextAreaElement, CombinedInputProps>(({
  label,
  required = false,
  size = 'medium',
  error = false,
  success = false,
  helpText,
  errorText,
  successText,
  leftIcon,
  rightIcon,
  className,
  multiline = false,
  ...props
}, ref) => {
  const inputClasses = clsx(
    styles.input,
    styles[size],
    {
      [styles.error]: error,
      [styles.success]: success && !error,
      [styles.withLeftIcon]: leftIcon,
      [styles.withRightIcon]: rightIcon,
      [styles.textarea]: multiline,
      [styles.search]: !multiline && (props as InputHTMLAttributes<HTMLInputElement>).type === 'search',
    },
    className
  );

  const displayErrorText = error && errorText;
  const displaySuccessText = success && !error && successText;
  const displayHelpText = !displayErrorText && !displaySuccessText && helpText;

  const inputElement = multiline ? (
    <textarea
      ref={ref as React.Ref<HTMLTextAreaElement>}
      className={inputClasses}
      {...(props as TextareaHTMLAttributes<HTMLTextAreaElement>)}
    />
  ) : (
    <input
      ref={ref as React.Ref<HTMLInputElement>}
      className={inputClasses}
      {...(props as InputHTMLAttributes<HTMLInputElement>)}
    />
  );

  return (
    <div className={styles.container}>
      {label && (
        <label className={styles.label}>
          {label}
          {required && <span className={styles.required}>*</span>}
        </label>
      )}
      <div className={styles.inputWrapper}>
        {leftIcon && <div className={styles.leftIcon}>{leftIcon}</div>}
        {inputElement}
        {rightIcon && <div className={styles.rightIcon}>{rightIcon}</div>}
      </div>
      {displayErrorText && (
        <div className={styles.errorText}>{errorText}</div>
      )}
      {displaySuccessText && (
        <div className={styles.successText}>{successText}</div>
      )}
      {displayHelpText && (
        <div className={styles.helpText}>{helpText}</div>
      )}
    </div>
  );
});

Input.displayName = 'Input';