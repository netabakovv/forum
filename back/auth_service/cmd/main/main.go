package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netabakovv/forum/back/auth_service/internal/delivery/grpc"
	"github.com/netabakovv/forum/back/auth_service/internal/repository"
	"github.com/netabakovv/forum/back/auth_service/internal/service"
	"github.com/netabakovv/forum/back/auth_service/internal/usecase"
	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log := logger.NewStdLogger()

	// Инициализация конфигурации
	if err := initConfig(); err != nil {
		log.Fatal("ошибка инициализации конфига",
			logger.NewField("error", err))
	}

	// Инициализация репозиториев
	db := initDB(log)
	defer db.Close()

	userRepo := repository.NewUserRepository(db, log)
	tokenRepo := repository.NewTokenRepository(db, log)

	log.Info("репозитории инициализированы успешно")

	// Инициализация сервисов с конфигом
	tokenService := service.NewTokenService(
		viper.GetString("auth.jwt_secret"),
		viper.GetDuration("auth.access_token_ttl"),
		viper.GetDuration("auth.refresh_token_ttl"),
		log,
	)

	log.Info("сервисы инициализированы успешно")

	// Инициализация usecase
	authUC := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, log)
	log.Info("use-case'ы инициализированы успешно")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Запускаем горутину для периодического вызова функции
	go func() {
		for {
			select {
			case <-ticker.C:
				authUC.DeleteExpired(context.Background())
			}
		}
	}()

	// Инициализация gRPC сервера
	server := grpc.NewAuthServer(authUC, tokenService, log)

	s := ggrpc.NewServer()
	pb.RegisterAuthServiceServer(s, server)

	// Для работы рефлексии (опционально)
	reflection.Register(s)

	// Запуск gRPC сервера
	port := viper.GetString("auth_service.port")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("не удалось запустить прослушивание порта",
			logger.NewField("error", err))
	}

	// Graceful shutdown
	go func() {
		log.Info("запуск gRPC сервера на порту",
			logger.NewField("port", port))
		if err := s.Serve(listener); err != nil {
			log.Fatal("ошибка при запуске сервера",
				logger.NewField("error", err))
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("начало graceful shutdown")
	s.GracefulStop()
	log.Info("сервер остановлен")
}

func initConfig() error {
	viper.SetConfigFile("/app/config.yaml")
	return viper.ReadInConfig()
}

func initDB(log logger.Logger) *sql.DB {
	// Используем строку подключения из конфига
	dbURL := viper.GetString("authPath")
	log.Info(dbURL)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("не удалось подключиться к базе данных",
			logger.NewField("error", err))
	}
	if err := db.Ping(); err != nil {
		log.Fatal("не удалось проверить соединение с базой данных",
			logger.NewField("error", err))
	}
	log.Info("успешное подключение к базе данных")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("ошибка при инициализации драйвера для базы данных",
			logger.NewField("error", err))
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/back/migrations/auth",
		"postgres", driver)
	if err != nil {
		log.Fatal("ошибка при создании миграций",
			logger.NewField("error", err))
	}

	if err := m.Up(); err != nil {
		log.Info("ошибка при выполнении миграций",
			logger.NewField("error", err))
	}

	log.Info("успешное выполнение миграций")
	return db
}
