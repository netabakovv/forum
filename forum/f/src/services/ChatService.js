import { API_URL } from './api';

const ChatService = {
  connect: (onMessage) => {
    const socket = new WebSocket(`ws://${API_URL.replace(/^http(s?):\/\//, '')}/ws/chat`);
    
    socket.onopen = () => {
      console.log('WebSocket соединение установлено');
      
      // Отправляем токен для авторизации в WebSocket
      const token = localStorage.getItem('accessToken');
      if (token) {
        socket.send(JSON.stringify({ type: 'auth', token }));
      }
    };
    
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      onMessage(message);
    };
    
    socket.onclose = () => {
      console.log('WebSocket соединение закрыто');
    };
    
    return socket;
  },
  
  sendMessage: (socket, message) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ 
        type: 'message', 
        content: message 
      }));
    }
  }
};

export default ChatService;