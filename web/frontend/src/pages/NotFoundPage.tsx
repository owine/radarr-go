import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card } from '../components/common';
import styles from './NotFoundPage.module.css';

export const NotFoundPage: React.FC = () => {
  const navigate = useNavigate();

  const goHome = () => {
    navigate('/dashboard');
  };

  const goBack = () => {
    navigate(-1);
  };

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        <Card size="spacious" className={styles.card}>
          <div className={styles.errorContent}>
            <div className={styles.errorNumber}>404</div>
            
            <div className={styles.errorMessage}>
              <h1>Page Not Found</h1>
              <p>
                Sorry, we couldn't find the page you're looking for. It might have been moved, 
                deleted, or you might have typed the wrong URL.
              </p>
            </div>

            <div className={styles.suggestions}>
              <h3>Here's what you can try:</h3>
              <ul>
                <li>Check if you typed the URL correctly</li>
                <li>Go back to the previous page</li>
                <li>Visit the dashboard to explore the app</li>
                <li>Use the search feature to find movies</li>
              </ul>
            </div>

            <div className={styles.actions}>
              <Button onClick={goHome} variant="primary">
                Go to Dashboard
              </Button>
              <Button onClick={goBack} variant="secondary">
                Go Back
              </Button>
            </div>
          </div>
        </Card>
        
        <div className={styles.illustration}>
          <svg
            width="200"
            height="200"
            viewBox="0 0 200 200"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            {/* Sad face illustration */}
            <circle
              cx="100"
              cy="100"
              r="80"
              stroke="var(--text-tertiary)"
              strokeWidth="3"
              fill="none"
            />
            {/* Eyes */}
            <circle cx="75" cy="85" r="8" fill="var(--text-tertiary)" />
            <circle cx="125" cy="85" r="8" fill="var(--text-tertiary)" />
            {/* Sad mouth */}
            <path
              d="M70 130 Q100 110 130 130"
              stroke="var(--text-tertiary)"
              strokeWidth="3"
              fill="none"
              strokeLinecap="round"
            />
            {/* Floating elements */}
            <circle cx="50" cy="50" r="4" fill="var(--color-primary)" opacity="0.6" />
            <circle cx="160" cy="60" r="3" fill="var(--color-secondary)" opacity="0.6" />
            <circle cx="40" cy="140" r="5" fill="var(--color-accent)" opacity="0.6" />
            <circle cx="170" cy="120" r="3" fill="var(--color-warning)" opacity="0.6" />
          </svg>
        </div>
      </div>
    </div>
  );
};