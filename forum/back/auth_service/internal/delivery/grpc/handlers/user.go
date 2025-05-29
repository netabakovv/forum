package handlers

import (
	"back/pkg/logger"
	pb "back/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// UserServiceClient предоставляет клиент для взаимодействия с сервисом пользователей
type UserServiceClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
	logger logger.Logger
}

// NewUserServiceClient создает новый экземпляр клиента сервиса пользователей
func NewUserServiceClient(addr string, logger logger.Logger) (*UserServiceClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("адрес сервера не может быть пустым")
	}
	if logger == nil {
		return nil, fmt.Errorf("логгер не может быть nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}

	return &UserServiceClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

// GetUserByID получает профиль пользователя по ID
func (c *UserServiceClient) GetUserByID(ctx context.Context, userID int64) (*pb.UserProfileResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("контекст не может быть nil")
	}
	if userID == 0 {
		return nil, fmt.Errorf("ID пользователя не может быть пустым")
	}

	c.logger.Info("получение пользователя по ID",
		logger.NewField("user_id", userID),
	)

	resp, err := c.client.GetUserByID(ctx, &pb.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			c.logger.Warn("пользователь не найден",
				logger.NewField("user_id", userID),
			)
			return nil, fmt.Errorf("пользователь не найден: %w", err)
		default:
			c.logger.Error("не удалось получить профиль пользователя",
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
	if c.conn == nil {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		c.logger.Error("не удалось закрыть соединение",
			logger.NewField("error", err),
		)
		return fmt.Errorf("не удалось закрыть соединение: %w", err)
	}
	return nil
}
