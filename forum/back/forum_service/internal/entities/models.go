package entities

import "time"

type Post struct {
	ID           int64  // идентификатор поста
	Title        string // заголовок
	Content      string // текст поста
	AuthorID     int64  // ID пользователя, написавшего пост
	AuthorName   string // необязательно, для фронта
	CreatedAt    time.Time
	UpdatedAt    *time.Time // может быть nil, если не обновлялся
	CommentCount int32      // количество комментариев (можно не хранить в БД — вычислять)
}

type Comment struct {
	ID         int64  // идентификатор комментария
	PostID     int64  // ID поста, к которому он относится
	AuthorID   int64  // ID пользователя
	AuthorName string // имя пользователя, для фронта
	Content    string // текст комментария
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type ChatMessage struct {
	ID        int64
	UserID    int64
	Username  string
	Content   string
	CreatedAt time.Time
}
