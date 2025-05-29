package repository

import (
	"back/forum_service/internal/entities"
	e "back/pkg/errors"
	"back/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultMessagesLimit = 100
	DefaultPostsLimit    = 20
)

const (
	TargetTypePost    = "post"
	TargetTypeComment = "comment"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, userID int64, content string) error
	DeleteOldMessages(ctx context.Context, before time.Time) error
	GetMessages(ctx context.Context) ([]*entities.ChatMessage, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, post *entities.Post) error
	GetPostByID(ctx context.Context, id int64) (*entities.Post, error)
	UpdatePost(ctx context.Context, post *entities.Post) error
	DeletePost(ctx context.Context, id int64) error
	Posts(ctx context.Context) ([]*entities.Post, error)
}

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *entities.Comment) error
	GetCommentByID(ctx context.Context, id int64) (*entities.Comment, error)
	GetByPostID(ctx context.Context, postID int64) ([]*entities.Comment, error)
	GetByUserID(ctx context.Context, userID int64) ([]*entities.Comment, error)
	UpdateComment(ctx context.Context, comment *entities.Comment) error
	DeleteComment(ctx context.Context, id int64) error
}

type Db struct {
	db     *sql.DB
	logger logger.Logger
}

func NewPostRepository(db *sql.DB, log logger.Logger) PostRepository {
	return &Db{db: db, logger: log}
}

func NewCommentRepository(db *sql.DB, log logger.Logger) CommentRepository {
	return &Db{db: db, logger: log}
}

func NewChatRepository(db *sql.DB, log logger.Logger) ChatRepository {
	return &Db{db: db, logger: log}
}

// --- Post Repository ---

func (r *Db) CreatePost(ctx context.Context, post *entities.Post) error {
	query := `
		INSERT INTO posts (title, content, author_id, username, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, post.Title, post.Content, post.AuthorID, post.AuthorName).
		Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		return fmt.Errorf("создание поста: %w", err)
	}
	return nil
}

func (r *Db) GetPostByID(ctx context.Context, id int64) (*entities.Post, error) {
	query := `
		SELECT id, title, content, author_id, username, created_at, updated_at,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p WHERE id = $1`

	post := &entities.Post{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.Title, &post.Content, &post.AuthorID,
		&post.AuthorName,
		&post.CreatedAt, &post.UpdatedAt, &post.CommentCount,
	)
	if err != nil {
		return nil, fmt.Errorf("получение поста: %w", err)
	}
	return post, nil
}

func (r *Db) UpdatePost(ctx context.Context, post *entities.Post) error {
	query := `UPDATE posts SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.ID)
	return err
}

func (r *Db) DeletePost(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *Db) Posts(ctx context.Context) ([]*entities.Post, error) {
	query := `
		SELECT id, title, content, author_id, username, created_at, updated_at,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
		FROM posts p ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("получение постов: %w", err)
	}
	defer rows.Close()

	var posts []*entities.Post
	for rows.Next() {
		post := &entities.Post{}
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.AuthorID,
			&post.AuthorName,
			&post.CreatedAt, &post.UpdatedAt, &post.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// --- Chat Repository ---

func (r *Db) SaveMessage(ctx context.Context, userID int64, content string) error {
	var id int64
	query := `INSERT INTO chat_messages (user_id, content, created_at) VALUES ($1, $2, NOW()) RETURNING id`
	return r.db.QueryRowContext(ctx, query, userID, content).Scan(&id)
}

func (r *Db) DeleteOldMessages(ctx context.Context, before time.Time) error {
	query := `DELETE FROM chat_messages WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		r.logger.Error("ошибка удаления старых сообщений", logger.NewField("error", err))
		return err
	}

	affected, _ := result.RowsAffected()
	r.logger.Info("удалены старые сообщения",
		logger.NewField("count", affected),
		logger.NewField("older_than", before))

	return nil
}

func (r *Db) GetMessages(ctx context.Context) ([]*entities.ChatMessage, error) {
	query := `
		SELECT cm.id, cm.user_id, cm.content, cm.created_at
		FROM chat_messages cm
		ORDER BY cm.created_at DESC
		`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		msg := &entities.ChatMessage{}
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования сообщения: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// ----------------------- CommentRepository

func (r *Db) CreateComment(ctx context.Context, comment *entities.Comment) error {
	query := `
        INSERT INTO comments (post_id, author_id, username, content, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = nil // new comment, no update yet

	return r.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.AuthorID,
		comment.AuthorName,
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)
}

func (r *Db) GetCommentByID(ctx context.Context, id int64) (*entities.Comment, error) {
	query := `
        SELECT id, post_id, author_id, username, content, created_at, updated_at
        FROM comments
        WHERE id = $1
    `
	row := r.db.QueryRowContext(ctx, query, id)

	var comment entities.Comment
	err := row.Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&comment.AuthorName,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.ErrCommentNotFound
		}
		return nil, err
	}

	return &comment, nil
}

func (r *Db) GetByPostID(ctx context.Context, postID int64) ([]*entities.Comment, error) {
	query := `
        SELECT id, post_id, author_id, username, content, created_at, updated_at
        FROM comments
        WHERE post_id = $1
        ORDER BY created_at ASC
    `
	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.AuthorName,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (r *Db) GetByUserID(ctx context.Context, userID int64) ([]*entities.Comment, error) {
	query := `
        SELECT id, post_id, author_id, username, content, created_at, updated_at
        FROM comments
        WHERE author_id = $1
        ORDER BY created_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*entities.Comment
	for rows.Next() {
		var comment entities.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.AuthorName,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	return comments, nil
}

func (r *Db) UpdateComment(ctx context.Context, comment *entities.Comment) error {
	now := time.Now()
	comment.UpdatedAt = &now

	query := `
        UPDATE comments
        SET content = $1, updated_at = $2
        WHERE id = $3
    `
	res, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no comment found with id %d", comment.ID)
	}

	return nil
}

func (r *Db) DeleteComment(ctx context.Context, id int64) error {
	query := `DELETE FROM comments WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no comment found with id %d", id)
	}

	return nil
}
