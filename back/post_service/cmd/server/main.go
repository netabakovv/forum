package server

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
)

func main() {
	// URL подключения: postgres://user:password@host:port/database
	connStr := "postgres://user:password@localhost:5432/forum_db?sslmode=disable"

	// Подключение к базе
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer conn.Close(context.Background()) // Закрыть подключение при выходе

	// Проверка подключения
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("Ошибка ping:", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL!")
}
