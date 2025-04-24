package main

import (
	"auth-service/internal/delivery/grpc"
	"auth-service/internal/repository/postgres"
	"auth-service/pkg/jwt"
	"auth-service/pkg/password"
	"log"
	"time"
)

func main() {
	// Инициализация репозитория
	db, err := postgres.NewPostgresDB("postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	userRepo := postgres.NewUserRepository(db)

	// Инициализация сервисов
	jwtService := jwt.NewJWTService("your_very_secret_key_here", 15*time.Minute, 720*time.Hour)
	hasher := password.NewBcryptHasher(12)

	// Создание gRPC сервера
	authServer := grpc.NewAuthServer(userRepo, jwtService, hasher)

	log.Println("Auth service starting on :50051...")
	if err := authServer.Run(":50051"); err != nil {
		log.Fatalf("failed to run auth server: %v", err)
	}
}
