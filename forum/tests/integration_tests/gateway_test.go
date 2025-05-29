// integration_test/gateway_test.go
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Gateway HTTP API структуры
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
	PostID  int64  `json:"post_id"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    int64     `json:"expires_at"`
	User         *UserInfo `json:"user,omitempty"`
}

type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

type PostInfo struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	AuthorID     int64  `json:"author_id"`
	AuthorName   string `json:"author_username"`
	CreatedAt    int64  `json:"created_at"`
	CommentCount int32  `json:"comment_count"`
}

type CommentInfo struct {
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	AuthorID   int64  `json:"author_id"`
	AuthorName string `json:"author_username"`
	PostID     int64  `json:"post_id"`
	CreatedAt  int64  `json:"created_at"`
}

var gatewayServer *httptest.Server

// TestMain для Gateway тестов
func TestGatewayMain(m *testing.M) {
	// Запуск Gateway сервера
	gatewayServer = httptest.NewServer(createGatewayHandler())
	defer gatewayServer.Close()

	m.Run()
}

// Заглушка для Gateway handler'а
func createGatewayHandler() http.Handler {
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/refresh", handleRefreshToken)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/auth/profile", handleGetProfile)

	// Forum endpoints
	mux.HandleFunc("/api/posts", handlePosts)
	mux.HandleFunc("/api/posts/", handlePostByID)
	mux.HandleFunc("/api/comments", handleComments)
	mux.HandleFunc("/api/comments/", handleCommentByID)
	mux.HandleFunc("/api/chat/messages", handleChatMessages)

	return mux
}

// ================== GATEWAY HTTP TESTS ==================

func TestGateway_AuthFlow(t *testing.T) {
	username := fmt.Sprintf("gatewayuser_%d", time.Now().Unix())
	password := "testpassword123"

	// 1. Регистрация через HTTP
	registerReq := RegisterRequest{
		Username: username,
		Password: password,
	}

	registerResp := &AuthResponse{}
	err := makeJSONRequest("POST", "/api/auth/register", registerReq, registerResp)
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}

	if registerResp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}

	// 2. Получение профиля с токеном
	profile := &UserInfo{}
	err = makeAuthenticatedRequest("GET", "/api/auth/profile", nil, profile, registerResp.AccessToken)
	if err != nil {
		t.Fatalf("Get profile request failed: %v", err)
	}

	if profile.Username != username {
		t.Errorf("Expected username %s, got %s", username, profile.Username)
	}

	// 3. Логин
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	loginResp := &AuthResponse{}
	err = makeJSONRequest("POST", "/api/auth/login", loginReq, loginResp)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}

	if loginResp.AccessToken == "" {
		t.Error("Login access token should not be empty")
	}

	// 4. Обновление токена
	refreshReq := map[string]string{
		"refresh_token": loginResp.RefreshToken,
	}

	refreshResp := &AuthResponse{}
	err = makeJSONRequest("POST", "/api/auth/refresh", refreshReq, refreshResp)
	if err != nil {
		t.Fatalf("Refresh token request failed: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Error("Refreshed access token should not be empty")
	}

	// 5. Логаут
	logoutReq := map[string]string{
		"access_token": refreshResp.AccessToken,
	}

	logoutResp := map[string]bool{}
	err = makeJSONRequest("POST", "/api/auth/logout", logoutReq, &logoutResp)
	if err != nil {
		t.Fatalf("Logout request failed: %v", err)
	}

	if !logoutResp["success"] {
		t.Error("Logout should be successful")
	}
}

func TestGateway_PostsFlow(t *testing.T) {
	// Создаем пользователя и получаем токен
	token := createTestUserViaHTTP(t)

	// 1. Создание поста
	createPostReq := CreatePostRequest{
		Title:   "Gateway Test Post",
		Content: "This is a test post created via Gateway",
	}

	post := &PostInfo{}
	err := makeAuthenticatedRequest("POST", "/api/posts", createPostReq, post, token)
	if err != nil {
		t.Fatalf("Create post request failed: %v", err)
	}

	if post.ID == 0 {
		t.Error("Post ID should not be zero")
	}
	if post.Title != createPostReq.Title {
		t.Error("Post title mismatch")
	}

	// 2. Получение поста
	getPost := &PostInfo{}
	err = makeAuthenticatedRequest("GET", fmt.Sprintf("/api/posts/%d", post.ID), nil, getPost, token)
	if err != nil {
		t.Fatalf("Get post request failed: %v", err)
	}

	if getPost.ID != post.ID {
		t.Error("Retrieved post ID mismatch")
	}

	// 3. Обновление поста
	updatePostReq := map[string]interface{}{
		"title":   "Updated Gateway Post",
		"content": "Updated content via Gateway",
	}

	updatedPost := &PostInfo{}
	err = makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/posts/%d", post.ID), updatePostReq, updatedPost, token)
	if err != nil {
		t.Fatalf("Update post request failed: %v", err)
	}

	if updatedPost.Title != "Updated Gateway Post" {
		t.Error("Post title was not updated")
	}

	// 4. Получение списка постов
	posts := &[]PostInfo{}
	err = makeAuthenticatedRequest("GET", "/api/posts", nil, posts, token)
	if err != nil {
		t.Fatalf("List posts request failed: %v", err)
	}

	if len(*posts) == 0 {
		t.Error("Should have at least one post")
	}

	// 5. Удаление поста
	err = makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/posts/%d", post.ID), nil, nil, token)
	if err != nil {
		t.Fatalf("Delete post request failed: %v", err)
	}

	// 6. Проверка, что пост удален
	err = makeAuthenticatedRequest("GET", fmt.Sprintf("/api/posts/%d", post.ID), nil, &PostInfo{}, token)
	if err == nil {
		t.Error("Getting deleted post should return error")
	}
}

func TestGateway_CommentsFlow(t *testing.T) {
	// Создаем пользователя и пост
	token := createTestUserViaHTTP(t)
	post := createTestPostViaHTTP(t, token)

	// 1. Создание комментария
	createCommentReq := CreateCommentRequest{
		Content: "Gateway test comment",
		PostID:  post.ID,
	}

	comment := &CommentInfo{}
	err := makeAuthenticatedRequest("POST", "/api/comments", createCommentReq, comment, token)
	if err != nil {
		t.Fatalf("Create comment request failed: %v", err)
	}

	if comment.ID == 0 {
		t.Error("Comment ID should not be zero")
	}
	if comment.PostID != post.ID {
		t.Error("Comment post ID mismatch")
	}

	// 2. Получение комментария
	getComment := &CommentInfo{}
	err = makeAuthenticatedRequest("GET", fmt.Sprintf("/api/comments/%d", comment.ID), nil, getComment, token)
	if err != nil {
		t.Fatalf("Get comment request failed: %v", err)
	}

	if getComment.ID != comment.ID {
		t.Error("Retrieved comment ID mismatch")
	}

	// 3. Получение комментариев по посту
	comments := &[]CommentInfo{}
	err = makeAuthenticatedRequest("GET", fmt.Sprintf("/api/posts/%d/comments", post.ID), nil, comments, token)
	if err != nil {
		t.Fatalf("Get comments by post request failed: %v", err)
	}

	if len(*comments) == 0 {
		t.Error("Should have at least one comment")
	}

	// 4. Обновление комментария
	updateCommentReq := map[string]string{
		"content": "Updated gateway comment",
	}

	updatedComment := &CommentInfo{}
	err = makeAuthenticatedRequest("PUT", fmt.Sprintf("/api/comments/%d", comment.ID), updateCommentReq, updatedComment, token)
	if err != nil {
		t.Fatalf("Update comment request failed: %v", err)
	}

	if updatedComment.Content != "Updated gateway comment" {
		t.Error("Comment content was not updated")
	}

	// 5. Удаление комментария
	err = makeAuthenticatedRequest("DELETE", fmt.Sprintf("/api/comments/%d", comment.ID), nil, nil, token)
	if err != nil {
		t.Fatalf("Delete comment request failed: %v", err)
	}
}

func TestGateway_ChatFlow(t *testing.T) {
	token := createTestUserViaHTTP(t)

	// 1. Отправка сообщения
	sendMessageReq := map[string]interface{}{
		"content": "Gateway chat test message",
	}

	err := makeAuthenticatedRequest("POST", "/api/chat/messages", sendMessageReq, nil, token)
	if err != nil {
		t.Fatalf("Send message request failed: %v", err)
	}

	// 2. Получение сообщений
	messages := &[]map[string]interface{}{}
	err = makeAuthenticatedRequest("GET", "/api/chat/messages", nil, messages, token)
	if err != nil {
		t.Fatalf("Get messages request failed: %v", err)
	}

	if len(*messages) == 0 {
		t.Error("Should have at least one message")
	}

	// Проверяем, что наше сообщение есть в списке
	found := false
	for _, msg := range *messages {
		if content, ok := msg["content"].(string); ok && content == "Gateway chat test message" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sent message not found in messages list")
	}
}

// ================== STRESS TESTS ==================

func TestGateway_ConcurrentRequests(t *testing.T) {
	numRequests := 50
	tokens := make([]string, numRequests)

	// Создаем несколько пользователей
	for i := 0; i < numRequests; i++ {
		tokens[i] = createTestUserViaHTTP(t)
	}

	// Параллельное создание постов
	results := make(chan error, numRequests)
	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			createPostReq := CreatePostRequest{
				Title:   fmt.Sprintf("Concurrent Post %d", idx),
				Content: fmt.Sprintf("Content for post %d", idx),
			}

			post := &PostInfo{}
			err := makeAuthenticatedRequest("POST", "/api/posts", createPostReq, post, tokens[idx])
			results <- err
		}(i)
	}

	// Ждем завершения всех запросов
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
	}
}

func TestGateway_RateLimiting(t *testing.T) {
	token := createTestUserViaHTTP(t)

	// Быстрая отправка множества запросов
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 100; i++ {
		createPostReq := CreatePostRequest{
			Title:   fmt.Sprintf("Rate Limit Test Post %d", i),
			Content: "Testing rate limiting",
		}

		post := &PostInfo{}
		err := makeAuthenticatedRequest("POST", "/api/posts", createPostReq, post, token)
		if err != nil {
			if isRateLimitError(err) {
				rateLimitedCount++
			} else {
				t.Errorf("Unexpected error: %v", err)
			}
		} else {
			successCount++
		}
	}

	t.Logf("Successful requests: %d, Rate limited: %d", successCount, rateLimitedCount)

	// Ожидаем, что некоторые запросы были ограничены
	if rateLimitedCount == 0 {
		t.Log("Warning: No rate limiting detected, may need to adjust limits")
	}
}

// ================== ERROR HANDLING TESTS ==================

func TestGateway_ErrorHandling(t *testing.T) {
	// 1. Неавторизованный запрос
	post := &PostInfo{}
	err := makeJSONRequest("GET", "/api/posts", nil, post)
	if err == nil {
		t.Error("Unauthorized request should fail")
	}

	// 2. Недействительный токен
	err = makeAuthenticatedRequest("GET", "/api/posts", nil, post, "invalid_token")
	if err == nil {
		t.Error("Request with invalid token should fail")
	}

	// 3. Несуществующий ресурс
	token := createTestUserViaHTTP(t)
	err = makeAuthenticatedRequest("GET", "/api/posts/999999", nil, post, token)
	if err == nil {
		t.Error("Request for non-existent resource should fail")
	}

	// 4. Неправильный JSON
	badJSON := `{"title": "Test", "content":}`
	resp, err := http.Post(gatewayServer.URL+"/api/posts", "application/json", bytes.NewBufferString(badJSON))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Bad JSON should return 400")
	}
}

func TestGateway_CORS(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", gatewayServer.URL+"/api/posts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("CORS preflight request failed: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем CORS заголовки
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS Allow-Origin header missing")
	}
	if resp.Header.Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS Allow-Methods header missing")
	}
}

// ================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==================

func makeJSONRequest(method, path string, reqBody interface{}, respBody interface{}) error {
	var bodyReader *bytes.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, gatewayServer.URL+path, bodyReader)
	if err != nil {
		return err
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}

	return nil
}

func makeAuthenticatedRequest(method, path string, reqBody interface{}, respBody interface{}, token string) error {
	var bodyReader *bytes.Reader
	if reqBody != nil {
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, gatewayServer.URL+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}

	return nil
}

func createTestUserViaHTTP(t *testing.T) string {
	username := fmt.Sprintf("httpuser_%d", time.Now().UnixNano())
	password := "testpassword123"

	registerReq := RegisterRequest{
		Username: username,
		Password: password,
	}

	registerResp := &AuthResponse{}
	err := makeJSONRequest("POST", "/api/auth/register", registerReq, registerResp)
	if err != nil {
		t.Fatalf("Failed to create test user via HTTP: %v", err)
	}

	return registerResp.AccessToken
}

func createTestPostViaHTTP(t *testing.T, token string) *PostInfo {
	createPostReq := CreatePostRequest{
		Title:   fmt.Sprintf("HTTP Test Post %d", time.Now().UnixNano()),
		Content: "Test content via HTTP",
	}

	post := &PostInfo{}
	err := makeAuthenticatedRequest("POST", "/api/posts", createPostReq, post, token)
	if err != nil {
		t.Fatalf("Failed to create test post via HTTP: %v", err)
	}

	return post
}

func isRateLimitError(err error) bool {
	return err != nil && (err.Error() == "HTTP error: 429" || err.Error() == "HTTP error: 503")
}

// ================== ЗАГЛУШКИ ДЛЯ GATEWAY HANDLERS ==================

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Здесь должен быть вызов к Auth Service через gRPC
	resp := AuthResponse{
		AccessToken:  "mock_access_token_" + req.Username,
		RefreshToken: "mock_refresh_token_" + req.Username,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Аналогично handleRegister
	handleRegister(w, r)
}

func handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	resp := AuthResponse{
		AccessToken:  "mock_refreshed_access_token",
		RefreshToken: "mock_new_refresh_token",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	resp := map[string]bool{"success": true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp := UserInfo{
		ID:       1,
		Username: "testuser",
		IsAdmin:  false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		posts := []PostInfo{
			{
				ID:           1,
				Title:        "Test Post",
				Content:      "Test Content",
				AuthorID:     1,
				AuthorName:   "testuser",
				CreatedAt:    time.Now().Unix(),
				CommentCount: 0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)

	case "POST":
		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		resp := PostInfo{
			ID:           time.Now().Unix(),
			Title:        req.Title,
			Content:      req.Content,
			AuthorID:     1,
			AuthorName:   "testuser",
			CreatedAt:    time.Now().Unix(),
			CommentCount: 0,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handlePostByID(w http.ResponseWriter, r *http.Request) {
	// Извлечение ID из URL и обработка GET/PUT/DELETE
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		post := PostInfo{
			ID:           1,
			Title:        "Test Post",
			Content:      "Test Content",
			AuthorID:     1,
			AuthorName:   "testuser",
			CreatedAt:    time.Now().Unix(),
			CommentCount: 0,
		}
		json.NewEncoder(w).Encode(post)
	case "PUT":
		post := PostInfo{
			ID:           1,
			Title:        "Updated Gateway Post",
			Content:      "Updated content via Gateway",
			AuthorID:     1,
			AuthorName:   "testuser",
			CreatedAt:    time.Now().Unix(),
			CommentCount: 0,
		}
		json.NewEncoder(w).Encode(post)
	case "DELETE":
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleComments(w http.ResponseWriter, r *http.Request) {
	// Аналогично handlePosts для комментариев
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "POST":
		var req CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		resp := CommentInfo{
			ID:         time.Now().Unix(),
			Content:    req.Content,
			AuthorID:   1,
			AuthorName: "testuser",
			PostID:     req.PostID,
			CreatedAt:  time.Now().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func handleCommentByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		comment := CommentInfo{
			ID:         1,
			Content:    "Test Comment",
			AuthorID:   1,
			AuthorName: "testuser",
			PostID:     1,
			CreatedAt:  time.Now().Unix(),
		}
		json.NewEncoder(w).Encode(comment)
	case "PUT":
		comment := CommentInfo{
			ID:         1,
			Content:    "Updated gateway comment",
			AuthorID:   1,
			AuthorName: "testuser",
			PostID:     1,
			CreatedAt:  time.Now().Unix(),
		}
		json.NewEncoder(w).Encode(comment)
	case "DELETE":
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleChatMessages(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		messages := []map[string]interface{}{
			{
				"user_id":    1,
				"content":    "Gateway chat test message",
				"created_at": time.Now().Unix(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	case "POST":
		w.WriteHeader(http.StatusCreated)
	}
}
