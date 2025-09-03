import React from 'react';
import { Button, Card } from './common';
import styles from './ErrorBoundary.module.css';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: React.ErrorInfo | null;
  errorId: string;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{
    error: Error;
    resetError: () => void;
  }>;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private resetTimeoutId: number | null = null;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    return {
      hasError: true,
      error,
      errorId,
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.setState({
      errorInfo,
    });

    // Log error to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group('🚨 Error Boundary Caught Error');
      console.error('Error:', error);
      console.error('Component Stack:', errorInfo.componentStack);
      console.groupEnd();
    }

    // In production, you might want to send to error reporting service
    this.reportError(error, errorInfo);
  }

  private reportError = (error: Error, errorInfo: React.ErrorInfo) => {
    // TODO: Send to error reporting service (e.g., Sentry, LogRocket)
    const errorReport = {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
    };

    console.warn('Error report (implement error service):', errorReport);
  };

  private resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: '',
    });

    // Clear any existing timeout
    if (this.resetTimeoutId) {
      window.clearTimeout(this.resetTimeoutId);
    }

    // Auto-retry after a delay to prevent infinite loops
    this.resetTimeoutId = window.setTimeout(() => {
      // Additional reset logic could go here
    }, 100);
  };

  private reloadPage = () => {
    window.location.reload();
  };

  private goHome = () => {
    window.location.href = '/dashboard';
  };

  componentWillUnmount() {
    if (this.resetTimeoutId) {
      window.clearTimeout(this.resetTimeoutId);
    }
  }

  render() {
    if (this.state.hasError && this.state.error) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return <FallbackComponent error={this.state.error} resetError={this.resetError} />;
      }

      // Default error UI
      return (
        <div className={styles.errorContainer}>
          <Card title="Something went wrong" variant="error" size="spacious">
            <div className={styles.errorContent}>
              <div className={styles.errorIcon}>
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none">
                  <circle 
                    cx="12" 
                    cy="12" 
                    r="10" 
                    stroke="currentColor" 
                    strokeWidth="2"
                    fill="none"
                  />
                  <line x1="15" y1="9" x2="9" y2="15" stroke="currentColor" strokeWidth="2"/>
                  <line x1="9" y1="9" x2="15" y2="15" stroke="currentColor" strokeWidth="2"/>
                </svg>
              </div>

              <div className={styles.errorMessage}>
                <h2>Oops! Something unexpected happened</h2>
                <p>
                  We're sorry, but something went wrong while loading this part of the application.
                  This error has been logged and we'll look into it.
                </p>

                {process.env.NODE_ENV === 'development' && (
                  <details className={styles.errorDetails}>
                    <summary>Error Details (Development Only)</summary>
                    <div className={styles.errorStack}>
                      <h4>Error ID: {this.state.errorId}</h4>
                      <h4>Error Message:</h4>
                      <pre>{this.state.error.message}</pre>
                      
                      {this.state.error.stack && (
                        <>
                          <h4>Stack Trace:</h4>
                          <pre>{this.state.error.stack}</pre>
                        </>
                      )}
                      
                      {this.state.errorInfo?.componentStack && (
                        <>
                          <h4>Component Stack:</h4>
                          <pre>{this.state.errorInfo.componentStack}</pre>
                        </>
                      )}
                    </div>
                  </details>
                )}
              </div>

              <div className={styles.errorActions}>
                <Button 
                  onClick={this.resetError}
                  variant="primary"
                >
                  Try Again
                </Button>
                
                <Button 
                  onClick={this.goHome}
                  variant="secondary"
                >
                  Go to Dashboard
                </Button>
                
                <Button 
                  onClick={this.reloadPage}
                  variant="ghost"
                >
                  Reload Page
                </Button>
              </div>

              <div className={styles.errorFooter}>
                <p>
                  Error ID: <code>{this.state.errorId}</code>
                </p>
                <p>
                  If this problem persists, please contact support with the error ID above.
                </p>
              </div>
            </div>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}

// Hook for handling async errors in function components
export const useErrorBoundary = () => {
  const [error, setError] = React.useState<Error | null>(null);

  const resetError = React.useCallback(() => {
    setError(null);
  }, []);

  const captureError = React.useCallback((error: Error | string) => {
    if (typeof error === 'string') {
      setError(new Error(error));
    } else {
      setError(error);
    }
  }, []);

  React.useEffect(() => {
    if (error) {
      throw error;
    }
  }, [error]);

  return { captureError, resetError };
};