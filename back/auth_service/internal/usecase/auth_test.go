package usecase_test

import (
	"context"
	"fmt"
	"github.com/netabakovv/forum/back/pkg/errors"
	logger2 "github.com/netabakovv/forum/back/pkg/logger"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/netabakovv/forum/back/auth_service/internal/entities"
	mock_repo "github.com/netabakovv/forum/back/auth_service/internal/repository/mocks"
	mock_service "github.com/netabakovv/forum/back/auth_service/internal/service/mocks"
	"github.com/netabakovv/forum/back/auth_service/internal/usecase"
	mock_logger "github.com/netabakovv/forum/back/pkg/logger/mocks"
	"github.com/stretchr/testify/require"
)

func TestAuthUsecase_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hashedPassword, _ := usecase.HashPassword(password)

	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	authUC := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)

	userRepo.EXPECT().
		GetByUsername(ctx, username).
		Return(nil, errors.ErrUserNotFound)

	userRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, u *entities.User) error {
			u.ID = 42 // задаём ID
			u.CreatedAt = time.Now()
			return nil
		})

	expectedTokens := &entities.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	tokenService.EXPECT().
		GenerateTokenPair(int64(42), username, false).
		Return(expectedTokens, nil)

	tokenRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(nil)

	logger.EXPECT().Info("attempting registration", gomock.Any())
	logger.EXPECT().Info("registration successful", gomock.Any(), gomock.Any())

	tokens, err := authUC.Register(ctx, username, hashedPassword)
	require.NoError(t, err)
	require.Equal(t, expectedTokens, tokens)
}

func TestAuthUsecase_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)

	username := "testuser"
	password := "testpass"

	t.Run("user already exists", func(t *testing.T) {
		logger.EXPECT().Info("attempting registration", logger2.NewField("username", username))
		userRepo.EXPECT().GetByUsername(ctx, username).Return(&entities.User{}, nil)
		logger.EXPECT().Warn("user already exists", logger2.NewField("username", username))

		tokens, err := uc.Register(ctx, username, password)
		require.Nil(t, tokens)
		require.ErrorIs(t, err, errors.ErrDuplicateUsername)
	})

	t.Run("failed to create user", func(t *testing.T) {
		logger.EXPECT().Info("attempting registration", logger2.NewField("username", username))
		userRepo.EXPECT().GetByUsername(ctx, username).Return(nil, errors.ErrNotFound)
		// реальный HashPassword
		// не подменяем, он рабочий
		userRepo.EXPECT().Create(ctx, gomock.Any()).Return(fmt.Errorf("create error"))
		logger.EXPECT().Error("failed to create user", gomock.Any(), logger2.NewField("username", username))

		tokens, err := uc.Register(ctx, username, password)
		require.Nil(t, tokens)
		require.EqualError(t, err, "create error")
	})

	t.Run("failed to generate tokens", func(t *testing.T) {
		logger.EXPECT().Info("attempting registration", logger2.NewField("username", username))
		userRepo.EXPECT().GetByUsername(ctx, username).Return(nil, errors.ErrNotFound)
		userRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		tokenService.EXPECT().GenerateTokenPair(gomock.Any(), username, false).Return(nil, fmt.Errorf("token gen error"))
		logger.EXPECT().Error("failed to generate tokens", gomock.Any(), logger2.NewField("user_id", int64(0)))

		tokens, err := uc.Register(ctx, username, password)
		require.Nil(t, tokens)
		require.EqualError(t, err, "token gen error")
	})

	t.Run("failed to save refresh token", func(t *testing.T) {
		logger.EXPECT().Info("attempting registration", logger2.NewField("username", username))
		userRepo.EXPECT().GetByUsername(ctx, username).Return(nil, errors.ErrNotFound)
		userRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		tokenPair := &entities.TokenPair{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}

		tokenService.EXPECT().GenerateTokenPair(gomock.Any(), username, false).Return(tokenPair, nil)

		tokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(fmt.Errorf("save token error"))
		logger.EXPECT().Error("failed to save refresh token", gomock.Any(), logger2.NewField("user_id", int64(0)))

		tokens, err := uc.Register(ctx, username, password)
		require.Nil(t, tokens)
		require.EqualError(t, err, "save token error")
	})
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	hash, _ := usecase.HashPassword(password)

	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)

	user := &entities.User{
		ID:           1,
		Username:     username,
		PasswordHash: hash,
		IsAdmin:      false,
	}

	userRepo.EXPECT().GetByUsername(ctx, username).Return(user, nil)
	tokenService.EXPECT().GenerateTokenPair(user.ID, username, false).Return(&entities.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}, nil)
	tokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	logger.EXPECT().Info("attempting login", gomock.Any())
	logger.EXPECT().Info("login successful", gomock.Any(), gomock.Any())

	tokens, gotUser, err := uc.Login(ctx, username, password)
	require.NoError(t, err)
	require.Equal(t, "access", tokens.AccessToken)
	require.Equal(t, user.ID, gotUser.ID)
}

func TestAuthUsecase_Login_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)

	username := "testuser"
	password := "testpass"

	t.Run("user not found", func(t *testing.T) {
		logger.EXPECT().Info("attempting login", logger2.NewField("username", username))
		userRepo.EXPECT().GetByUsername(ctx, username).Return(nil, errors.ErrNotFound)
		logger.EXPECT().Warn("failed to get user",
			logger2.NewField("error", errors.ErrNotFound),
			logger2.NewField("username", username),
		)

		tokens, _, err := uc.Login(ctx, username, password)
		require.Nil(t, tokens)
		require.ErrorIs(t, err, errors.ErrNotFound)
	})

	t.Run("invalid password", func(t *testing.T) {
		logger.EXPECT().Info("attempting login", logger2.NewField("username", username))

		user := &entities.User{
			ID:           1,
			Username:     username,
			PasswordHash: "$2a$10$invalidhash", // заведомо плохой bcrypt-хеш
			IsAdmin:      false,
		}
		userRepo.EXPECT().GetByUsername(ctx, username).Return(user, nil)
		logger.EXPECT().Warn("invalid password attempt", logger2.NewField("username", username))

		tokens, _, err := uc.Login(ctx, username, "wrongpass")
		require.Nil(t, tokens)
		require.ErrorIs(t, err, errors.ErrWrongPassword)
	})

	t.Run("failed to generate tokens", func(t *testing.T) {
		logger.EXPECT().Info("attempting login", logger2.NewField("username", username))

		hash, _ := usecase.HashPassword(password)
		user := &entities.User{
			ID:           1,
			Username:     username,
			PasswordHash: hash,
			IsAdmin:      false,
		}

		userRepo.EXPECT().GetByUsername(ctx, username).Return(user, nil)
		tokenService.EXPECT().GenerateTokenPair(user.ID, username, false).Return(nil, fmt.Errorf("token gen error"))
		logger.EXPECT().Error("failed to generate tokens", logger2.NewField("error", fmt.Errorf("token gen error")), logger2.NewField("user_id", int64(1)))

		tokens, _, err := uc.Login(ctx, username, password)
		require.Nil(t, tokens)
		require.EqualError(t, err, "token gen error")
	})

	t.Run("failed to save refresh token", func(t *testing.T) {
		logger.EXPECT().Info("attempting login", logger2.NewField("username", username))

		hash, _ := usecase.HashPassword(password)
		user := &entities.User{
			ID:           1,
			Username:     username,
			PasswordHash: hash,
			IsAdmin:      false,
		}

		userRepo.EXPECT().GetByUsername(ctx, username).Return(user, nil)

		tokenPair := &entities.TokenPair{
			AccessToken:  "access",
			RefreshToken: "refresh",
		}
		tokenService.EXPECT().GenerateTokenPair(user.ID, username, false).Return(tokenPair, nil)
		tokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(fmt.Errorf("save token error"))
		logger.EXPECT().Error("failed to save refresh token", gomock.Any(), logger2.NewField("user_id", int64(1)))

		tokens, _, err := uc.Login(ctx, username, password)
		require.Nil(t, tokens)
		require.EqualError(t, err, "save token error")
	})
}

func TestAuthUsecase_IsAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)

	userRepo.EXPECT().GetByID(ctx, int64(1)).Return(&entities.User{
		ID:      1,
		IsAdmin: true,
	}, nil)

	isAdmin, err := uc.IsAdmin(ctx, 1)
	require.NoError(t, err)
	require.True(t, isAdmin)
}

func TestAuthUsecase_RevokeTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := logger2.NewStdLogger()

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)
	userID := int64(1)

	logger.Info("attempting to revoke all user tokens",
		logger2.NewField("user_id", userID),
	)

	logger.Info("successfully revoked all user tokens",
		logger2.NewField("user_id", userID),
	)

	tokenRepo.EXPECT().RevokeAllUserTokens(ctx, userID).Return(nil)

	err := uc.RevokeTokens(ctx, userID)

	require.NoError(t, err)
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)
	token := "refresh"

	tokenRepo.EXPECT().Revoke(ctx, token).Return(nil)

	err := uc.Logout(ctx, token)
	require.NoError(t, err)
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userRepo := mock_repo.NewMockUserRepository(ctrl)
	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	logger := mock_logger.NewMockLogger(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, logger)
	token := "access"

	expectedClaims := &entities.TokenClaims{UserID: 1, Username: "test"}
	tokenService.EXPECT().ValidateToken(token).Return(expectedClaims, nil)

	claims, err := uc.ValidateToken(ctx, token)
	require.NoError(t, err)
	require.Equal(t, expectedClaims.UserID, claims.UserID)
}

func TestAuthUsecase_RefreshTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	token := "refresh"

	userID := int64(1)
	username := "user"
	isAdmin := false

	tokenClaims := &entities.TokenClaims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
	}

	tokenRepo := mock_repo.NewMockTokenRepository(ctrl)
	tokenService := mock_service.NewMockTokenServiceInterface(ctrl)
	userRepo := mock_repo.NewMockUserRepository(ctrl)

	uc := usecase.NewAuthUsecase(userRepo, tokenRepo, tokenService, nil)

	tokenService.EXPECT().ValidateToken(token).Return(tokenClaims, nil)
	tokenRepo.EXPECT().GetByToken(ctx, token).Return(&entities.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	tokenService.EXPECT().GenerateTokenPair(userID, username, isAdmin).Return(&entities.TokenPair{
		AccessToken:  "new_access",
		RefreshToken: "new_refresh",
	}, nil)
	tokenRepo.EXPECT().Revoke(ctx, token).Return(nil)
	tokenRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	tokens, err := uc.RefreshTokens(ctx, token)
	require.NoError(t, err)
	require.Equal(t, "new_access", tokens.AccessToken)
}
