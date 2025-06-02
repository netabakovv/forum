package usecase

import (
	"context"
	"time"

	"github.com/netabakovv/forum/back/pkg/errors"
	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/netabakovv/forum/back/auth_service/internal/entities"
	"github.com/netabakovv/forum/back/auth_service/internal/repository"
	"github.com/netabakovv/forum/back/auth_service/internal/service"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseInterface interface {
	Register(ctx context.Context, username, password string) (*entities.TokenPair, error)
	Login(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*entities.TokenPair, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	RevokeTokens(ctx context.Context, userID int64) error
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error)
}

type LoginResponse struct {
	Tokens *entities.TokenPair
	User   *entities.User
}

type AuthUsecase struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	tokenService    service.TokenServiceInterface
	logger          logger.Logger
	RefreshTokenTTL time.Duration
}

func NewAuthUsecase(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, tokenService service.TokenServiceInterface, logger logger.Logger) *AuthUsecase {
	return &AuthUsecase{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		tokenService:    tokenService,
		logger:          logger,
		RefreshTokenTTL: time.Hour * 24 * 30, // 30 days
	}
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
	uc.logger.Info("attempting login",
		logger.NewField("username", username),
	)

	user, err := uc.userRepo.GetByUsername(ctx, username)
	if err != nil {
		uc.logger.Error("failed to get user",
			logger.NewField("error", err),
			logger.NewField("username", username),
		)
		return nil, nil, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		uc.logger.Warn("invalid password attempt",
			logger.NewField("username", username),
		)
		return nil, nil, errors.ErrWrongPassword
	}

	tokens, err := uc.tokenService.GenerateTokenPair(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		uc.logger.Error("failed to generate tokens",
			logger.NewField("error", err),
			logger.NewField("user_id", user.ID),
		)
		return nil, nil, err
	}

	if err := uc.tokenRepo.Create(ctx, &entities.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: time.Now().Add(uc.RefreshTokenTTL),
	}); err != nil {
		uc.logger.Error("failed to save refresh token",
			logger.NewField("error", err),
			logger.NewField("user_id", user.ID),
		)
		return nil, nil, err
	}

	uc.logger.Info("login successful",
		logger.NewField("user_id", user.ID),
		logger.NewField("username", username),
	)

	return tokens, user, nil
}

func (uc *AuthUsecase) RefreshTokens(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
	// Валидируем refresh token
	claims, err := uc.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	// Проверяем токен в базе
	dbToken, err := uc.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil || dbToken.Revoked || time.Now().After(dbToken.ExpiresAt) {
		return nil, errors.ErrTokenInvalid
	}

	// Генерируем новую пару токенов
	newTokens, err := uc.tokenService.GenerateTokenPair(claims.UserID, claims.Username, claims.IsAdmin)
	if err != nil {
		return nil, err
	}

	// Отзываем старый refresh token
	if err := uc.tokenRepo.Revoke(ctx, refreshToken); err != nil {
		return nil, err
	}

	// Сохраняем новый refresh token
	if err := uc.tokenRepo.Create(ctx, &entities.RefreshToken{
		UserID:    claims.UserID,
		Token:     newTokens.RefreshToken,
		ExpiresAt: time.Now().Add(uc.RefreshTokenTTL),
	}); err != nil {
		return nil, err
	}

	return newTokens, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, username, password string) (*entities.TokenPair, error) {
	uc.logger.Info("attempting registration",
		logger.NewField("username", username),
	)

	// // Валидация входных данных
	// if err := validateCredentials(username, password); err != nil {
	// 	uc.logger.Warn("invalid credentials provided",
	// 		logger.NewField("error", err),
	// 		logger.NewField("username", username),
	// 	)
	// 	return nil, err
	// }

	// Проверяем, не существует ли пользователь
	_, err := uc.userRepo.GetByUsername(ctx, username)
	if err == nil {
		uc.logger.Warn("user already exists",
			logger.NewField("username", username),
		)
		return nil, errors.ErrDuplicateUsername
	}

	// Хешируем пароль
	passwordHash, err := HashPassword(password)
	if err != nil {
		uc.logger.Error("failed to hash password",
			logger.NewField("error", err),
		)
		return nil, err
	}

	// Создаем пользователя
	user := &entities.User{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		IsAdmin:      false,
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		uc.logger.Error("failed to create user",
			logger.NewField("error", err),
			logger.NewField("username", username),
		)
		return nil, err
	}

	// Генерируем токены
	tokens, err := uc.tokenService.GenerateTokenPair(user.ID, user.Username, user.IsAdmin)
	if err != nil {
		uc.logger.Error("failed to generate tokens",
			logger.NewField("error", err),
			logger.NewField("user_id", user.ID),
		)
		return nil, err
	}

	// Сохраняем refresh token
	if err := uc.tokenRepo.Create(ctx, &entities.RefreshToken{
		UserID:    user.ID,
		Token:     tokens.RefreshToken,
		ExpiresAt: time.Now().Add(uc.RefreshTokenTTL),
	}); err != nil {
		uc.logger.Error("failed to save refresh token",
			logger.NewField("error", err),
			logger.NewField("user_id", user.ID),
		)
		return nil, err
	}

	uc.logger.Info("registration successful",
		logger.NewField("user_id", user.ID),
		logger.NewField("username", username),
	)

	return tokens, nil
}

func (uc *AuthUsecase) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin, nil
}

func (uc *AuthUsecase) RevokeTokens(ctx context.Context, userID int64) error {
	uc.logger.Info("attempting to revoke all user tokens",
		logger.NewField("user_id", userID),
	)

	if err := uc.tokenRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		uc.logger.Error("failed to revoke user tokens",
			logger.NewField("error", err),
			logger.NewField("user_id", userID),
		)
		return err
	}

	uc.logger.Info("successfully revoked all user tokens",
		logger.NewField("user_id", userID),
	)

	return nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	return uc.tokenRepo.Revoke(ctx, refreshToken)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func validateCredentials(username, password string) error {
	if len(username) < 3 {
		return errors.ErrUsernameTooShort
	}
	if len(username) > 50 {
		return errors.ErrUsernameTooLong
	}
	if len(password) < 4 {
		return errors.ErrWeakPassword
	}
	return nil
}

func (a *AuthUsecase) ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error) {
	claims, err := a.tokenService.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (a *AuthUsecase) DeleteExpired(ctx context.Context) error {
	a.logger.Info("DELETE EXPIRED TOKENS")
	return a.tokenRepo.DeleteExpired(ctx)
}
