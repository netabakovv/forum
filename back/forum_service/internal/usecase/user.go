// usecase/user/user.go
package usecase

import (
	"back/internal/entities"
	"back/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"os/user"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (s *UserUsecase) Register(username, password string) (bool, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	user := &entities.User{
		Username: username,
		Password: string(hashedPassword),
	}

	return s.repo.Create(user)
}

func (s *UserUsecase) Login(username, password string) (*entities.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}
