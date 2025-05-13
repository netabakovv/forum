package usecase

import (
	"back/auth_service/internal/entities"
	"back/errors"
	"context"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthUsecase struct {
	userRepo UserRepository // Интерфейс из user-service
	jwtKey   string
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

func (uc *AuthUsecase) generateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(uc.jwtKey))
}
