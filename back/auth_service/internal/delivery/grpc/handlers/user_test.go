package handlers_test

import (
	"github.com/netabakovv/forum/back/auth_service/internal/delivery/grpc/handlers/mocks"
	//	logmock "back/pkg/logger/mocks"
	"context"
	"errors"
	"testing"

	"github.com/netabakovv/forum/back/proto"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClientInterface(ctrl)

	resp := &proto.UserProfileResponse{
		UserId:   1,
		Username: "test_user",
	}

	mockClient.EXPECT().
		GetUserByID(gomock.Any(), int64(1)).
		Return(resp, nil)

	user, err := mockClient.GetUserByID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), user.UserId)
	require.Equal(t, "test_user", user.Username)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClientInterface(ctrl)

	mockClient.EXPECT().
		GetUserByID(gomock.Any(), int64(2)).
		Return(nil, status.Error(codes.NotFound, "user not found"))

	_, err := mockClient.GetUserByID(context.Background(), 2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user not found")
}

func TestGetUserByID_OtherError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClientInterface(ctrl)

	mockClient.EXPECT().
		GetUserByID(gomock.Any(), int64(3)).
		Return(nil, errors.New("connection error"))

	_, err := mockClient.GetUserByID(context.Background(), 3)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connection error")
}

func TestClose_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockUserServiceClientInterface(ctrl)

	mockClient.EXPECT().
		Close().
		Return(nil)

	err := mockClient.Close()
	require.NoError(t, err)
}
