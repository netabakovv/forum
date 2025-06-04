package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/netabakovv/forum/back/auth_service/internal/service"
	"github.com/netabakovv/forum/back/pkg/logger"
	"github.com/netabakovv/forum/back/pkg/logger/mocks"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidateTokenPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, mockLogger)

	mockLogger.
		EXPECT().
		Info("generating token pair",
			logger.NewField("user_id", int64(1)),
			logger.NewField("username", "testuser"),
		)

	tokenPair, err := ts.GenerateTokenPair(1, "testuser", true)
	require.NoError(t, err)
	require.NotEmpty(t, tokenPair.AccessToken)
	require.NotEmpty(t, tokenPair.RefreshToken)
	require.NotZero(t, tokenPair.ExpiresAt)

	claims, err := ts.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.True(t, claims.IsAdmin)
	assert.Equal(t, service.AccessTokenType, claims.TokenType)
	assert.Greater(t, claims.ExpiresAt, time.Now().Unix())
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, &mocks.MockLogger{})

	_, err := ts.ValidateToken("invalid.token.string")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestValidateToken_EmptyToken(t *testing.T) {
	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, &mocks.MockLogger{})

	_, err := ts.ValidateToken("")
	assert.Error(t, err)
	assert.EqualError(t, err, "empty token")
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	ts := service.NewTokenService("test-secret", time.Millisecond, time.Hour*24, mockLogger)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  int64(1),
		"username": "testuser",
		"is_admin": false,
		"type":     service.AccessTokenType,
		"exp":      time.Now().Add(-time.Minute).Unix(), // прошедшее время
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	_, err = ts.ValidateToken(tokenString)
	require.Error(t, err)
	assert.EqualError(t, err, "token expired")
}

func TestValidateToken_InvalidClaims(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(ctrl)
	invalidToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"invalid": "data",
	})

	signed, err := invalidToken.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	ts := service.NewTokenService("test-secret", time.Minute, time.Hour, mockLogger)
	_, err = ts.ValidateToken(signed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user_id claim")
}
