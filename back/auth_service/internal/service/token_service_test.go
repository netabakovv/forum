package service_test

// import (
// 	"back/auth_service/internal/service"
// 	"back/pkg/logger/mocks"
// 	"testing"
// 	"time"

// 	"github.com/golang-jwt/jwt"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestGenerateAndValidateTokenPair(t *testing.T) {
// 	mockLogger := &mocks.MockLogger{}
// 	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, mockLogger)

// 	tokenPair, err := ts.GenerateTokenPair(1, "testuser", true)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, tokenPair.AccessToken)
// 	require.NotEmpty(t, tokenPair.RefreshToken)
// 	require.NotZero(t, tokenPair.ExpiresAt)

// 	claims, err := ts.ValidateToken(tokenPair.AccessToken)
// 	require.NoError(t, err)
// 	assert.Equal(t, int64(1), claims.UserID)
// 	assert.Equal(t, "testuser", claims.Username)
// 	assert.True(t, claims.IsAdmin)
// 	assert.Equal(t, service.AccessTokenType, claims.TokenType)
// 	assert.Greater(t, claims.ExpiresAt, time.Now().Unix())
// }

// func TestValidateToken_InvalidToken(t *testing.T) {
// 	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, &mocks.MockLogger{})

// 	_, err := ts.ValidateToken("invalid.token.string")
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "failed to parse token")
// }

// func TestValidateToken_EmptyToken(t *testing.T) {
// 	ts := service.NewTokenService("test-secret", time.Minute*15, time.Hour*24, &mocks.MockLogger{})

// 	_, err := ts.ValidateToken("")
// 	assert.Error(t, err)
// 	assert.EqualError(t, err, "empty token")
// }

// func TestValidateToken_ExpiredToken(t *testing.T) {
// 	ts := service.NewTokenService("test-secret", time.Millisecond, time.Hour*24, &mocks.MockLogger{})

// 	token, err := ts.GenerateTokenPair(1, "testuser", false)
// 	require.NoError(t, err)

// 	time.Sleep(time.Millisecond * 2) // wait for token to expire

// 	_, err = ts.ValidateToken(token.AccessToken)
// 	require.Error(t, err)
// 	assert.EqualError(t, err, "token expired")
// }

// func TestValidateToken_InvalidClaims(t *testing.T) {
// 	invalidToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"invalid": "data",
// 	})

// 	signed, err := invalidToken.SignedString([]byte("test-secret"))
// 	require.NoError(t, err)

// 	ts := service.NewTokenService("test-secret", time.Minute, time.Hour, &mocks.MockLogger{})
// 	_, err = ts.ValidateToken(signed)
// 	require.Error(t, err)
// 	assert.Contains(t, err.Error(), "invalid user_id claim")
// }
