import { useEffect } from 'react';
import { useAppSelector } from '../hooks/redux';

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const theme = useAppSelector(state => state.ui.theme);

  useEffect(() => {
    const root = document.documentElement;
    root.setAttribute('data-theme', theme);
  }, [theme]);

  return <>{children}</>;
};