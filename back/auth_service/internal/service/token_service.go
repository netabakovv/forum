package service

import (
	"fmt"
	"github.com/netabakovv/forum/back/pkg/errors"
	"time"

	"github.com/netabakovv/forum/back/auth_service/internal/entities"
	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/golang-jwt/jwt"
)

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"
)

type TokenService struct {
	secretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	logger          logger.Logger
}

type TokenServiceInterface interface {
	GenerateTokenPair(userID int64, username string, isAdmin bool) (*entities.TokenPair, error)
	ValidateToken(tokenString string) (*entities.TokenClaims, error)
}

func NewTokenService(secretKey string, accessTTL, refreshTTL time.Duration, logger logger.Logger) *TokenService {
	return &TokenService{
		secretKey:       secretKey,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		logger:          logger,
	}
}

func (s *TokenService) GenerateTokenPair(userID int64, username string, isAdmin bool) (*entities.TokenPair, error) {
	s.logger.Info("generating token pair",
		logger.NewField("user_id", userID),
		logger.NewField("username", username),
	)

	// Генерируем Access Token
	accessToken, err := s.generateToken(userID, username, isAdmin, s.AccessTokenTTL, AccessTokenType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Генерируем Refresh Token
	refreshToken, err := s.generateToken(userID, username, isAdmin, s.RefreshTokenTTL, RefreshTokenType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.AccessTokenTTL)

	return &entities.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *TokenService) ValidateToken(tokenString string) (*entities.TokenClaims, error) {
	if tokenString == "" {
		return nil, errors.ErrTokenInvalid
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrTokenInvalid
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.ErrTokenExpired
			}
		}
		return nil, errors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.ErrTokenInvalid
	}

	userID, ok := claims["user_id"].(float64)
	username, ok1 := claims["username"].(string)
	isAdmin, ok2 := claims["is_admin"].(bool)
	tokenType, ok3 := claims["type"].(string)
	exp, ok4 := claims["exp"].(float64)

	if !ok || !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, errors.ErrTokenInvalid
	}

	return &entities.TokenClaims{
		UserID:    int64(userID),
		Username:  username,
		IsAdmin:   isAdmin,
		ExpiresAt: int64(exp),
		TokenType: tokenType,
	}, nil
}

func (s *TokenService) generateToken(userID int64, username string, isAdmin bool, expiration time.Duration, tokenType string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"is_admin": isAdmin,
		"type":     tokenType,
		"iat":      now.Unix(),
		"exp":      now.Add(expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
