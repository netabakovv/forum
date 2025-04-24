// frontend/src/pages/HomePage.tsx
import React, { useEffect, useState } from 'react';
import { useNavigate } from '../../node_modules/react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useChat } from '../hooks/useChat';
import { fetchCategories } from '../api/forum';
import { Category } from '../types/forum';
import { Chat } from '../components/Chat/Chat';
import { CategoryList } from '../components/Forum/CategoryList';

export const HomePage: React.FC = () => {
    const { user, isAuthenticated } = useAuth();
    const { messages, sendMessage } = useChat();
    const [categories, setCategories] = useState<Category[]>([]);
    const navigate = useNavigate();

    useEffect(() => {
        const loadCategories = async () => {
            try {
                const data = await fetchCategories();
                setCategories(data);
            } catch (error) {
                console.error('Failed to load categories:', error);
            }
        };

        loadCategories();
    }, []);

    const handleCategoryClick = (categoryId: number) => {
        navigate(`/category/${categoryId}`);
    };

    return (
        <div className="home-page">
            <div className="forum-content">
                <h1>Forum Categories</h1>
                <CategoryList
                    categories={categories}
                    onCategoryClick={handleCategoryClick}
                />
            </div>

            <div className="chat-sidebar">
                <Chat
                    messages={messages}
                    onSendMessage={sendMessage}
                    isAuthenticated={isAuthenticated}
                />
            </div>
        </div>
    );
};