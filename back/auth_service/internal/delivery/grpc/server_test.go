package grpc_test

import (
	"context"
	"testing"
	"time"

	"back/auth_service/internal/delivery/grpc"
	"back/auth_service/internal/entities"
	"back/pkg/errors"
	"back/pkg/logger"
	pb "back/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Моки для зависимостей
type mockAuthUsecase struct {
	registerFunc      func(ctx context.Context, username, password string) (*entities.TokenPair, error)
	loginFunc         func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error)
	refreshTokensFunc func(ctx context.Context, refreshToken string) (*entities.TokenPair, error)
	revokeTokensFunc  func(ctx context.Context, userID int64) error
	isAdminFunc       func(ctx context.Context, userID int64) (bool, error)
	logoutFunc        func(ctx context.Context, refreshToken string) error
	validateTokenFunc func(ctx context.Context, token string) (*entities.TokenClaims, error)
}

func (m *mockAuthUsecase) Register(ctx context.Context, username, password string) (*entities.TokenPair, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, username, password)
	}
	return nil, nil
}

func (m *mockAuthUsecase) Login(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, username, password)
	}
	return nil, nil, nil
}

func (m *mockAuthUsecase) RefreshTokens(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
	if m.refreshTokensFunc != nil {
		return m.refreshTokensFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *mockAuthUsecase) RevokeTokens(ctx context.Context, userID int64) error {
	if m.revokeTokensFunc != nil {
		return m.revokeTokensFunc(ctx, userID)
	}
	return nil
}

func (m *mockAuthUsecase) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	if m.isAdminFunc != nil {
		return m.isAdminFunc(ctx, userID)
	}
	return false, nil
}

func (m *mockAuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, refreshToken)
	}
	return nil
}

func (m *mockAuthUsecase) ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, token)
	}
	return nil, nil
}

type mockTokenService struct {
	generateTokenPairFunc func(userID int64, username string, isAdmin bool) (*entities.TokenPair, error)
	validateTokenFunc     func(tokenString string) (*entities.TokenClaims, error)
}

func (m *mockTokenService) GenerateTokenPair(userID int64, username string, isAdmin bool) (*entities.TokenPair, error) {
	if m.generateTokenPairFunc != nil {
		return m.generateTokenPairFunc(userID, username, isAdmin)
	}
	return nil, nil
}

func (m *mockTokenService) ValidateToken(tokenString string) (*entities.TokenClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(tokenString)
	}
	return nil, nil
}

type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Warn(msg string, fields ...logger.Field)  {}
func (m *mockLogger) Error(msg string, fields ...logger.Field) {}
func (m *mockLogger) Debug(msg string, fields ...logger.Field) {}
func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {}

// Тесты для Register
func TestAuthServer_Register(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.RegisterRequest
		mockSetup      func(*mockAuthUsecase)
		expectedError  codes.Code
		expectedResult bool
	}{
		{
			name: "успешная регистрация",
			request: &pb.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.registerFunc = func(ctx context.Context, username, password string) (*entities.TokenPair, error) {
					return &entities.TokenPair{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
						ExpiresAt:    time.Now().Add(time.Hour),
					}, nil
				}
			},
			expectedResult: true,
		},
		{
			name: "пустое имя пользователя",
			request: &pb.RegisterRequest{
				Username: "",
				Password: "password123",
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "пустой пароль",
			request: &pb.RegisterRequest{
				Username: "testuser",
				Password: "",
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "ошибка usecase",
			request: &pb.RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.registerFunc = func(ctx context.Context, username, password string) (*entities.TokenPair, error) {
					return nil, errors.ErrRegister
				}
			},
			expectedError: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := grpc.NewAuthServer(mockUC, &mockTokenService{}, &mockLogger{})

			resp, err := server.Register(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if tt.expectedResult && resp == nil {
				t.Error("ожидался успешный ответ, но получили nil")
			}
		})
	}
}

// Тесты для Login
func TestAuthServer_Login(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.LoginRequest
		mockSetup      func(*mockAuthUsecase)
		expectedError  codes.Code
		expectedResult bool
	}{
		{
			name: "успешный вход",
			request: &pb.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.loginFunc = func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
					return &entities.TokenPair{
							AccessToken:  "access_token",
							RefreshToken: "refresh_token",
							ExpiresAt:    time.Now().Add(time.Hour),
						}, &entities.User{
							ID:        1,
							Username:  "testuser",
							CreatedAt: time.Now(),
							IsAdmin:   false,
						}, nil
				}
			},
			expectedResult: true,
		},
		{
			name: "неверные учетные данные",
			request: &pb.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.loginFunc = func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
					return nil, nil, errors.ErrInvalidCredentials
				}
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "пользователь не найден",
			request: &pb.LoginRequest{
				Username: "nonexistentuser",
				Password: "password123",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.loginFunc = func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
					return nil, nil, errors.ErrUserNotFound
				}
			},
			expectedError: codes.NotFound,
		},
		{
			name: "пустые учетные данные",
			request: &pb.LoginRequest{
				Username: "",
				Password: "",
			},
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := grpc.NewAuthServer(mockUC, &mockTokenService{}, &mockLogger{})

			resp, err := server.Login(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if tt.expectedResult && resp == nil {
				t.Error("ожидался успешный ответ, но получили nil")
			}
		})
	}
}

// Тесты для RefreshToken
func TestAuthServer_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.RefreshTokenRequest
		mockSetup      func(*mockAuthUsecase)
		expectedError  codes.Code
		expectedResult bool
	}{
		{
			name: "успешное обновление токена",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "valid_refresh_token",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.refreshTokensFunc = func(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
					return &entities.TokenPair{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
						ExpiresAt:    time.Now().Add(time.Hour),
					}, nil
				}
			},
			expectedResult: true,
		},
		{
			name: "пустой refresh token",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "",
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "недействительный токен",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.refreshTokensFunc = func(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
					return nil, errors.ErrTokenInvalid
				}
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "просроченный токен",
			request: &pb.RefreshTokenRequest{
				RefreshToken: "expired_token",
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.refreshTokensFunc = func(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
					return nil, errors.ErrTokenExpired
				}
			},
			expectedError: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := grpc.NewAuthServer(mockUC, &mockTokenService{}, &mockLogger{})

			resp, err := server.RefreshToken(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if tt.expectedResult && resp == nil {
				t.Error("ожидался успешный ответ, но получили nil")
			}
		})
	}
}

// Тесты для ValidateToken
func TestAuthServer_ValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.ValidateRequest
		mockSetup     func(*mockTokenService)
		expectedError codes.Code
		expectedValid bool
	}{
		{
			name: "валидный токен",
			request: &pb.ValidateRequest{
				AccessToken: "valid_token",
			},
			mockSetup: func(m *mockTokenService) {
				m.validateTokenFunc = func(token string) (*entities.TokenClaims, error) {
					return &entities.TokenClaims{
						UserID:   1,
						Username: "testuser",
						IsAdmin:  false,
					}, nil
				}
			},
			expectedValid: true,
		},
		{
			name: "пустой токен",
			request: &pb.ValidateRequest{
				AccessToken: "",
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "просроченный токен",
			request: &pb.ValidateRequest{
				AccessToken: "expired_token",
			},
			mockSetup: func(m *mockTokenService) {
				m.validateTokenFunc = func(token string) (*entities.TokenClaims, error) {
					return nil, errors.ErrTokenExpired
				}
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "недействительный токен",
			request: &pb.ValidateRequest{
				AccessToken: "invalid_token",
			},
			mockSetup: func(m *mockTokenService) {
				m.validateTokenFunc = func(token string) (*entities.TokenClaims, error) {
					return nil, errors.ErrTokenInvalid
				}
			},
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTS := &mockTokenService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockTS)
			}

			server := grpc.NewAuthServer(&mockAuthUsecase{}, mockTS, &mockLogger{})

			resp, err := server.ValidateToken(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if resp.IsValid != tt.expectedValid {
				t.Errorf("ожидался IsValid=%v, получили %v", tt.expectedValid, resp.IsValid)
			}
		})
	}
}

// Тесты для Logout
func TestAuthServer_Logout(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.LogoutRequest
		mockTSSetup    func(*mockTokenService)
		mockUCSetup    func(*mockAuthUsecase)
		expectedError  codes.Code
		expectedResult bool
	}{
		{
			name: "успешный выход",
			request: &pb.LogoutRequest{
				AccessToken: "valid_token",
			},
			mockTSSetup: func(m *mockTokenService) {
				m.validateTokenFunc = func(token string) (*entities.TokenClaims, error) {
					return &entities.TokenClaims{
						UserID:   1,
						Username: "testuser",
					}, nil
				}
			},
			mockUCSetup: func(m *mockAuthUsecase) {
				m.revokeTokensFunc = func(ctx context.Context, userID int64) error {
					return nil
				}
			},
			expectedResult: true,
		},
		{
			name: "пустой токен",
			request: &pb.LogoutRequest{
				AccessToken: "",
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "недействительный токен",
			request: &pb.LogoutRequest{
				AccessToken: "invalid_token",
			},
			mockTSSetup: func(m *mockTokenService) {
				m.validateTokenFunc = func(token string) (*entities.TokenClaims, error) {
					return nil, errors.ErrTokenInvalid
				}
			},
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTS := &mockTokenService{}
			mockUC := &mockAuthUsecase{}

			if tt.mockTSSetup != nil {
				tt.mockTSSetup(mockTS)
			}
			if tt.mockUCSetup != nil {
				tt.mockUCSetup(mockUC)
			}

			server := grpc.NewAuthServer(mockUC, mockTS, &mockLogger{})

			resp, err := server.Logout(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if tt.expectedResult && (resp == nil || !resp.Success) {
				t.Error("ожидался успешный ответ")
			}
		})
	}
}

// Тесты для CheckAdminStatus
func TestAuthServer_CheckAdminStatus(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.CheckAdminRequest
		mockSetup     func(*mockAuthUsecase)
		expectedError codes.Code
		expectedAdmin bool
	}{
		{
			name: "пользователь администратор",
			request: &pb.CheckAdminRequest{
				UserId: 1,
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.isAdminFunc = func(ctx context.Context, userID int64) (bool, error) {
					return true, nil
				}
			},
			expectedAdmin: true,
		},
		{
			name: "пользователь не администратор",
			request: &pb.CheckAdminRequest{
				UserId: 2,
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.isAdminFunc = func(ctx context.Context, userID int64) (bool, error) {
					return false, nil
				}
			},
			expectedAdmin: false,
		},
		{
			name: "пустой user id",
			request: &pb.CheckAdminRequest{
				UserId: 0,
			},
			expectedError: codes.InvalidArgument,
		},
		{
			name: "ошибка usecase",
			request: &pb.CheckAdminRequest{
				UserId: 1,
			},
			mockSetup: func(m *mockAuthUsecase) {
				m.isAdminFunc = func(ctx context.Context, userID int64) (bool, error) {
					return false, errors.ErrDB
				}
			},
			expectedError: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockUC)
			}

			server := grpc.NewAuthServer(mockUC, &mockTokenService{}, &mockLogger{})

			resp, err := server.CheckAdminStatus(context.Background(), tt.request)

			if tt.expectedError != codes.OK {
				if err == nil {
					t.Errorf("ожидалась ошибка %v, но получили nil", tt.expectedError)
					return
				}
				if status.Code(err) != tt.expectedError {
					t.Errorf("ожидался код ошибки %v, получили %v", tt.expectedError, status.Code(err))
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка: %v", err)
				return
			}

			if resp.IsAdmin != tt.expectedAdmin {
				t.Errorf("ожидался IsAdmin=%v, получили %v", tt.expectedAdmin, resp.IsAdmin)
			}
		})
	}
}
