package http_test

// import (
// 	"back/auth_service/internal/service/mocks"
// 	"back/gateway/internal/handler"
// 	myhttp "back/gateway/internal/http"
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// )

// func setupRouter() *gin.Engine {
// 	gin.SetMode(gin.TestMode)

// 	// Создаём мок-хендлеры (в реальных тестах лучше использовать моки)
// 	h := &handler.Handler{
// 		Auth: mocks, // реализуй интерфейс
// 		// остальные хендлеры можно тоже замокать или оставить пустыми
// 	}

// 	r := gin.New()
// 	myhttp.RegisterRoutes(r, h)

// 	return r
// }

// // Пример мок клиента для Auth
// type mockAuthClient struct{}

// func (m mockAuthClient) ValidateToken(_, _ string) (bool, string, string, bool, error) {
// 	return true, "123", "testuser", true, nil
// }

// func TestGetPosts(t *testing.T) {
// 	router := setupRouter()

// 	req, _ := http.NewRequest("GET", "/posts", nil)
// 	resp := httptest.NewRecorder()
// 	router.ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusOK, resp.Code) // если реализовано
// }

// func TestProtectedProfileWithoutToken(t *testing.T) {
// 	router := setupRouter()

// 	req, _ := http.NewRequest("GET", "/api/profile", nil)
// 	resp := httptest.NewRecorder()
// 	router.ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusUnauthorized, resp.Code)
// }

// func TestProtectedProfileWithToken(t *testing.T) {
// 	router := setupRouter()

// 	req, _ := http.NewRequest("GET", "/api/profile", nil)
// 	req.Header.Set("Authorization", "Bearer valid-token")
// 	resp := httptest.NewRecorder()
// 	router.ServeHTTP(resp, req)

// 	assert.Equal(t, http.StatusOK, resp.Code)
// 	var body map[string]interface{}
// 	_ = json.NewDecoder(resp.Body).Decode(&body)

// 	assert.Equal(t, "123", body["userID"])
// 	assert.Equal(t, "testuser", body["username"])
// 	assert.Equal(t, true, body["isAdmin"])
// }
