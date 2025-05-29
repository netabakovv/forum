import React, { useState, useEffect, useRef } from 'react';
import { AuthContext } from '../../context/AuthContext';
import ChatService from '../../services/ChatService';
import { Link, Send } from 'lucide-react';

const Chat = () => {
  const [messages, setMessages] = useState([]);
  const [message, setMessage] = useState('');
  const [socket, setSocket] = useState(null);
  const { currentUser } = React.useContext(AuthContext);
  const messagesEndRef = useRef(null);
  
  useEffect(() => {
    // Подключение к WebSocket
    const chatSocket = ChatService.connect((newMessage) => {
      setMessages((prevMessages) => [...prevMessages, newMessage]);
    });
    
    setSocket(chatSocket);
    
    // Отключение при демонтировании компонента
    return () => {
      if (chatSocket) {
        chatSocket.close();
      }
    };
  }, []);
  
  useEffect(() => {
    // Прокрутка вниз при получении новых сообщений
    scrollToBottom();
  }, [messages]);
  
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };
  
  const handleSendMessage = (e) => {
    e.preventDefault();
    
    if (!message.trim() || !currentUser) return;
    
    ChatService.sendMessage(socket, message);
    setMessage('');
  };
  
  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      <div className="h-96 overflow-y-auto p-4">
        {messages.length > 0 ? (
          messages.map((msg, index) => (
            <div key={index} className="mb-3">
              <div className={`flex ${msg.userId === currentUser?.id ? 'justify-end' : 'justify-start'}`}>
                <div className={`rounded-lg px-4 py-2 max-w-xs lg:max-w-md ${
                  msg.userId === currentUser?.id 
                    ? 'bg-blue-100 text-blue-800' 
                    : 'bg-gray-100 text-gray-800'
                }`}>
                  <div className="font-semibold text-xs text-gray-500">
                    {msg.username} • {new Date(msg.timestamp).toLocaleTimeString()}
                  </div>
                  <div className="mt-1">{msg.content}</div>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="h-full flex items-center justify-center text-gray-500">
            Сообщений пока нет
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      
      <div className="border-t p-4">
        {currentUser ? (
          <form onSubmit={handleSendMessage} className="flex">
            <input
              type="text"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Введите сообщение..."
              className="flex-1 border rounded-l-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              type="submit"
              className="bg-blue-600 text-white px-4 py-2 rounded-r-lg hover:bg-blue-700 flex items-center"
            >
              <Send className="h-5 w-5" />
            </button>
          </form>
        ) : (
          <div className="text-center text-gray-500 py-2">
            <Link to="/login" className="text-blue-600 hover:underline">Войдите</Link>, чтобы писать сообщения
          </div>
        )}
      </div>
    </div>
  );
};
export default Chat;