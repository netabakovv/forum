import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { AuthContext } from '../../context/AuthContext';
import { User } from 'lucide-react';


const Header = () => {
  const { currentUser, logout } = React.useContext(AuthContext);
  const [menuOpen, setMenuOpen] = useState(false);
  
  return (
    <header className="bg-blue-600 text-white shadow-md">
      <div className="container mx-auto px-4 py-3 flex justify-between items-center">
        <div className="flex items-center">
          <Link to="/" className="text-xl font-bold">ФорумGo</Link>
        </div>
        
        <div className="relative">
          {currentUser ? (
            <>
              <button 
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center space-x-2 focus:outline-none"
              >
                <span>{currentUser.username}</span>
                <User className="h-5 w-5" />
              </button>
              
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-10">
                  <Link 
                    to="/profile" 
                    className="block px-4 py-2 text-gray-800 hover:bg-gray-100"
                    onClick={() => setMenuOpen(false)}
                  >
                    Профиль
                  </Link>
                  {currentUser.isAdmin && (
                    <Link 
                      to="/admin" 
                      className="block px-4 py-2 text-gray-800 hover:bg-gray-100"
                      onClick={() => setMenuOpen(false)}
                    >
                      Админ панель
                    </Link>
                  )}
                  <button 
                    onClick={() => {
                      logout();
                      setMenuOpen(false);
                    }}
                    className="block w-full text-left px-4 py-2 text-gray-800 hover:bg-gray-100"
                  >
                    Выйти
                  </button>
                </div>
              )}
            </>
          ) : (
            <div className="flex space-x-4">
              <Link to="/login" className="hover:underline">Вход</Link>
              <Link to="/register" className="hover:underline">Регистрация</Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;