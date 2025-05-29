import React, { createContext, useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import AuthService from '../services/AuthService';

export const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(AuthService.getCurrentUser());
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const refreshTimeoutId = useRef(null);

  /** 
   * Запускает setTimeout так, чтобы за 30 сек до метки expiresAt перевызвать refreshToken.
   * expiresAtTimestamp — Unix-время в секундах, когда accessToken истекает.
   */
  const scheduleRefresh = (expiresAtTimestamp) => {
    console.log('scheduleRefresh:', { expiresAtTimestamp });
    if (refreshTimeoutId.current) {
      console.log('scheduleRefresh: убираю старый таймер');
      clearTimeout(refreshTimeoutId.current);
    }

    const nowMs = Date.now();
    const expiresMs = expiresAtTimestamp * 1000;
    const refreshAtMs = expiresMs - 30 * 1000;
    const delayMs = refreshAtMs - nowMs;
    console.log('scheduleRefresh:', { nowMs, expiresMs, refreshAtMs, delayMs });

    if (delayMs <= 0) {
      console.log('scheduleRefresh: delayMs <= 0 → immediate triggerRefresh');
      triggerRefresh();
    } else {
      console.log(`scheduleRefresh: ставлю таймер на ${delayMs}ms`);
      refreshTimeoutId.current = setTimeout(() => {
        console.log('setTimeout сработал → triggerRefresh()');
        triggerRefresh();
      }, delayMs);
    }
  };


  /**
   * Вызывается либо сразу (delayMs <= 0), либо через setTimeout.
   * Делает refreshToken и заново вызывает scheduleRefresh для нового expires_at.
   */
  const triggerRefresh = async () => {
    console.log('triggerRefresh: был вызван');
    try {
      const newExpiresAt = await AuthService.refreshToken();
      console.log('triggerRefresh: новый expiresAt =', newExpiresAt);

      if (newExpiresAt) {
        console.log('triggerRefresh: ставлю новый таймер');
        scheduleRefresh(newExpiresAt);
      } else {
        console.log('triggerRefresh: пришёл undefined expiresAt, больше не ставлю таймер');
        logout();
      }
    } catch (err) {
      console.error('triggerRefresh: не удалось обновить токен:', err);
      logout();
    }
  };


  // При монтировании контекста проверяем, есть ли уже залогиненный пользователь
  useEffect(() => {
    const user = AuthService.getCurrentUser();
    console.log('useEffect (mount): currentUser =', user);
    setCurrentUser(user);
    setLoading(false);

    if (user) {
      const storedExpiresAt = parseInt(localStorage.getItem('expiresAt'), 10);
      console.log('useEffect: storedExpiresAt =', storedExpiresAt);
      if (storedExpiresAt) {
        scheduleRefresh(storedExpiresAt);
      } else {
        console.log('useEffect: нет expiresAt, таймер не ставлю');
      }
    }
    // …
  }, []);


  /** Логика логина: 
   * 1) вызываем AuthService.login
   * 2) устанавливаем currentUser
   * 3) получаем новый expiresAt из localStorage (записано login) и расставляем таймер
   */
  const login = async (username, password) => {
    const user = await AuthService.login(username, password);
    setCurrentUser(user);

    // Берём expiresAt сразу из AuthService.login (который записал в localStorage):
    const expiresAt = parseInt(localStorage.getItem('expiresAt'), 10);
    if (expiresAt) {
      scheduleRefresh(expiresAt);
    }

    return user;
  };

  /** Логика логаута: очистка localStorage и переход на /login */
  const logout = () => {
    AuthService.logout();
    setCurrentUser(null);
    if (refreshTimeoutId.current) {
      clearTimeout(refreshTimeoutId.current);
    }
    navigate('/login');
  };

  const register = async (username, password) => {
    return await AuthService.register(username, password);
  };

  return (
    <AuthContext.Provider value={{ currentUser, loading, login, logout, register }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContext;
