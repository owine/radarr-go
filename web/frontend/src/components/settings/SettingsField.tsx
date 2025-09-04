import React from 'react';
import clsx from 'clsx';
import { Input } from '../common/Input/Input';
import { useSettingsForm } from './SettingsForm';
import styles from './SettingsField.module.css';

export interface SettingsFieldProps {
  name: string;
  label: string;
  description?: string;
  type?: 'text' | 'password' | 'number' | 'email' | 'url' | 'select' | 'checkbox' | 'textarea';
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  options?: { value: string | number | boolean; label: string }[];
  min?: number;
  max?: number;
  step?: number;
  rows?: number;
  helpText?: string;
  className?: string;
  suffix?: React.ReactNode;
  prefix?: React.ReactNode;
}

export const SettingsField: React.FC<SettingsFieldProps> = ({
  name,
  label,
  description,
  type = 'text',
  placeholder,
  required = false,
  disabled = false,
  options = [],
  min,
  max,
  step,
  rows = 4,
  helpText,
  className,
  suffix,
  prefix,
}) => {
  const { data, updateField, errors, loading } = useSettingsForm();

  const value = data[name] ?? '';
  const error = errors[name];
  const isDisabled = disabled || loading;

  const handleChange = (newValue: any) => {
    updateField(name, newValue);
  };

  const renderField = () => {
    switch (type) {
      case 'select':
        return (
          <select
            id={name}
            value={value}
            onChange={(e) => {
              const option = options.find(opt => String(opt.value) === e.target.value);
              handleChange(option?.value ?? e.target.value);
            }}
            disabled={isDisabled}
            required={required}
            className={clsx(styles.select, error && styles.error)}
          >
            <option value="">Choose an option...</option>
            {options.map((option) => (
              <option key={String(option.value)} value={String(option.value)}>
                {option.label}
              </option>
            ))}
          </select>
        );

      case 'checkbox':
        return (
          <label className={clsx(styles.checkboxWrapper, isDisabled && styles.disabled)}>
            <input
              type="checkbox"
              checked={Boolean(value)}
              onChange={(e) => handleChange(e.target.checked)}
              disabled={isDisabled}
              className={styles.checkbox}
            />
            <span className={styles.checkboxLabel}>
              {label}
              {required && <span className={styles.required}>*</span>}
            </span>
          </label>
        );

      case 'textarea':
        return (
          <textarea
            id={name}
            value={value}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={placeholder}
            disabled={isDisabled}
            required={required}
            rows={rows}
            className={clsx(styles.textarea, error && styles.error)}
          />
        );

      default:
        return (
          <div className={styles.inputWithSuffix}>
            <Input
              id={name}
              type={type}
              value={value}
              onChange={(e) => handleChange(e.target.value)}
              placeholder={placeholder}
              disabled={isDisabled}
              required={required}
              min={min}
              max={max}
              step={step}
              label={label}
              error={Boolean(error)}
              errorText={error}
              helpText={helpText}
              rightIcon={suffix}
              leftIcon={prefix}
            />
          </div>
        );
    }
  };

  // For checkbox, render differently
  if (type === 'checkbox') {
    return (
      <div className={clsx(styles.fieldContainer, styles.checkboxContainer, className)}>
        {renderField()}
        {description && (
          <p className={styles.description}>{description}</p>
        )}
        {helpText && (
          <p className={styles.helpText}>{helpText}</p>
        )}
        {error && (
          <p className={styles.errorText}>{error}</p>
        )}
      </div>
    );
  }

  return (
    <div className={clsx(styles.fieldContainer, className)}>
      {description && (
        <p className={styles.description}>{description}</p>
      )}
      {renderField()}
    </div>
  );
};

SettingsField.displayName = 'SettingsField';
