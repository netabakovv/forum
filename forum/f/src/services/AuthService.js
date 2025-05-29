import { API_URL } from './api';

const AuthService = {
  login: async (username, password) => {
    try {
      const response = await fetch(`${API_URL}/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) throw new Error('Ошибка авторизации');

      const data = await response.json();
      console.log('Login response:', data);

      localStorage.setItem('accessToken', data.access_token);
      localStorage.setItem('refreshToken', data.refresh_token);

      // Сохраняем expiresAt, только если сервер вернул это поле
      if (data.expires_at != null) {
        localStorage.setItem('expiresAt', data.expires_at.toString());
      } else {
        // Если нет expires_at, удалим старое:
        localStorage.removeItem('expiresAt');
      }

      if (data.user) {
        localStorage.setItem('user', JSON.stringify(data.user));
      } else {
        localStorage.removeItem('user');
      }

      return data.user;
    } catch (error) {
      console.error('Ошибка входа:', error);
      throw error;
    }
  },



  register: async (username, password) => {
    try {
      const response = await fetch(`${API_URL}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });
      if (!response.ok) {
        const errorMessage = await response.text(); // или .json() — если сервер возвращает JSON
        console.error('Ответ от сервера:', errorMessage);
        throw new Error(errorMessage || 'Ошибка регистрации');
      }


      const data = await response.json();
      return data;
    } catch (error) {
      console.error('Ошибка регистрации:', error);
      throw error;
    }
  },

  logout: async () => {
    try {
      const accessToken = localStorage.getItem('accessToken');
      if (accessToken) {
        const response = await fetch(`${API_URL}/api/logout`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ access_token: accessToken }),
        });
        if (!response.ok) {
          // Можно здесь обработать ошибку, например, лог или throw
          console.warn('Ошибка выхода на сервере');
        }
      }
    } catch (error) {
      console.error('Ошибка при выходе:', error);
    } finally {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
    }
  },

  getCurrentUser: () => {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    return JSON.parse(userStr);
  },

  refreshToken: async () => {
    try {
      const refreshToken = localStorage.getItem('refreshToken');
      if (!refreshToken) throw new Error('Refresh token not found');

      const response = await fetch(`${API_URL}/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        // Если сервер вернул не-200, очищаем все и бросаем ошибку
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('expiresAt');
        localStorage.removeItem('user');
        throw new Error('Invalid refresh token');
      }

      const data = await response.json();
      console.log('Refresh response payload:', data);

      // Сохраняем полученные поля
      localStorage.setItem('accessToken', data.access_token);
      localStorage.setItem('refreshToken', data.refresh_token);

      // ВАЖНО: возвращаем именно expires_at, а не access_token
      if (data.expires_at != null) {
        localStorage.setItem('expiresAt', data.expires_at.toString());
      } else {
        localStorage.removeItem('expiresAt');
      }

      if (data.user) {
        localStorage.setItem('user', JSON.stringify(data.user));
      }

      // ★ Здесь мы возвращаем expires_at (Unix-секунды), а не access_token
      return data.expires_at;
    } catch (error) {
      console.error('Error refreshing token:', error);
      throw error;
    }
  },

  
  authFetch: async (url, options = {}) => {
    const accessToken = localStorage.getItem('accessToken');

    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`
    };

    try {
      let response = await fetch(url, { ...options, headers });

      if (response.status === 401) {
        const newToken = await AuthService.refreshToken();
        headers.Authorization = `Bearer ${newToken}`;
        response = await fetch(url, { ...options, headers });
      }

      if (!response.ok) throw new Error(`Ошибка API: ${response.status}`);

      return response;
    } catch (error) {
      console.error('Ошибка запроса:', error);
      throw error;
    }
  },

  // ---- Добавленные методы для работы с API ----

  // Профиль
  getProfile: async () => {
    const res = await AuthService.authFetch(`${API_URL}/profile`);
    return res.json();
  },

  // Посты
  createPost: async (postData) => {
    const res = await AuthService.authFetch(`${API_URL}/posts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(postData),
    });
    return res.json();
  },

  getPosts: async () => {
    const res = await fetch(`${API_URL}/posts`);
    return res.json();
  },

  getPostById: async (id) => {
    const res = await fetch(`${API_URL}/posts/${id}`);
    return res.json();
  },

  deletePost: async (id) => {
    const res = await AuthService.authFetch(`${API_URL}/posts/${id}`, { method: 'DELETE' });
    return res.json();
  },

  // Комментарии
  createComment: async (commentData) => {
    const res = await AuthService.authFetch(`${API_URL}/comments`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(commentData),
    });
    return res.json();
  },

  getCommentById: async (id) => {
    const res = await fetch(`${API_URL}/comments/${id}`);
    return res.json();
  },

  deleteComment: async (id) => {
    const res = await AuthService.authFetch(`${API_URL}/comments/${id}`, { method: 'DELETE' });
    return res.json();
  },

  getCommentsByPostId: async (postID) => {
    const res = await fetch(`${API_URL}/comments/post/${postID}`);
    return res.json();
  },

  // Чат
  sendMessage: async (message) => {
    const res = await AuthService.authFetch(`${API_URL}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(message),
    });
    return res.json();
  },

  getMessages: async () => {
    const res = await fetch(`${API_URL}/chat`);
    return res.json();
  },
};

export default AuthService;
