package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"back/auth_service/internal/entities"
)

// MockAuthUsecase мок для AuthUsecaseInterface
type MockAuthUsecase struct {
	RegisterFunc      func(ctx context.Context, username, password string) (*entities.TokenPair, error)
	LoginFunc         func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error)
	RefreshTokensFunc func(ctx context.Context, refreshToken string) (*entities.TokenPair, error)
	IsAdminFunc       func(ctx context.Context, userID int64) (bool, error)
	RevokeTokensFunc  func(ctx context.Context, userID int64) error
	LogoutFunc        func(ctx context.Context, refreshToken string) error
	ValidateTokenFunc func(ctx context.Context, token string) (*entities.TokenClaims, error)
}

func (m *MockAuthUsecase) Register(ctx context.Context, username, password string) (*entities.TokenPair, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, username, password)
	}
	return nil, nil
}

func (m *MockAuthUsecase) Login(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, username, password)
	}
	return nil, nil, nil
}

func (m *MockAuthUsecase) RefreshTokens(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
	if m.RefreshTokensFunc != nil {
		return m.RefreshTokensFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockAuthUsecase) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(ctx, userID)
	}
	return false, nil
}

func (m *MockAuthUsecase) RevokeTokens(ctx context.Context, userID int64) error {
	if m.RevokeTokensFunc != nil {
		return m.RevokeTokensFunc(ctx, userID)
	}
	return nil
}

func (m *MockAuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(ctx, refreshToken)
	}
	return nil
}

func (m *MockAuthUsecase) ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return nil, nil
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *entities.TokenPair
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockResponse: &entities.TokenPair{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid input",
		},
		{
			name: "usecase error",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockError:      errors.New("registration failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "registration failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка
			mockUC := &MockAuthUsecase{
				RegisterFunc: func(ctx context.Context, username, password string) (*entities.TokenPair, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			handler := &AuthHandler{uc: mockUC}

			// Создание запроса
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Выполнение
			handler.Register(w, req)

			// Проверка
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				if !strings.Contains(w.Body.String(), tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
				}
			}

			if tt.mockResponse != nil && w.Code == http.StatusOK {
				var response entities.TokenPair
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.AccessToken != tt.mockResponse.AccessToken {
					t.Errorf("expected access token %q, got %q", tt.mockResponse.AccessToken, response.AccessToken)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *entities.TokenPair
		mockUser       *entities.User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockResponse: &entities.TokenPair{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			mockUser:       &entities.User{ID: 1, Username: "testuser"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid input",
		},
		{
			name: "login failed",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			mockError:      errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка
			mockUC := &MockAuthUsecase{
				LoginFunc: func(ctx context.Context, username, password string) (*entities.TokenPair, *entities.User, error) {
					return tt.mockResponse, tt.mockUser, tt.mockError
				},
			}

			handler := &AuthHandler{uc: mockUC}

			// Создание запроса
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Выполнение
			handler.Login(w, req)

			// Проверка
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				if !strings.Contains(w.Body.String(), tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
				}
			}

			if tt.mockResponse != nil && w.Code == http.StatusOK {
				var response entities.TokenPair
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.AccessToken != tt.mockResponse.AccessToken {
					t.Errorf("expected access token %q, got %q", tt.mockResponse.AccessToken, response.AccessToken)
				}
			}
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *entities.TokenPair
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful token refresh",
			requestBody: map[string]string{
				"refresh_token": "valid_refresh_token",
			},
			mockResponse: &entities.TokenPair{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid input",
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refresh_token": "invalid_refresh_token",
			},
			mockError:      errors.New("invalid token"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка
			mockUC := &MockAuthUsecase{
				RefreshTokensFunc: func(ctx context.Context, refreshToken string) (*entities.TokenPair, error) {
					return tt.mockResponse, tt.mockError
				},
			}

			handler := &AuthHandler{uc: mockUC}

			// Создание запроса
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Выполнение
			handler.Refresh(w, req)

			// Проверка
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				if !strings.Contains(w.Body.String(), tt.expectedBody) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
				}
			}

			if tt.mockResponse != nil && w.Code == http.StatusOK {
				var response entities.TokenPair
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if response.AccessToken != tt.mockResponse.AccessToken {
					t.Errorf("expected access token %q, got %q", tt.mockResponse.AccessToken, response.AccessToken)
				}
			}
		})
	}
}

// Дополнительные тесты для edge cases
func TestAuthHandler_EmptyBody(t *testing.T) {
	mockUC := &MockAuthUsecase{}
	handler := &AuthHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
