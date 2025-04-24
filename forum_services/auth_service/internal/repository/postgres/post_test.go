// internal/repository/postgres/post_test.go
package postgres_test

import (
	"context"
	"database/sql"
	"forum-service/internal/repository/postgres"
	"forum_project/forum_services/auth_service/internal/entity"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPostRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := postgres.NewPostRepository(db)

	now := time.Now()
	post := entity.Post{
		TopicID:   1,
		UserID:    1,
		Content:   "Test content",
		CreatedAt: now,
	}

	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.TopicID, post.UserID, post.Content, post.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), post)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
