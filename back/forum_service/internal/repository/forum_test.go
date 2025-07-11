package repository_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/netabakovv/forum/back/forum_service/internal/entities"
	"github.com/netabakovv/forum/back/forum_service/internal/repository"
	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.PostRepository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	log := logger.NewStdLogger()
	repo := repository.NewPostRepository(db, log)
	return db, mock, repo
}

func TestCreatePost(t *testing.T) {
	db, mock, repo := setup(t)
	defer db.Close()

	post := &entities.Post{
		Title:      "Test Title",
		Content:    "Test Content",
		AuthorID:   1,
		AuthorName: "user",
	}

	mock.ExpectQuery(`INSERT INTO posts`).
		WithArgs(post.Title, post.Content, post.AuthorID, post.AuthorName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(1, time.Now()))

	err := repo.CreatePost(context.Background(), post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPostByID(t *testing.T) {
	db, mock, repo := setup(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id, title, content, author_id, username, created_at, updated_at,
	       (SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count
	FROM posts p
	WHERE id = $1
`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "content", "author_id", "username", "created_at", "updated_at", "comment_count",
		}).AddRow(1, "Title", "Content", 2, "user", now, sql.NullTime{}, 3))

	post, err := repo.GetPostByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)
	assert.Equal(t, "user", post.AuthorName)
	assert.Equal(t, "Title", post.Title)
	assert.Equal(t, int32(3), post.CommentCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePost(t *testing.T) {
	db, mock, repo := setup(t)
	defer db.Close()

	post := &entities.Post{
		ID:      1,
		Title:   "Updated",
		Content: "Updated content",
	}

	mock.ExpectExec(`UPDATE posts SET`).
		WithArgs(post.Title, post.Content, post.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdatePost(context.Background(), post)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePost(t *testing.T) {
	db, mock, repo := setup(t)
	defer db.Close()

	mock.ExpectExec(`DELETE FROM posts`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeletePost(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPosts(t *testing.T) {
	db, mock, repo := setup(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, title, content, author_id, username, created_at, updated_at,`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "content", "author_id", "username", "created_at", "updated_at", "comment_count",
		}).AddRow(1, "Title", "Content", 2, "user", now, sql.NullTime{}, 0))

	posts, err := repo.Posts(context.Background())
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, int64(1), posts[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func setupComment(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.CommentRepository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	repo := repository.NewCommentRepository(db, logger.NewStdLogger())
	return db, mock, repo
}

func TestCreateComment(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	comment := &entities.Comment{
		PostID:     1,
		AuthorID:   2,
		Content:    "Test comment",
		AuthorName: "user",
	}

	mock.ExpectQuery(`INSERT INTO comments`).
		WithArgs(comment.PostID, comment.AuthorID, comment.AuthorName, comment.Content, sqlmock.AnyArg(), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.CreateComment(context.Background(), comment)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), comment.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCommentByID(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, post_id, author_id, username, content, created_at, updated_at`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "post_id", "author_id", "username", "content", "created_at", "updated_at",
		}).AddRow(1, 1, 2, "user", "test", now, nil))

	comment, err := repo.GetCommentByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), comment.ID)
	assert.Equal(t, int64(1), comment.PostID)
	assert.Equal(t, int64(2), comment.AuthorID)
	assert.Equal(t, "user", comment.AuthorName)
	assert.Equal(t, "test", comment.Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestGetByPostID(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, post_id, author_id, username, content, created_at, updated_at`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "post_id", "author_id", "username", "content", "created_at", "updated_at",
		}).AddRow(1, 1, 2, "user", "content1", now, nil).
			AddRow(2, 1, 3, "user2", "content2", now, nil))

	comments, err := repo.GetByPostID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, int64(1), comments[0].PostID)
	assert.Equal(t, "content1", comments[0].Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByUserID(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT id, post_id, author_id, username, content, created_at, updated_at`).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "post_id", "author_id", "username", "content", "created_at", "updated_at",
		}).AddRow(1, 1, 2, "user", "text", now, nil))

	comments, err := repo.GetByUserID(context.Background(), 2)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, int64(2), comments[0].AuthorID)
	assert.Equal(t, "user", comments[0].AuthorName)
	assert.Equal(t, "text", comments[0].Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateComment(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	comment := &entities.Comment{
		ID:      1,
		Content: "Updated content",
	}

	mock.ExpectExec(`UPDATE comments SET content = \$1, updated_at = \$2 WHERE id = \$3`).
		WithArgs(comment.Content, sqlmock.AnyArg(), comment.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateComment(context.Background(), comment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteComment(t *testing.T) {
	db, mock, repo := setupComment(t)
	defer db.Close()

	mock.ExpectExec(`DELETE FROM comments WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteComment(context.Background(), 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func setupChat(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.ChatRepository) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	repo := repository.NewChatRepository(db, logger.NewStdLogger())
	return db, mock, repo
}

func TestSaveMessage(t *testing.T) {
	db, mock, repo := setupChat(t)
	defer db.Close()

	msg := &entities.ChatMessage{
		UserID:   1,
		Content:  "Hello",
		Username: "user",
	}

	mock.ExpectQuery(`INSERT INTO chat_messages \(user_id, username, content, created_at\) VALUES \(\$1, \$2, \$3, NOW\(\)\) RETURNING id`).
		WithArgs(msg.UserID, msg.Username, msg.Content).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.SaveMessage(context.Background(), msg.UserID, msg.Username, msg.Content)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMessages(t *testing.T) {
	db, mock, repo := setupChat(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery(`SELECT cm\.id, cm\.user_id, cm\.username, cm\.content, cm\.created_at FROM chat_messages cm ORDER BY cm\.created_at`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "username", "content", "created_at",
		}).AddRow(1, 1, "alice", "Hello", now).
			AddRow(2, 2, "bob", "Hi", now))

	messages, err := repo.GetMessages(context.Background())
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, int64(2), messages[1].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteOldMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := logger.NewStdLogger()
	repo := repository.NewChatRepository(db, log)

	ctx := context.Background()
	cutoffTime := time.Now().Add(-24 * time.Hour)

	mock.ExpectExec("DELETE FROM chat_messages WHERE created_at <").
		WithArgs(cutoffTime).
		WillReturnResult(sqlmock.NewResult(0, 5)) // допустим, удалено 5 строк

	err = repo.DeleteOldMessages(ctx, cutoffTime)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
