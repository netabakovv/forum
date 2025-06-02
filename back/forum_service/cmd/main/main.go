package main

import (
	serv "back/forum_service/internal/delivery/grpc"
	"back/forum_service/internal/delivery/ws"
	"back/forum_service/internal/repository"
	"back/forum_service/internal/usecase"
	"back/pkg/logger"
	pb "back/proto"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.NewStdLogger()

	if err := initConfig(); err != nil {
		log.Fatal("ошибка инициализации конфига",
			logger.NewField("error", err))
	}

	// Инициализация репозиториев
	db := initDB(log)
	defer db.Close()

	// Подключение к auth service
	authConn, err := grpc.Dial(
		fmt.Sprintf("localhost:%s", viper.GetString("auth_service.port")),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("ошибка подключения к auth service", logger.NewField("error", err))
	}
	defer authConn.Close()

	authClient := pb.NewAuthServiceClient(authConn)

	// Репозитории
	postRepo := repository.NewPostRepository(db, log)
	commentRepo := repository.NewCommentRepository(db, log)
	chatRepo := repository.NewChatRepository(db, log)

	// Use cases
	postUC := usecase.NewPostUsecase(postRepo, log)
	commentUC := usecase.NewCommentUsecase(commentRepo, log)
	chatUC := usecase.NewChatUsecase(chatRepo, log, &pb.ChatConfig{
		MessageLifetimeMinutes: 1,
		MaxMessageLength:       1000,
		OnlyAuthenticated:      true})
	cleanup := usecase.NewCleanupService(chatUC, log)
	cleanup.Start(viper.GetDuration("chat.cleanup_interval"), viper.GetDuration("chat.message_lifetime"))
	defer cleanup.Stop()

	// gRPC сервер
	grpcServer := grpc.NewServer()

	// Форум сервер
	forumServer := serv.NewForumServer(authClient, postUC, commentUC, chatUC)
	pb.RegisterForumServiceServer(grpcServer, forumServer)

	// WebSocket чат

	chatHandler := ws.NewChatHandler(chatUC, log, &pb.ChatConfig{
		MessageLifetimeMinutes: 1,
		MaxMessageLength:       1000,
		OnlyAuthenticated:      true,
	}, authClient)
	http.HandleFunc("/ws/chat", chatHandler.HandleWebSocket)

	// Запуск серверов
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ошибка WebSocket сервера", logger.NewField("error", err))
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("ошибка запуска gRPC сервера", logger.NewField("error", err))
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("ошибка работы gRPC сервера", logger.NewField("error", err))
	}
}

func initConfig() error {
	viper.SetConfigFile("/app/config.yaml")
	return viper.ReadInConfig()
}

func initDB(log logger.Logger) *sql.DB {
	// Используем строку подключения из конфига
	dbURL := viper.GetString("forumPath")
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
		"file:///app/back/migrations/forum",
		"postgres", driver)
	if err != nil {
		log.Fatal("ошибка при создании миграций",
			logger.NewField("error", err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("ошибка при выполнении миграций",
			logger.NewField("error", err))
	}

	return db
}
