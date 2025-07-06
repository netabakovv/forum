package handlers

import (
	"context"
	"fmt"
	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceClientInterface interface {
	GetUserByID(ctx context.Context, userID int64) (*pb.UserProfileResponse, error)
	Close() error
}

// UserServiceClient предоставляет клиент для взаимодействия с сервисом пользователей
type UserServiceClient struct {
	Client pb.AuthServiceClient
	Conn   *grpc.ClientConn
	Logger logger.Logger
}

// GetUserByID получает профиль пользователя по ID
func (c *UserServiceClient) GetUserByID(ctx context.Context, userID int64) (*pb.UserProfileResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("контекст не может быть nil")
	}
	if userID == 0 {
		return nil, fmt.Errorf("ID пользователя не может быть пустым")
	}

	c.Logger.Info("получение пользователя по ID",
		logger.NewField("user_id", userID),
	)

	resp, err := c.Client.GetUserByID(ctx, &pb.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			c.Logger.Warn("пользователь не найден",
				logger.NewField("user_id", userID),
			)
			return nil, fmt.Errorf("пользователь не найден: %w", err)
		default:
			c.Logger.Error("не удалось получить профиль пользователя",
				logger.NewField("error", err),
				logger.NewField("user_id", userID),
			)
			return nil, fmt.Errorf("не удалось получить профиль пользователя: %w", err)
		}
	}

	return resp, nil
}

// Close закрывает соединение с сервером
func (c *UserServiceClient) Close() error {
	if c.Conn == nil {
		return nil
	}

	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть соединение: %w", err)
	}
	return nil
}
