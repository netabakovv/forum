package repository

import (
	"database/sql"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db sql.DB) *PostRepository {
	return &PostRepository{db: &db}
}
