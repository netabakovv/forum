import React, { useState, useEffect, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { AuthContext } from "../context/AuthContext";
import ForumService from "../services/ForumService";
import { User, Shield, Trash2, Link } from 'lucide-react'


const TopicPage = () => {
  const { id } = useParams();
  const [topic, setTopic] = useState(null);
  const [replies, setReplies] = useState([]);
  const [reply, setReply] = useState('');
  const [loading, setLoading] = useState(true);
  const { currentUser } = React.useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchTopicAndComments = async () => {
      setLoading(true);
      try {
        const topicData = await ForumService.getTopic(id);
        setTopic(topicData);

        const comments = await ForumService.getCommentsByPostId(id);
        setReplies(comments);
        console.log(topicData)
        console.log(comments)
      } catch (error) {
        setTopic(null);
        setReplies([]);
      } finally {
        setLoading(false);
      }
    };

    fetchTopicAndComments();
  }, [id]);




  const handleSubmitReply = async (e) => {
    e.preventDefault();
    if (!reply.trim() || !currentUser) return;

    try {
      await ForumService.createReply(id, reply);
      const updatedComments = await ForumService.getCommentsByPostId(id);
      setReplies(updatedComments);
      setReply('');
    } catch (error) {
      console.error('Ошибка отправки ответа:', error);
    }
  };


  const handleDeleteTopic = async () => {
    if (window.confirm('Вы уверены, что хотите удалить эту тему?')) {
      try {
        await ForumService.deleteTopic(id);
        navigate('/');
      } catch (error) {
        console.error('Ошибка удаления темы:', error);
      }
    }
  };

  const handleDeleteReply = async (replyId) => {
    if (window.confirm('Вы уверены, что хотите удалить этот ответ?')) {
      try {
        await ForumService.deleteReply(id, replyId);
        setReplies(replies.filter(r => r.id !== replyId));
      } catch (error) {
        console.error('Ошибка удаления ответа:', error);
      }
    }
  };

  if (loading) {
    return <div className="container mx-auto px-4 py-8 text-center">Загрузка...</div>;
  }

  if (!topic) {
    return <div className="container mx-auto px-4 py-8 text-center">Тема не найдена</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Link to="/" className="text-blue-600 hover:underline mb-4 inline-block">
        &larr; Назад к списку тем
      </Link>

      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <div className="flex justify-between items-start">
          <h1 className="text-2xl font-bold text-gray-900">{topic.title}</h1>
          {(currentUser?.isAdmin || currentUser?.id === topic.userId) && (
            <button
              onClick={handleDeleteTopic}
              className="text-red-600 hover:text-red-800"
            >
              <Trash2 className="h-5 w-5" />
            </button>
          )}
        </div>

        <div className="flex items-center text-sm text-gray-500 mt-2">
          <span className="flex items-center">
            <User className="w-4 h-4 mr-1" />
            {topic.author}
          </span>
          <span className="ml-4">
            {new Date(topic.createdAt).toLocaleString()}
          </span>
          {topic.isAdmin && (
            <span className="ml-4 flex items-center text-blue-600">
              <Shield className="w-4 h-4 mr-1" />
              Администратор
            </span>
          )}
        </div>

        <div className="mt-4 text-gray-700 whitespace-pre-line">
          {topic.content}
        </div>
      </div>

      <div className="mb-8">
        <h2 className="text-xl font-bold mb-4">Ответы ({replies.length})</h2>

        {replies.length > 0 ? (
          <div className="space-y-4">
            {replies.map((reply) => (
              <div key={reply.id} className="bg-white rounded-lg shadow p-4">
                <div className="flex justify-between">
                  <div className="flex items-center text-sm text-gray-500">
                    <span className="flex items-center">
                      <User className="w-4 h-4 mr-1" />
                      {reply.author}
                    </span>
                    <span className="ml-4">
                      {new Date(reply.createdAt).toLocaleString()}
                    </span>
                    {reply.isAdmin && (
                      <span className="ml-4 flex items-center text-blue-600">
                        <Shield className="w-4 h-4 mr-1" />
                        Администратор
                      </span>
                    )}
                  </div>

                  {(currentUser?.isAdmin || currentUser?.id === reply.userId) && (
                    <button
                      onClick={() => handleDeleteReply(reply.id)}
                      className="text-red-600 hover:text-red-800"
                    >
                      <Trash2 className="h-5 w-5" />
                    </button>
                  )}
                </div>

                <div className="mt-2 text-gray-700 whitespace-pre-line">
                  {reply.content}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow p-6 text-center text-gray-500">
            Пока нет ответов. Будьте первым!
          </div>
        )}
      </div>

      {currentUser ? (
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-bold mb-4">Добавить ответ</h3>
          <form onSubmit={handleSubmitReply}>
            <textarea
              value={reply}
              onChange={(e) => setReply(e.target.value)}
              className="w-full border rounded-lg p-3 min-h-32 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Введите ваш ответ..."
              required
            />
            <button
              type="submit"
              className="mt-4 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
            >
              Отправить
            </button>
          </form>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow p-6 text-center">
          <p className="text-gray-500">
            <Link to="/login" className="text-blue-600 hover:underline">Войдите</Link>, чтобы оставить ответ
          </p>
        </div>
      )}
    </div>
  );
};

export default TopicPage;