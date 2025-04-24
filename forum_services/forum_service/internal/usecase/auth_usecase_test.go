package usecase

import (
	"context"
	"forum-service/internal/entity"
	"forum-service/pkg/grpc/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthGRPCUseCase_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := auth.NewMockAuthServiceClient(ctrl)
	uc := NewAuthGRPCUseCase(mockClient)

	t.Run("valid token", func(t *testing.T) {
		mockClient.EXPECT().ValidateToken(
			gomock.Any(),
			&auth.TokenRequest{Token: "valid-token"},
		).Return(&auth.TokenResponse{
			Valid:  true,
			UserId: 1,
			Role:   "user",
		}, nil)

		user, err := uc.ValidateToken("valid-token")

		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "user", user.Role)
	})

	t.Run("invalid token", func(t *testing.T) {
		mockClient.EXPECT().ValidateToken(
			gomock.Any(),
			&auth.TokenRequest{Token: "invalid-token"},
		).Return(&auth.TokenResponse{
			Valid: false,
		}, nil)

		_, err := uc.ValidateToken("invalid-token")

		assert.ErrorIs(t, err, entity.ErrInvalidToken)
	})
}
