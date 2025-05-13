package grpc

import (
	"back/internal/usecase"
	"back/proto"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ForumServer struct {
	userUC    *usecase.UserUsecase
	postUC    *usecase.PostUsecase
	commentUC *usecase.CommentUsecase
}

// NewForumServer — конструктор (удобно для внедрения зависимостей)
func NewForumServer(
	userUC *usecase.UserUsecase,
	postUC *usecase.PostUsecase,
	commentUC *usecase.CommentUsecase,
) *ForumServer {
	return &ForumServer{
		userUC:    userUC,
		postUC:    postUC,
		commentUC: commentUC,
	}
}

// Реализация метода Register из proto-файла
func (s *ForumServer) Register(
	ctx context.Context,
	req *proto.RegisterRequest,
) (*pb.RegisterResponse, error) {
	// 1. Валидация (можно вынести в отдельный слой)
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// 2. Преобразование gRPC-запроса в доменную модель
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Пароль будет захэширован в usecase
	}

	// 3. Вызов бизнес-логики
	err := s.userUC.Register(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	// 4. Формирование gRPC-ответа
	return &pb.RegisterResponse{
		UserId: user.ID,
	}, nil
}

// Аналогично для CreatePost, AddComment...
