import React from 'react';
import { Clock } from 'lucide-react';

export const HistoryPage: React.FC = () => {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100%',
      textAlign: 'center',
      padding: '40px 20px',
      color: 'var(--text-muted)'
    }}>
      <Clock size={64} style={{ marginBottom: '20px' }} />
      <h2 style={{ 
        margin: '0 0 12px 0', 
        fontSize: '24px', 
        fontWeight: '600',
        color: 'var(--text-primary)'
      }}>
        History
      </h2>
      <p style={{ 
        margin: '0', 
        fontSize: '16px',
        maxWidth: '400px',
        lineHeight: '1.5'
      }}>
        Movie download and activity history will be available here. This feature will be implemented in a future sprint.
      </p>
    </div>
  );
};