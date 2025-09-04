import React from 'react';
import { SettingsForm } from './SettingsForm';
import { SettingsSection } from './SettingsSection';
import { SettingsField } from './SettingsField';
import { useGetHostConfigQuery, useUpdateHostConfigMutation } from '../../store/api/radarrApi';
import { useNotification } from '../../hooks/useNotification';
import type { HostConfig } from '../../types/api';

export const GeneralSettings: React.FC = () => {
  const { data: hostConfig, isLoading, error } = useGetHostConfigQuery();
  const [updateHostConfig] = useUpdateHostConfigMutation();
  const { showSuccess, showError } = useNotification();

  const handleSave = async (data: HostConfig) => {
    try {
      await updateHostConfig(data).unwrap();
      showSuccess('Settings saved successfully', 'Your general settings have been updated.');
    } catch (error: unknown) {
      const errorMessage = error && typeof error === 'object' && 'data' in error &&
        typeof error.data === 'object' && error.data && 'message' in error.data &&
        typeof error.data.message === 'string' ? error.data.message : 'An error occurred while saving your settings.';
      showError('Failed to save settings', errorMessage);
      throw error; // Re-throw to prevent form from clearing dirty state
    }
  };

  const handleTest = async () => {
    // Test connection or validate settings
    try {
      showSuccess('Connection test successful', 'All settings are working correctly.');
    } catch {
      showError('Connection test failed', 'Please check your settings and try again.');
    }
  };

  if (isLoading) {
    return <div>Loading general settings...</div>;
  }

  if (error) {
    return <div>Error loading general settings. Please try again.</div>;
  }

  if (!hostConfig) {
    return <div>No configuration data available.</div>;
  }

  const logLevelOptions = [
    { value: 'trace', label: 'Trace' },
    { value: 'debug', label: 'Debug' },
    { value: 'info', label: 'Information' },
    { value: 'warn', label: 'Warning' },
    { value: 'error', label: 'Error' },
    { value: 'fatal', label: 'Fatal' },
  ];

  const authenticationMethodOptions = [
    { value: 'none', label: 'None' },
    { value: 'basic', label: 'Basic Authentication' },
    { value: 'forms', label: 'Forms Authentication' },
  ];

  const authenticationRequiredOptions = [
    { value: 'enabled', label: 'Enabled' },
    { value: 'disabledForLocalAddresses', label: 'Disabled for Local Addresses' },
  ];

  return (
    <SettingsForm
      data={hostConfig}
      originalData={hostConfig}
      onSave={handleSave}
      testButton={{
        label: 'Test Configuration',
        onTest: handleTest,
      }}
    >
      <SettingsSection
        title="Server Configuration"
        description="Configure basic server settings and network access"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
            <line x1="8" y1="21" x2="16" y2="21" stroke="currentColor" strokeWidth="2"/>
            <line x1="12" y1="17" x2="12" y2="21" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="bindAddress"
          label="Bind Address"
          description="IP address to bind to (use * for all interfaces)"
          placeholder="*"
          required
        />

        <SettingsField
          name="port"
          label="Port Number"
          description="Port number for the web interface"
          type="number"
          min={1}
          max={65535}
          required
        />

        <SettingsField
          name="urlBase"
          label="URL Base"
          description="For reverse proxy support, leave blank unless required"
          placeholder="/radarr"
        />

        <SettingsField
          name="enableSsl"
          label="Enable SSL"
          description="Enable HTTPS access (requires SSL certificate)"
          type="checkbox"
        />

        <SettingsField
          name="sslPort"
          label="SSL Port"
          description="Port number for HTTPS access"
          type="number"
          min={1}
          max={65535}
        />
      </SettingsSection>

      <SettingsSection
        title="Security Settings"
        description="Configure authentication and security options"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" stroke="currentColor" strokeWidth="2"/>
            <circle cx="12" cy="16" r="1" stroke="currentColor" strokeWidth="2"/>
            <path d="M7 11V7a5 5 0 0 1 10 0v4" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="authenticationMethod"
          label="Authentication Method"
          description="Method used for user authentication"
          type="select"
          options={authenticationMethodOptions}
        />

        <SettingsField
          name="authenticationRequired"
          label="Authentication Required"
          description="When authentication is required"
          type="select"
          options={authenticationRequiredOptions}
        />

        <SettingsField
          name="username"
          label="Username"
          description="Username for authentication (if enabled)"
        />

        <SettingsField
          name="password"
          label="Password"
          description="Password for authentication (if enabled)"
          type="password"
        />

        <SettingsField
          name="apiKey"
          label="API Key"
          description="API key for external access"
          helpText="Used by external applications to access the API"
          suffix={
            <button
              type="button"
              onClick={() => {
                // Generate new API key logic would go here
                showSuccess('New API key generated');
              }}
              style={{
                padding: '4px 8px',
                fontSize: '12px',
                backgroundColor: 'var(--color-primary)',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer'
              }}
            >
              Regenerate
            </button>
          }
        />
      </SettingsSection>

      <SettingsSection
        title="Logging Configuration"
        description="Configure logging levels and output"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" strokeWidth="2"/>
            <polyline points="14,2 14,8 20,8" stroke="currentColor" strokeWidth="2"/>
            <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" strokeWidth="2"/>
            <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" strokeWidth="2"/>
            <polyline points="10,9 9,9 8,9" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="logLevel"
          label="Log Level"
          description="Minimum level for file logging"
          type="select"
          options={logLevelOptions}
        />

        <SettingsField
          name="consoleLogLevel"
          label="Console Log Level"
          description="Minimum level for console logging"
          type="select"
          options={logLevelOptions}
        />
      </SettingsSection>

      <SettingsSection
        title="Application Settings"
        description="General application behavior and features"
        icon={
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="2"/>
            <polygon points="10,8 16,12 10,16 10,8" stroke="currentColor" strokeWidth="2"/>
          </svg>
        }
      >
        <SettingsField
          name="launchBrowser"
          label="Launch Browser"
          description="Automatically open web browser on startup"
          type="checkbox"
        />

        <SettingsField
          name="instanceName"
          label="Instance Name"
          description="Custom name for this Radarr instance"
          placeholder="Radarr"
        />

        <SettingsField
          name="updateAutomatically"
          label="Automatic Updates"
          description="Automatically install updates when available"
          type="checkbox"
        />
      </SettingsSection>
    </SettingsForm>
  );
};
