# Settings Components

This directory contains the comprehensive settings infrastructure for Radarr-Go. The settings system is designed to be:

- **Modular**: Each settings category is a separate component
- **Reusable**: Common patterns are abstracted into reusable components
- **Type-safe**: Full TypeScript support with proper typing
- **Responsive**: Works on desktop, tablet, and mobile devices
- **Accessible**: Proper ARIA labels and keyboard navigation
- **Validated**: Form validation with error handling
- **Optimistic**: Optimistic updates with rollback on failure

## Architecture

### Core Components

#### `SettingsForm`

Main form wrapper that provides:

- Form state management with dirty tracking
- Save/reset functionality with confirmation
- Error handling and validation
- Loading states and optimistic updates
- Test functionality for configuration validation
- Context for child components

#### `SettingsSection`

Collapsible section container that provides:

- Title and description with optional icons
- Collapsible functionality
- Consistent spacing and styling
- Badge support for notifications

#### `SettingsField`

Individual form field wrapper that provides:

- Integration with existing Input component
- Multiple input types (text, password, number, select, checkbox, textarea)
- Label, description, and help text support
- Error state display
- Prefix/suffix icon support
- Required field indication

#### `SettingsActions`

Sticky footer with action buttons:

- Save Changes (enabled when form is dirty)
- Reset (revert to original values)
- Test Configuration (optional, for validating settings)
- Change indicator when form has unsaved changes

### Usage Examples

#### Basic Settings Component

```tsx
import React from 'react';
import { SettingsForm, SettingsSection, SettingsField } from '../settings';
import { useGetConfigQuery, useUpdateConfigMutation } from '../../store/api/radarrApi';
import { useNotification } from '../../hooks/useNotification';

export const MySettings: React.FC = () => {
  const { data: config, isLoading } = useGetConfigQuery();
  const [updateConfig] = useUpdateConfigMutation();
  const { showSuccess, showError } = useNotification();

  const handleSave = async (data) => {
    try {
      await updateConfig(data).unwrap();
      showSuccess('Settings saved');
    } catch (error) {
      showError('Failed to save', error.message);
      throw error; // Prevent form from clearing dirty state
    }
  };

  if (isLoading) return <div>Loading...</div>;

  return (
    <SettingsForm data={config} onSave={handleSave}>
      <SettingsSection title="Basic Settings">
        <SettingsField
          name="serverName"
          label="Server Name"
          description="Name of your server instance"
          required
        />

        <SettingsField
          name="port"
          label="Port"
          type="number"
          min={1}
          max={65535}
          required
        />

        <SettingsField
          name="enableSsl"
          label="Enable SSL"
          type="checkbox"
        />
      </SettingsSection>
    </SettingsForm>
  );
};
```

#### Advanced Settings with Sections and Test

```tsx
export const AdvancedSettings: React.FC = () => {
  const { data: config } = useGetConfigQuery();
  const [updateConfig] = useUpdateConfigMutation();
  const { showSuccess, showError } = useNotification();

  const handleSave = async (data) => {
    await updateConfig(data).unwrap();
    showSuccess('Settings saved');
  };

  const handleTest = async () => {
    // Test the configuration
    try {
      await testConnection(config);
      showSuccess('Connection test successful');
    } catch (error) {
      showError('Connection test failed');
    }
  };

  return (
    <SettingsForm
      data={config}
      onSave={handleSave}
      testButton={{
        label: 'Test Connection',
        onTest: handleTest,
      }}
    >
      <SettingsSection
        title="Connection Settings"
        description="Configure external service connections"
        icon={<NetworkIcon />}
        collapsible
      >
        <SettingsField
          name="apiUrl"
          label="API URL"
          type="url"
          required
          placeholder="https://api.example.com"
        />

        <SettingsField
          name="apiKey"
          label="API Key"
          type="password"
          required
          suffix={<RegenerateButton />}
        />
      </SettingsSection>

      <SettingsSection
        title="Advanced Options"
        collapsible
        defaultExpanded={false}
      >
        <SettingsField
          name="timeout"
          label="Request Timeout (seconds)"
          type="number"
          min={1}
          max={300}
          helpText="How long to wait for responses"
        />
      </SettingsSection>
    </SettingsForm>
  );
};
```

## Field Types

### Text Input

```tsx
<SettingsField
  name="serverName"
  label="Server Name"
  placeholder="My Server"
  required
/>
```

### Number Input

```tsx
<SettingsField
  name="port"
  label="Port"
  type="number"
  min={1}
  max={65535}
/>
```

### Select Dropdown

```tsx
<SettingsField
  name="logLevel"
  label="Log Level"
  type="select"
  options={[
    { value: 'debug', label: 'Debug' },
    { value: 'info', label: 'Information' },
    { value: 'error', label: 'Error' },
  ]}
/>
```

### Checkbox

```tsx
<SettingsField
  name="enableFeature"
  label="Enable Feature"
  description="Enable this awesome feature"
  type="checkbox"
/>
```

### Textarea

```tsx
<SettingsField
  name="description"
  label="Description"
  type="textarea"
  rows={4}
/>
```

### Password Input

```tsx
<SettingsField
  name="password"
  label="Password"
  type="password"
  required
/>
```

## Adding New Settings Categories

1. **Create API Endpoints** (if needed):

   ```tsx
   // In radarrApi.ts
   getMyConfig: builder.query<MyConfig, void>({
     query: () => 'config/myconfig',
     providesTags: [{ type: 'Config', id: 'MY_CONFIG' }],
   }),

   updateMyConfig: builder.mutation<MyConfig, Partial<MyConfig>>({
     query: (config) => ({
       url: 'config/myconfig',
       method: 'PUT',
       body: config,
     }),
     invalidatesTags: [{ type: 'Config', id: 'MY_CONFIG' }],
   }),
   ```

2. **Create Settings Component**:

   ```tsx
   // MySettings.tsx
   export const MySettings: React.FC = () => {
     const { data: config, isLoading } = useGetMyConfigQuery();
     const [updateConfig] = useUpdateMyConfigMutation();

     // ... implementation
   };
   ```

3. **Add to Settings Navigation**:

   ```tsx
   // In SettingsPage.tsx
   {
     id: 'my-settings',
     label: 'My Settings',
     icon: <MyIcon />,
     component: MySettings,
   }
   ```

## Best Practices

1. **Always handle errors**: Re-throw errors in save handlers to prevent clearing form state
2. **Use proper loading states**: Show loading indicators while fetching data
3. **Provide helpful descriptions**: Use field descriptions and help text
4. **Group related settings**: Use SettingsSection to organize fields
5. **Make sections collapsible**: For advanced or less commonly used settings
6. **Add test functionality**: Where applicable, provide test buttons
7. **Use proper field types**: Choose the right input type for the data
8. **Add validation**: Use required, min, max, and custom validation
9. **Follow naming conventions**: Use camelCase for field names
10. **Keep it responsive**: Settings work on all screen sizes

## Styling

The settings components use CSS modules for styling with proper responsive design:

- **Desktop**: Two-column layout with sidebar navigation
- **Tablet**: Collapsible sidebar, full-width content
- **Mobile**: Horizontal tabs, stacked layout

All components follow the existing design system and color scheme, supporting both light and dark themes.

## Testing

When creating new settings components:

1. Test form validation (required fields, min/max values)
2. Test save/reset functionality
3. Test error handling (network errors, validation errors)
4. Test responsive design on different screen sizes
5. Test keyboard navigation and accessibility
6. Test loading states and optimistic updates

The settings infrastructure provides a solid foundation for all configuration management in Radarr-Go.
