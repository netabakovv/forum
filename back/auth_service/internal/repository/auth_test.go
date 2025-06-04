package repository_test

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/netabakovv/forum/back/auth_service/internal/entities"
	"github.com/netabakovv/forum/back/auth_service/internal/repository"
	"github.com/netabakovv/forum/back/pkg/logger/mocks"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

func TestUserRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewUserRepository(db, mockLogger)

	user := &entities.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		IsAdmin:      true,
	}

	mockLogger.EXPECT().Info("creating user", gomock.Any())

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(user.Username, user.PasswordHash, user.IsAdmin).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, int64(1), user.ID)
}

func TestUserRepo_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)

	repo := repository.NewUserRepository(db, mockLogger)

	user := &entities.User{
		Username:     "testuser",
		PasswordHash: "hash",
		IsAdmin:      true,
	}

	mockLogger.EXPECT().
		Info(gomock.Any(), gomock.Any()).
		Times(1)

	mockLogger.EXPECT().
		Error(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	query := ``

	mock.ExpectQuery(query).
		WithArgs(user.Username, user.PasswordHash, user.IsAdmin).
		WillReturnError(sql.ErrConnDone) // имитация ошибки подключения

	err = repo.Create(context.Background(), user)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create user")
}

func TestUserRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewUserRepository(db, mockLogger)

	username := "testuser"
	id := int64(1)
	expected := &entities.User{
		ID:           id,
		Username:     username,
		PasswordHash: "hash",
		IsAdmin:      true,
	}

	query := regexp.QuoteMeta(`
        SELECT id, username, password_hash, created_at, is_admin
        FROM users 
        WHERE id = $1`)

	mock.ExpectQuery(query).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at", "is_admin"}).
			AddRow(expected.ID, expected.Username, expected.PasswordHash, time.Now(), expected.IsAdmin))

	user, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, expected.ID, user.ID)
}

func TestUserRepo_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewUserRepository(db, mockLogger)

	username := "testuser"
	expected := &entities.User{
		ID:           1,
		Username:     username,
		PasswordHash: "hash",
		IsAdmin:      true,
	}

	query := regexp.QuoteMeta(`
        SELECT id, username, password_hash, created_at , is_admin
        FROM users 
        WHERE username = $1`)

	mock.ExpectQuery(query).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "created_at", "is_admin"}).
			AddRow(expected.ID, expected.Username, expected.PasswordHash, time.Now(), expected.IsAdmin))

	user, err := repo.GetByUsername(context.Background(), username)
	require.NoError(t, err)
	require.Equal(t, expected.Username, user.Username)
}

func TestTokenRepo_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewTokenRepository(db, mockLogger)

	token := &entities.RefreshToken{
		UserID:    1,
		Token:     "refreshtoken",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	query := regexp.QuoteMeta(`
        INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`)

	mock.ExpectExec(query).
		WithArgs(token.UserID, token.Token, token.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)
}

func TestTokenRepo_GetByToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewTokenRepository(db, mockLogger)

	now := time.Now()
	query := regexp.QuoteMeta(`
        SELECT id, user_id, token, expires_at, created_at 
        FROM refresh_tokens 
        WHERE token = $1`)

	mockLogger.EXPECT().Info("getting refresh token", gomock.Any())

	mock.ExpectQuery(query).
		WithArgs("refreshtoken").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(1, 1, "refreshtoken", now, now))

	rt, err := repo.GetByToken(context.Background(), "refreshtoken")
	require.NoError(t, err)
	require.Equal(t, "refreshtoken", rt.Token)
}

func TestTokenRepo_Revoke(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewTokenRepository(db, mockLogger)

	query := regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE token = $1`)
	mock.ExpectExec(query).
		WithArgs("some-token").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Revoke(context.Background(), "some-token")
	require.NoError(t, err)
}

func TestTokenRepo_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewTokenRepository(db, mockLogger)

	query := regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	mock.ExpectExec(query).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err = repo.DeleteExpired(context.Background())
	require.NoError(t, err)
}

func TestTokenRepo_RevokeAllUserTokens(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	repo := repository.NewTokenRepository(db, mockLogger)

	query := regexp.QuoteMeta(`DELETE FROM refresh_tokens WHERE user_id = $1`)
	mock.ExpectExec(query).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	err = repo.RevokeAllUserTokens(context.Background(), 1)
	require.NoError(t, err)
}
