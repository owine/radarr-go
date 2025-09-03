import React, { useState, useEffect, useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '../hooks/redux';
import { loginStart, loginSuccess, loginFailure, setRememberMe } from '../store/slices/authSlice';
import { addNotification } from '../store/slices/uiSlice';
import { useGetSystemStatusQuery } from '../store/api/radarrApi';
import { Button, Input, Card } from '../components/common';
import styles from './LoginPage.module.css';

export const LoginPage = () => {
  const dispatch = useAppDispatch();
  const { error: authError, isLoading, rememberMe } = useAppSelector(state => state.auth);
  const [apiKey, setApiKeyInput] = useState('');
  const [rememberMeLocal, setRememberMeLocal] = useState(rememberMe);
  const [validationKey, setValidationKey] = useState('');
  const [formErrors, setFormErrors] = useState<{ apiKey?: string }>({});

  // Validate API key format
  const validateApiKey = useCallback((key: string): string | null => {
    if (!key.trim()) {
      return 'API key is required';
    }
    if (key.length < 10) {
      return 'API key must be at least 10 characters long';
    }
    if (!/^[a-zA-Z0-9]+$/.test(key)) {
      return 'API key should contain only letters and numbers';
    }
    return null;
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate form
    const apiKeyError = validateApiKey(apiKey);
    if (apiKeyError) {
      setFormErrors({ apiKey: apiKeyError });
      return;
    }
    
    setFormErrors({});
    dispatch(loginStart());
    dispatch(setRememberMe(rememberMeLocal));
    
    // Set validation key to trigger API validation
    setValidationKey(apiKey);
  };

  // Use the system status query to validate the API key
  const { data: statusData, error: statusError, isLoading: statusLoading } = useGetSystemStatusQuery(undefined, {
    skip: !validationKey,
  });

  // Handle API key validation result
  useEffect(() => {
    if (!validationKey || statusLoading) return;

    if (statusError) {
      let errorMessage = 'Invalid API key. Please check your credentials.';
      
      if ('status' in statusError) {
        switch (statusError.status) {
          case 401:
            errorMessage = 'Invalid API key. Please check your credentials.';
            break;
          case 403:
            errorMessage = 'Access denied. Your API key may not have sufficient permissions.';
            break;
          case 404:
            errorMessage = 'Radarr server not found. Please check the server URL.';
            break;
          case 500:
            errorMessage = 'Server error. Please try again later.';
            break;
          default:
            errorMessage = 'Connection failed. Please check your network connection.';
        }
      }
      
      dispatch(loginFailure(errorMessage));
      setValidationKey('');
    } else if (statusData) {
      // Extract user info from system status if available
      const user = {
        username: statusData.authentication?.username || 'User',
        permissions: statusData.authentication?.permissions || [],
      };
      
      dispatch(loginSuccess({ 
        apiKey: validationKey, 
        user, 
        rememberMe: rememberMeLocal 
      }));
      
      dispatch(addNotification({
        type: 'success',
        title: 'Login Successful',
        message: `Welcome to Radarr${user.username !== 'User' ? `, ${user.username}` : ''}!`,
      }));
      
      setValidationKey('');
    }
  }, [validationKey, statusLoading, statusError, statusData, dispatch, rememberMeLocal]);

  // Clear form errors when user starts typing
  useEffect(() => {
    if (apiKey && formErrors.apiKey) {
      setFormErrors(prev => ({ ...prev, apiKey: undefined }));
    }
  }, [apiKey, formErrors.apiKey]);

  return (
    <div className={styles.container}>
      <div className={styles.loginCard}>
        <Card title="Welcome to Radarr" size="spacious">
          <form onSubmit={handleSubmit} className={styles.form}>
            <div className={styles.description}>
              <p>Enter your API key to access the Radarr interface.</p>
              <p className={styles.helpText}>
                You can find your API key in the Radarr settings under General → Security.
              </p>
            </div>
            
            <Input
              label="API Key"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKeyInput(e.target.value)}
              placeholder="Enter your API key (e.g., abc123def456...)"
              required
              error={!!(formErrors.apiKey || authError)}
              errorText={formErrors.apiKey || authError || undefined}
              disabled={isLoading}
              autoComplete="off"
              spellCheck={false}
            />
            
            <div className={styles.checkboxContainer}>
              <label className={styles.checkbox}>
                <input
                  type="checkbox"
                  checked={rememberMeLocal}
                  onChange={(e) => setRememberMeLocal(e.target.checked)}
                  disabled={isLoading}
                />
                <span className={styles.checkboxLabel}>
                  Remember me (keeps you logged in)
                </span>
              </label>
            </div>
            
            <Button
              type="submit"
              fullWidth
              loading={isLoading}
              disabled={!apiKey.trim() || isLoading}
            >
              {isLoading ? 'Authenticating...' : 'Sign In'}
            </Button>
            
            <div className={styles.helpLinks}>
              <p className={styles.helpLink}>
                Need help? <a href="#" onClick={(e) => {
                  e.preventDefault();
                  dispatch(addNotification({
                    type: 'info',
                    title: 'Finding Your API Key',
                    message: 'Go to Settings → General → Security in your Radarr interface to find or generate a new API key.',
                  }));
                }}>How to find your API key</a>
              </p>
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
};