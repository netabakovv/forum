package entities

import "time"

// @Description Модель поста
type Post struct {
	ID           int64      // идентификатор поста
	Title        string     // заголовок
	Content      string     // текст поста
	AuthorID     int64      // ID пользователя, написавшего пост
	AuthorName   string     // имя автора
	CreatedAt    time.Time  // время создания
	UpdatedAt    *time.Time // может быть nil, если не обновлялся
	CommentCount int32      // количество комментариев
}

// @Description Модель комментария
type Comment struct {
	ID         int64      // идентификатор комментария
	PostID     int64      // ID поста, к которому он относится
	AuthorID   int64      // ID пользователя
	AuthorName string     // имя пользователя, для фронта
	Content    string     // текст комментария
	CreatedAt  time.Time  // время создания
	UpdatedAt  *time.Time // время изменения
}

// @Description Модель сообщения в чате
type ChatMessage struct {
	ID        int64     // идентификатор сообщения
	UserID    int64     // идентификатор пользователя
	Username  string    // имя пользователя
	Content   string    // сообщение
	CreatedAt time.Time // время создания
}
