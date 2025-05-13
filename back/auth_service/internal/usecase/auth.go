package usecase

import (
	"back/auth_service/internal/entities"
	"back/auth_service/internal/errors"
	"back/auth_service/internal/repository"
	"back/proto"
	"context"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthUsecase struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	userService user.UserServiceClient // GRPC клиент к UserService
}

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.ErrWrongPassword
	}

	return uc.generateJWT(user.ID)
}

func (uc *AuthUsecase) Logout(ctx context.Context) {

}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*entities.AuthResponse, error) {
	// Проверяем refresh token в БД
	token, err := uc.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil || token.Revoked || time.Now().After(token.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	// Отзываем старый токен
	if err := uc.tokenRepo.Revoke(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Генерируем новую пару токенов
	return uc.generateTokens(token.UserID)
}

func (uc *AuthUsecase) generateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(uc.jwtKey))
}
