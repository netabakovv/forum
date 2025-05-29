package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func NewPostgresConnection(connStr string) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), connStr)
}
