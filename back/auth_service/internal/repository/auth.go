// repository/auth.go
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/netabakovv/forum/back/auth_service/internal/entities"
	"github.com/netabakovv/forum/back/pkg/errors"
	"github.com/netabakovv/forum/back/pkg/logger"
)

// UserRepository определяет методы для работы с пользователями в БД
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	UpdateUsername(ctx context.Context, id int64, newName string) error
	Delete(ctx context.Context, id int64) error
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

// TokenRepository определяет методы для работы с refresh токенами в БД
type TokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	RevokeAllUserTokens(ctx context.Context, userID int64) error
}

// Реализация для PostgreSQL
type userRepo struct {
	db  *sql.DB
	log logger.Logger // Используем интерфейс из пакета logger
}

type tokenRepo struct {
	db  *sql.DB
	log logger.Logger // Используем интерфейс из пакета logger
}

func (r *userRepo) Create(ctx context.Context, user *entities.User) error { // Изменен возвращаемый тип
	query := `
        INSERT INTO users (username, password_hash, created_at, is_admin)
        VALUES ($1, $2, CURRENT_TIMESTAMP, $3)
        RETURNING id`

	r.log.Info("creating user",
		logger.NewField("username", user.Username),
	)

	err := r.db.QueryRowContext(ctx, query,
		user.Username, user.PasswordHash, user.IsAdmin,
	).Scan(&user.ID)

	if err != nil {
		r.log.Error("failed to create user",
			logger.NewField("error", err),
			logger.NewField("username", user.Username),
		)
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	user := &entities.User{}
	query := `
        SELECT id, username, password_hash, created_at 
        FROM users 
        WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	user := &entities.User{}
	query := `
        SELECT id, username, password_hash, created_at , is_admin
        FROM users 
        WHERE username = $1`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.IsAdmin,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) UpdateUsername(ctx context.Context, id int64, newName string) error { // Изменен возвращаемый тип
	query := `UPDATE users SET username = $1 WHERE id = $2`

	r.log.Info("updating username",
		logger.NewField("user_id", id),
		logger.NewField("new_name", newName),
	)

	result, err := r.db.ExecContext(ctx, query, newName, id) // Используем ExecContext вместо QueryContext
	if err != nil {
		r.log.Error("failed to update username",
			logger.NewField("error", err),
			logger.NewField("user_id", id),
		)
		return fmt.Errorf("update username: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.ErrDeleteFailed
	}
	return err
}

func (r *userRepo) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	var isAdmin bool
	query := `SELECT is_admin FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&isAdmin)
	if err == sql.ErrNoRows {
		return false, errors.ErrUserNotFound
	}
	return isAdmin, err
}

func NewUserRepository(db *sql.DB, logger logger.Logger) UserRepository {
	return &userRepo{
		db:  db,
		log: logger,
	}
}

func NewTokenRepository(db *sql.DB, logger logger.Logger) TokenRepository {
	return &tokenRepo{
		db:  db,
		log: logger,
	}
}

func (r *tokenRepo) Create(ctx context.Context, token *entities.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query,
		token.UserID, token.Token, token.ExpiresAt,
	)
	return err
}

func (r *tokenRepo) GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	rt := &entities.RefreshToken{}
	query := `
        SELECT id, user_id, token, expires_at, created_at 
        FROM refresh_tokens 
        WHERE token = $1`

	r.log.Info("getting refresh token",
		logger.NewField("token", token),
	)

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrTokenNotFound
	}

	if err != nil {
		r.log.Error("failed to get refresh token",
			logger.NewField("error", err),
			logger.NewField("token", token),
		)
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	return rt, nil
}

func (r *tokenRepo) Revoke(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		r.log.Error("failed to revoke token",
			logger.NewField("error", err),
			logger.NewField("token", token),
		)
	}
	return err
}

func (r *tokenRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *tokenRepo) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.log.Error("failed to revoke user tokens",
			logger.NewField("error", err),
			logger.NewField("user_id", userID),
		)
		return err
	}
	return nil
}
