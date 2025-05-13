// repository/user.go
package repository

import (
	"back/user_service/internal/entities"
	"back/user_service/internal/errors"
	"context"
	"database/sql"
	"log"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByCreatedAt(ctx context.Context, created string) (*entities.User, error)
	UpdateUsername(ctx context.Context, id int64, newName string) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

// Реализация для PostgreSQL
type userRepo struct {
	db  *sql.DB
	log log.Logger
}

func (r *userRepo) Create(ctx context.Context, user *entities.User) (bool, error) {
	query := `
        INSERT INTO users (username, password, created_at)
        VALUES ($1, $2, $3)
        RETURNING id`
	_, err := r.db.QueryContext(ctx, query,
		user.Username, user.Password, user.CreatedAt,
	)
	if err != nil {
		return false, errors.ErrDuplicateUsername
	}
	return true, err
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	user := &entities.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	user := &entities.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE username = $1`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) GetByCreatedAt(ctx context.Context, created string) (*entities.User, error) {
	user := &entities.User{}
	query := `SELECT id, username, email, password, created_at FROM users WHERE created_at = $1`
	err := r.db.QueryRowContext(ctx, query, created).Scan(
		&user.ID, &user.Username, &user.Password, &user.CreatedAt,
	)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) UpdateUsername(ctx context.Context, id int64, newName string) (bool, error) {
	query := `UPDATE users SET username = $1 WHERE id = $2`
	_, err := r.db.QueryContext(ctx, query, newName, id)
	if err != nil {
		return false, errors.ErrDuplicateUsername
	}
	return true, err
}

func (r *userRepo) Delete(ctx context.Context, id int64) (bool, error) {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return false, errors.ErrDeleteFailed
	}
	return true, err
}

// Аналогично для GetByEmail, Update, Delete...
