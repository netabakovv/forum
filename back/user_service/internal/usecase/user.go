// usecase/user/user.go
package usecase

import (
	"back/user_service/internal/entities"
	"back/user_service/internal/repository"
	"context"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUsecase struct {
	repo   repository.UserRepository
	jwtKey string
}

// Регистрация
func (uc *UserUsecase) Register(ctx context.Context, username, password string) (*entities.SafeUser, error) {
	// Хеширование пароля
	req := &entities.User{}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	req.Username = username
	req.Password = string(hashedPass)
	req.CreatedAt = time.Now()

	// Сохранение в БД
	if err := uc.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return req.ToSafe(), err

}

// Получение пользователя
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*entities.SafeUser, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToSafe(), nil
}

// Аналогично для UpdateUser, DeleteUser...
