import React from 'react';

export interface SettingsFormContextValue<T = Record<string, unknown>> {
  data: T;
  updateField: (field: keyof T, value: unknown) => void;
  errors: Record<string, string>;
  loading: boolean;
}

const SettingsFormContext = React.createContext<SettingsFormContextValue | null>(null);

export const SettingsFormProvider = SettingsFormContext.Provider;

export function useSettingsForm<T = Record<string, unknown>>(): SettingsFormContextValue<T> {
  const context = React.useContext(SettingsFormContext);
  if (!context) {
    throw new Error('useSettingsForm must be used within a SettingsForm');
  }
  return context as SettingsFormContextValue<T>;
}
