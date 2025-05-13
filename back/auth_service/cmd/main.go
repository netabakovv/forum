package main

import (
	"back/internal/delivery"
	"back/internal/repository"
	"back/internal/usecase"
	"back/proto"
	"database/sql"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	// Подключение к Postgres
	db, err := sql.Open("postgres", "user=postgres password=1 dbname=auth sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Инициализация слоев
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	userUC := usecase.NewUserUsecase(userRepo)
	postUC := usecase.NewPostUsecase(postRepo)
	commentUC := usecase.NewCommentUsecase(commentRepo)

	grpcServer := grpc.NewServer()
	forumServer := grpc.NewForumServer(userUC, postUC, commentUC)
	pb.RegisterForumServer(grpcServer, forumServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
