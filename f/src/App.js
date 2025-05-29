import { BrowserRouter as Router } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import Header from './components/layout/Header';
import AppRoutes from './routes/AppRoutes';

const App = () => {
  return (
    <Router>
      <AuthProvider>
        <div className="min-h-screen bg-gray-100">
          <Header />
          <main className="py-4">
            <AppRoutes />
          </main>
        </div>
      </AuthProvider>
    </Router>
  );
};

export default App;
