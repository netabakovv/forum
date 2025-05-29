package usecase_test

import (
	"back/auth_service/internal/entities"
	"back/auth_service/internal/usecase"
	"back/pkg/errors"
	"back/pkg/logger"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Моки ---

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepo) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepo) UpdateUsername(ctx context.Context, id int64, newName string) error {
	args := m.Called(ctx, id, newName)
	return args.Error(0)
}

// Только нужные методы, остальные можно добавить по мере необходимости

type mockTokenRepo struct {
	mock.Mock
}

func (m *mockTokenRepo) Create(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockTokenRepo) GetByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *mockTokenRepo) Revoke(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockTokenRepo) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockTokenRepo) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// ---

type mockTokenService struct {
	mock.Mock
}

func (m *mockTokenService) GenerateTokenPair(userID int64, username string, isAdmin bool) (*entities.TokenPair, error) {
	args := m.Called(userID, username, isAdmin)
	return args.Get(0).(*entities.TokenPair), args.Error(1)
}
func (m *mockTokenService) ValidateToken(tokenString string) (*entities.TokenClaims, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*entities.TokenClaims), args.Error(1)
}

// --- Тест ---

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	tokenService := new(mockTokenService)
	log := logger.NewStdLogger() // или с параметрами, если надо

	username := "new_user"
	password := "securepass"

	userRepo.On("GetByUsername", ctx, username).Return((*entities.User)(nil), errors.ErrUserNotFound)
	userRepo.On("Create", ctx, mock.Anything).Return(nil)
	tokenRepo.On("Create", ctx, mock.Anything).Return(nil)

	tokenPair := &entities.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	tokenService.On("GenerateTokenPair", mock.Anything, username, false).Return(tokenPair, nil)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, log)
	tokens, err := uc.Register(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, "access", tokens.AccessToken)
	assert.Equal(t, "refresh", tokens.RefreshToken)
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	tokenService := new(mockTokenService)
	log := logger.NewStdLogger() // или с параметрами, если надо

	username := "new_user"
	password := "securepass"

	userRepo.On("GetByUsername", ctx, username).Return((*entities.User)(nil), errors.ErrUserNotFound)
	userRepo.On("Create", ctx, mock.Anything).Return(nil)
	tokenRepo.On("Create", ctx, mock.Anything).Return(nil)

	tokenPair := &entities.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	tokenService.On("GenerateTokenPair", mock.Anything, username, false).Return(tokenPair, nil)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, log)
	tokens, err := uc.Register(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, "access", tokens.AccessToken)
	assert.Equal(t, "refresh", tokens.RefreshToken)
}
