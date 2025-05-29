import { useContext, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const ProtectedRoute = ({ isAdmin, children }) => {
  const { currentUser, loading } = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    if (!loading) {
      if (!currentUser) {
        navigate('/login');
      } else if (isAdmin && !currentUser.isAdmin) {
        navigate('/');
      }
    }
  }, [currentUser, loading, navigate, isAdmin]);

  if (loading) {
    return <div className="text-center py-8">Загрузка...</div>;
  }

  return children;
};

export default ProtectedRoute;