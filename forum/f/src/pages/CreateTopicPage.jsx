import React from "react";
import { useState, useEffect, useContext } from "react";
import { AuthContext } from "../context/AuthContext";
import ForumService from "../services/ForumService";
import { useNavigate, Link } from "react-router-dom";

const CreateTopicPage = () => {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const { currentUser } = React.useContext(AuthContext);
  const navigate = useNavigate();
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    Error("");

    if (!title.trim() || !content.trim()) {
      Error("Заголовок и текст темы не могут быть пустыми");
      return;
    }

    try {
      const newTopic = await ForumService.createTopic(title, content);
      navigate(`/posts/${newTopic.post.id}`);
    } catch (err) {
      console.error("Ошибка при создании темы:", err);
      Error("Не удалось создать тему. Попробуйте ещё раз.");
    }
  };
  
  return (
    <div className="container mx-auto px-4 py-8">
      <Link to="/" className="text-blue-600 hover:underline mb-4 inline-block">
        &larr; Назад к списку тем
      </Link>
      
      <div className="bg-white rounded-lg shadow p-6">
        <h1 className="text-2xl font-bold mb-6">Создать новую тему</h1>
        
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label htmlFor="title" className="block text-gray-700 font-bold mb-2">
              Заголовок
            </label>
            <input
              type="text"
              id="title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full border rounded-lg p-3 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Введите заголовок темы"
              required
            />
          </div>
          
          <div className="mb-4">
            <label htmlFor="content" className="block text-gray-700 font-bold mb-2">
              Содержание
            </label>
            <textarea
              id="content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              className="w-full border rounded-lg p-3 min-h-64 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Введите содержание темы"
              required
            />
          </div>
          
          <button
            type="submit"
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            Создать тему
          </button>
        </form>
      </div>
    </div>
  );
};

export default CreateTopicPage;