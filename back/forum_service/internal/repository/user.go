// repository/user.go
package repository

import (
	"back/internal/entities"
	"context"
	"database/sql"
	"errors"
	"log"
)

type UserRepository struct {
	db     *sql.DB
	logger log.Logger
}

func NewUserRepository(db sql.DB) *UserRepository {
	return &UserRepository{db: &db}
}

func (r *UserRepository) Create(user *entities.User) (bool, error) {
	_, err := r.db.Exec(
		"INSERT INTO users (username, password) VALUES ($1, $2)",
		user.Username, user.Password,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *UserRepository) GetById(id int64) (*entities.User, error) {
	_, err := r.db.Exec(
		"SELECT username FROM users WHERE users.id == $1",
		id,
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *UserRepository) GetByName(ctx context.Context, username string) (*entities.User, error) {
	user := &entities.User{}
	query := `SELECT id, username, email, password FROM users WHERE username = $1`

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrUserNotFound
	case err != nil:
		r.logger.Println("GetByName failed", "username", username, "error", err)
		return nil, err
	default:
		return user, nil
	}
}

// геттеры тут+работа с бд
