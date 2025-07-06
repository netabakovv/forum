package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/netabakovv/forum/back/auth_service/internal/delivery/grpc/handlers"
	"github.com/netabakovv/forum/back/pkg/logger/mocks"
	pb "github.com/netabakovv/forum/back/proto"
	pbMocks "github.com/netabakovv/forum/back/proto/mocks"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/require"
)

func TestGetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockGrpcClient := pbMocks.NewMockAuthServiceClient(ctrl)

	// Ожидаем вызов логгера Info
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)

	// Ожидаем вызов GetUserByID на grpc клиенте и возвращаем успешный ответ
	mockGrpcClient.EXPECT().
		GetUserByID(gomock.Any(), &pb.GetUserRequest{UserId: 1}).
		Return(&pb.UserProfileResponse{UserId: 1, Username: "testuser"}, nil).
		Times(1)

	client := &handlers.UserServiceClient{
		Client: mockGrpcClient,
		Logger: mockLogger,
	}

	resp, err := client.GetUserByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.UserId)
	require.Equal(t, "testuser", resp.Username)
}

func TestGetUserByID_NilContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockGrpcClient := pbMocks.NewMockAuthServiceClient(ctrl)

	client := &handlers.UserServiceClient{
		Client: mockGrpcClient,
		Logger: mockLogger,
	}

	resp, err := client.GetUserByID(nil, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "контекст не может быть nil")
	require.Nil(t, resp)
}

func TestGetUserByID_ZeroUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockGrpcClient := pbMocks.NewMockAuthServiceClient(ctrl)

	client := &handlers.UserServiceClient{
		Client: mockGrpcClient,
		Logger: mockLogger,
	}

	resp, err := client.GetUserByID(context.Background(), 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ID пользователя не может быть пустым")
	require.Nil(t, resp)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockGrpcClient := pbMocks.NewMockAuthServiceClient(ctrl)

	// Ожидаем вызов логгера Info и Warn
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).Times(1)

	mockGrpcClient.EXPECT().
		GetUserByID(gomock.Any(), &pb.GetUserRequest{UserId: 2}).
		Return(nil, status.Error(codes.NotFound, "not found")).
		Times(1)

	client := &handlers.UserServiceClient{
		Client: mockGrpcClient,
		Logger: mockLogger,
	}

	resp, err := client.GetUserByID(context.Background(), 2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "пользователь не найден")
	require.Nil(t, resp)
}

func TestGetUserByID_OtherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockGrpcClient := pbMocks.NewMockAuthServiceClient(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).Times(1)
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	mockGrpcClient.EXPECT().
		GetUserByID(gomock.Any(), &pb.GetUserRequest{UserId: 3}).
		Return(nil, errors.New("connection error")).
		Times(1)

	client := &handlers.UserServiceClient{
		Client: mockGrpcClient,
		Logger: mockLogger,
	}

	resp, err := client.GetUserByID(context.Background(), 3)
	require.Error(t, err)
	require.Contains(t, err.Error(), "не удалось получить профиль пользователя")
	require.Nil(t, resp)
}

func TestClose_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)

	client := &handlers.UserServiceClient{
		Logger: mockLogger,
		// c.conn == nil, значит просто вернет nil
	}

	err := client.Close()
	require.NoError(t, err)
}
