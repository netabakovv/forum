import React from "react";
import { useState, useEffect, useContext } from "react";
import { AuthContext } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";


const ProfilePage = () => {
  const { currentUser } = React.useContext(AuthContext);
  const navigate = useNavigate();
  
  useEffect(() => {
    if (!currentUser) {
      navigate('/login');
    }
  }, [currentUser, navigate]);
  
  if (!currentUser) {
    return null;
  }
  
  return (
    <div className="container mx-auto px-4 py-8 max-w-md">
      <h1 className="text-2xl font-bold mb-6">Профиль пользователя</h1>
      
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-center mb-6">
          <div className="w-24 h-24 bg-blue-600 rounded-full flex items-center justify-center text-white text-2xl font-bold">
            {currentUser.username.charAt(0).toUpperCase()}
          </div>
        </div>
        
        <div className="mb-4">
          <h2 className="text-gray-500 text-sm">Имя пользователя</h2>
          <p className="text-lg font-semibold">{currentUser.username}</p>
        </div>

        
        <div className="mb-4">
          <h2 className="text-gray-500 text-sm">Роль</h2>
          <p className="text-lg font-semibold">{currentUser.isAdmin ? 'Администратор' : 'Пользователь'}</p>
        </div>
        
        <div className="mb-4">
          <h2 className="text-gray-500 text-sm">Дата регистрации</h2>
          <p className="text-lg font-semibold">{new Date(currentUser.created_at).toLocaleDateString()}</p>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;