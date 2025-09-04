import React, { useState } from 'react';
import { SettingsActions } from './SettingsActions';
import { SettingsFormProvider } from './useSettingsForm';
import styles from './SettingsForm.module.css';

export interface SettingsFormProps<T extends Record<string, unknown>> {
  data: T;
  originalData?: T;
  onSave: (data: T) => Promise<void>;
  onReset?: () => void;
  loading?: boolean;
  children: React.ReactNode;
  title?: string;
  description?: string;
  testButton?: {
    label: string;
    onTest: () => Promise<void>;
    loading?: boolean;
  };
}

export function SettingsForm<T extends Record<string, unknown>>({
  data,
  originalData,
  onSave,
  onReset,
  loading = false,
  children,
  title,
  description,
  testButton,
}: SettingsFormProps<T>) {
  const [formData, setFormData] = useState<T>(data);
  const [isDirty, setIsDirty] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const hasChanges = React.useMemo(() => {
    if (!originalData) return isDirty;
    return JSON.stringify(formData) !== JSON.stringify(originalData);
  }, [formData, originalData, isDirty]);

  const updateField = React.useCallback((field: keyof T, value: unknown) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
    setIsDirty(true);

    // Clear field error when user starts typing
    setErrors(prev => {
      if (prev[field as string]) {
        const newErrors = { ...prev };
        delete newErrors[field as string];
        return newErrors;
      }
      return prev;
    });
  }, []);

  const handleSave = async () => {
    try {
      setIsSaving(true);
      setErrors({});
      await onSave(formData);
      setIsDirty(false);
    } catch (error: unknown) {
      // Handle validation errors
      if (error && typeof error === 'object' && 'response' in error) {
        const apiError = error as { response?: { data?: { errors?: Record<string, string> } } };
        if (apiError.response?.data?.errors) {
          setErrors(apiError.response.data.errors);
        }
      } else {
        console.error('Failed to save settings:', error);
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    if (originalData) {
      setFormData(originalData);
    } else if (onReset) {
      onReset();
    }
    setIsDirty(false);
    setErrors({});
  };

  const formContext = React.useMemo(
    () => ({
      data: formData,
      updateField,
      errors,
      loading: loading || isSaving,
    }),
    [formData, updateField, errors, loading, isSaving]
  );

  return (
    <form
      className={styles.settingsForm}
      onSubmit={(e) => {
        e.preventDefault();
        handleSave();
      }}
    >
      {(title || description) && (
        <div className={styles.header}>
          {title && <h2 className={styles.title}>{title}</h2>}
          {description && <p className={styles.description}>{description}</p>}
        </div>
      )}

      <div className={styles.content}>
        <SettingsFormProvider value={formContext}>
          {children}
        </SettingsFormProvider>
      </div>

      <SettingsActions
        hasChanges={hasChanges}
        onSave={handleSave}
        onReset={handleReset}
        saveLoading={isSaving}
        testButton={testButton}
      />
    </form>
  );
}

