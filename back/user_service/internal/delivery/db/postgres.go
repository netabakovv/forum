package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
	"log"
)

func SimpleConnection() {
	// Формат: postgres://username:password@host:port/database?sslmode=disable
	connStr := "postgres://user:password@localhost:5555/forum_db?sslmode=disable"

	// Установка соединения
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	defer conn.Close(context.Background())

	// Проверка соединения
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatalf("Unable to ping: %v\n", err)
	}

	fmt.Println("Successfully connected!")
}
