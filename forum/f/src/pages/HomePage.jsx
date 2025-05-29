import React from "react";
import { useState, useEffect, useContext } from "react";
import { AuthContext } from "../context/AuthContext";
import { Link } from "react-router-dom"
import ForumService from "../services/ForumService";
import Chat from "../components/chat/Chat";

const HomePage = () => {
  const [topics, setTopics] = useState([]);
  const [loading, setLoading] = useState(true);
  const { currentUser } = React.useContext(AuthContext);

  useEffect(() => {
    const fetchTopics = async () => {
      try {
        const data = await ForumService.getTopics();
        setTopics(data);
      } catch (error) {
        console.error('Ошибка загрузки тем:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchTopics();
  }, []);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Темы форума</h1>
        {currentUser && (
          <Link
            to="/posts/create"
            className="inline-block bg-blue-600 text-white font-semibold px-4 py-2 rounded-lg shadow hover:bg-blue-700 transition duration-200"
          >
            Создать тему
          </Link>
        )}
      </div>


      {loading ? (
        <div className="text-center py-8">Загрузка...</div>
      ) : topics.length > 0 ? (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          {topics.map((topic) => (
            <a
              key={topic.id}
              href={`/posts/${topic.id}`}
              className="block border border-gray-300 rounded-md p-4 mb-4 bg-white hover:bg-gray-100"
            >
              <h2 className="text-xl font-bold mb-2 text-blue-700">{topic.title}</h2>
              <p className="mb-2 text-gray-700">{topic.content}</p>
              <p className="text-sm text-gray-500">Автор: {topic.author}</p>
              <p className="text-sm text-gray-500">
                Дата: {new Date(topic.createdAt).toLocaleDateString()}
              </p>
              <p className="text-sm text-gray-500">Ответов: {topic.repliesCount ?? 0}</p>
            </a>
          ))}

        </div>
      ) : (
        <div className="text-center py-8 text-gray-500">
          Нет доступных тем. {currentUser?.isAdmin ? 'Создайте первую тему!' : 'Темы появятся позже.'}
        </div>

      )}


      {/* Блок чата */}
      <div className="mt-12">
        <h2 className="text-xl font-bold mb-4">Общий чат</h2>
        <Chat messagesOnly={!currentUser} />
      </div>
    </div>
  );
};

export default HomePage;