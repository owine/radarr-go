import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import { store } from './store';
import { AppRouter } from './components/AppRouter';
import { ThemeProvider } from './components/ThemeProvider';
import { NotificationContainer } from './components/NotificationContainer';
import './styles/globals.css';

function App() {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <ThemeProvider>
          <div id="app">
            <AppRouter />
            <NotificationContainer />
          </div>
        </ThemeProvider>
      </BrowserRouter>
    </Provider>
  );
}

export default App;
