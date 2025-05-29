import { API_URL } from './api.js';
import AuthService from './AuthService.js';

const ForumService = {
  getTopics: async () => {
    try {
      const response = await fetch(`${API_URL}/posts`);
      if (!response.ok) throw new Error('Ошибка получения тем');
      const json = await response.json();
      const normalized = json.map(post => ({
        id: post.id,
        title: post.title,
        content: post.content,
        author: post.author_username ?? 'Неизвестный',
        createdAt: (post.created_at ?? 0) * 1000,
        repliesCount: post.comment_count ?? 0,
      }));
      return normalized;

    } catch (error) {
      console.error('Ошибка получения тем:', error);
      throw error;
    }
  },

  getTopic: async (id) => {
    try {
      const response = await fetch(`${API_URL}/posts/${id}`);
      if (!response.ok) throw new Error('Ошибка получения темы');
      const json = await response.json();
      const normalized = {
        id: json.post.id,
        title: json.post.title,
        content: json.post.content,
        author: json.post.author_username ?? 'Неизвестный',
        createdAt: (json.post.created_at ?? 0) * 1000,
        repliesCount: json.post.comment_count ?? 0,
      };
      return normalized;
    } catch (error) {
      console.error('Ошибка получения темы:', error);
      throw error;
    }
  },

  getCommentsByPostId: async (postId) => {
    try {
      const response = await fetch(`${API_URL}/comments/post/${postId}`);
      if (!response.ok) throw new Error('Ошибка получения комментариев');
      const json = await response.json();
      console.log("JSIN JSIMJSON");
      console.log(json);
      const normalized = json.map(c => ({
        id: c.id,
        content: c.content,
        author: c.author_username ?? 'Неизвестный',  // Если есть username, иначе можно добавить отдельно
        userId: c.author_id,
        createdAt: (c.created_at ?? 0) * 1000,
        isAdmin: c.is_admin ?? false,  // если есть в данных
      }));
      console.log(normalized);
      return normalized;
    } catch (error) {
      console.error('Ошибка получения комментариев:', error);
      return [];
    }
  },




  createTopic: async (title, content) => {
    try {
      const user = AuthService.getCurrentUser();
      if (!user) {
        throw new Error("Пользователь не авторизован");
      }

      const body = {
        content: content,
        author_id: user.userId,   // author_id в protobuf
        author_username: user.username, // author_username в protobuf
        title: title,
      };

      const response = await AuthService.authFetch(`${API_URL}/api/posts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      return await response.json();
    } catch (error) {
      console.error('Ошибка создания темы:', error);
      throw error;
    }
  },

  createReply: async (postId, content) => {
    try {
      // Забираем currentUser из AuthService / localStorage
      // Предполагаем, что AuthService.getCurrentUser() возвращает { userId, username, … }
      const user = AuthService.getCurrentUser();
      if (!user) {
        throw new Error("Пользователь не авторизован");
      }

      const body = {
        content: content,
        author_id: user.userId,   // author_id в protobuf
        author_username: user.username, // author_username в protobuf
        post_id: parseInt(postId), // post_id в protobuf (TopicPage вызывает createReply(topicId, content))
      };

      const response = await AuthService.authFetch(`${API_URL}/api/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        throw new Error(`Ошибка API при создании комментария: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error("Ошибка создания ответа:", error);
      throw error;
    }
  },

  deleteTopic: async (id) => {
    try {
      await AuthService.authFetch(`${API_URL}/forum/topics/${id}`, {
        method: 'DELETE',
      });
      return true;
    } catch (error) {
      console.error('Ошибка удаления темы:', error);
      throw error;
    }
  },

  deleteReply: async (topicId, replyId) => {
    try {
      await AuthService.authFetch(`${API_URL}/forum/topics/${topicId}/replies/${replyId}`, {
        method: 'DELETE',
      });
      return true;
    } catch (error) {
      console.error('Ошибка удаления ответа:', error);
      throw error;
    }
  }
};

export default ForumService;