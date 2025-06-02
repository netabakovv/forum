package main

// @title Gateway API
// @version 1.0
// @description API для форума с авторизацией, постами и комментариями

// @host localhost:8090
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

import (
	"time"

	"github.com/netabakovv/forum/back/gateway/internal/delivery/http"
	"github.com/netabakovv/forum/back/gateway/internal/handler"
	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := logger.NewStdLogger()

	authConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("не удалось подключиться к gRPC", logger.NewField("error", err))
	}
	defer authConn.Close()

	forumConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("не удалось подключиться к gRPC", logger.NewField("error", err))
	}
	defer forumConn.Close()

	authClient := pb.NewAuthServiceClient(authConn)
	forumClient := pb.NewForumServiceClient(forumConn)

	//forumClient := pb.NewForumServiceClient(conn)

	// Инициализация Gin
	router := gin.Default()

	// Разрешить CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // адрес фронта
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	handler := handler.NewHandler(forumClient, authClient, log)
	http.RegisterRoutes(router, handler)

	// Запуск gateway
	if err := router.Run(":8090"); err != nil {
		log.Fatal("не удалось запустить gateway", logger.NewField("error", err))
	}
}
