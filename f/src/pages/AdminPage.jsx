import { React, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';
import AuthService from '../services/AuthService';
import { API_URL } from '../services/api';

const AdminPage = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const { currentUser } = React.useContext(AuthContext);
  const navigate = useNavigate();
  
  useEffect(() => {
    // Проверка прав администратора
    if (!currentUser || !currentUser.isAdmin) {
      navigate('/');
      return;
    }
    
    // Загрузка списка пользователей
    const fetchUsers = async () => {
      try {
        const response = await AuthService.authFetch(`${API_URL}/auth/users`);
        const data = await response.json();
        setUsers(data);
      } catch (error) {
        console.error('Ошибка загрузки пользователей:', error);
      } finally {
        setLoading(false);
      }
    };
    
    fetchUsers();
  }, [currentUser, navigate]);
  
  const toggleAdminStatus = async (userId, isAdmin) => {
    try {
      const response = await AuthService.authFetch(`${API_URL}/auth/users/${userId}/admin`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ isAdmin: !isAdmin }),
      });
      
      const updatedUser = await response.json();
      
      setUsers(users.map(user => 
        user.id === userId ? { ...user, isAdmin: updatedUser.isAdmin } : user
      ));
    } catch (error) {
      console.error('Ошибка изменения статуса администратора:', error);
    }
  };
  
  const deleteUser = async (userId) => {
    if (window.confirm('Вы уверены, что хотите удалить этого пользователя?')) {
      try {
        await AuthService.authFetch(`${API_URL}/auth/users/${userId}`, {
          method: 'DELETE',
        });
        
        setUsers(users.filter(user => user.id !== userId));
      } catch (error) {
        console.error('Ошибка удаления пользователя:', error);
      }
    }
  };
  
  if (!currentUser || !currentUser.isAdmin) {
    return null;
  }
  
  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Административная панель</h1>
      
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="p-4 bg-gray-100 border-b">
          <h2 className="font-bold">Управление пользователями</h2>
        </div>
        
        {loading ? (
          <div className="p-4 text-center">Загрузка...</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Пользователь
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Статус
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Дата регистрации
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Действия
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {users.map((user) => (
                  <tr key={user.id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="font-medium text-gray-900">{user.username}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-gray-500">{user.email}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        user.isAdmin 
                          ? 'bg-blue-100 text-blue-800' 
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {user.isAdmin ? 'Администратор' : 'Пользователь'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-gray-500">
                      {new Date(user.createdAt).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => toggleAdminStatus(user.id, user.isAdmin)}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        {user.isAdmin ? 'Снять админа' : 'Сделать админом'}
                      </button>
                      <button
                        onClick={() => deleteUser(user.id)}
                        className="text-red-600 hover:text-red-900"
                        disabled={user.id === currentUser.id}
                      >
                        Удалить
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
      
      <div className="mt-8 bg-white rounded-lg shadow overflow-hidden">
        <div className="p-4 bg-gray-100 border-b">
          <h2 className="font-bold">Статистика форума</h2>
        </div>
        
        <div className="p-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-blue-50 p-4 rounded-lg">
              <div className="text-blue-500 text-lg font-bold">{users.length}</div>
              <div className="text-gray-600">Пользователей</div>
            </div>
            
            <div className="bg-green-50 p-4 rounded-lg">
              <div className="text-green-500 text-lg font-bold">--</div>
              <div className="text-gray-600">Тем</div>
            </div>
            
            <div className="bg-purple-50 p-4 rounded-lg">
              <div className="text-purple-500 text-lg font-bold">--</div>
              <div className="text-gray-600">Сообщений</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminPage;