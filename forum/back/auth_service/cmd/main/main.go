package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"back/auth_service/internal/delivery/grpc"
	"back/auth_service/internal/repository"
	"back/auth_service/internal/service"
	"back/auth_service/internal/usecase"
	"back/pkg/logger"
	pb "back/proto"

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
	viper.SetConfigFile("./config.yaml")
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
	return db
}
