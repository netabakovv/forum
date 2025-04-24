import * as React from 'react';
const { useState } = React;import { Message } from '../../types/chat';

interface ChatProps {
    messages: Message[];
    sendMessage: (text: string) => void;
    isAuthenticated: boolean;
}

export const Chat: React.FC<ChatProps> = ({ messages, sendMessage, isAuthenticated }) => {
    const [message, setMessage] = useState('');

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (message.trim()) {
            sendMessage(message);
            setMessage('');
        }
    };

    return (
        <div className="chat-container">
            <div className="messages">
                {messages.map((msg, i) => (
                    <div key={i} className="message">
                        <strong>{msg.username}:</strong> {msg.text}
                    </div>
                ))}
            </div>

            {isAuthenticated && (
                <form onSubmit={handleSubmit}>
                    <input
                        type="text"
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        placeholder="Type your message..."
                    />
                    <button type="submit">Send</button>
                </form>
            )}
        </div>
    );
};