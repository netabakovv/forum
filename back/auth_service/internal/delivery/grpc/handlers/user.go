package handlers

import (
	"auth_service/pkg/user_service/proto"
	"google.golang.org/grpc"
)

type UserServiceClient struct {
	client user.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(addr string) (*UserServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &UserServiceClient{
		client: user.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *UserServiceClient) GetUserByEmail(ctx context.Context, email string) (*user.UserResponse, error) {
	return c.client.GetUserByUsername(ctx, &user.GetUserByEmailRequest{Email: email})
}

// Закрытие соединения при необходимости
func (c *UserServiceClient) Close() error {
	return c.conn.Close()
}
