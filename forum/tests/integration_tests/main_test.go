// integration_test/main_test.go
package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"back/proto"
)

var (
	authClient  proto.AuthServiceClient
	forumClient proto.ForumServiceClient
	db          *sql.DB
	testDBName  = "forum_test"
)

// TestMain настраивает тестовую среду
func TestMain(m *testing.M) {
	// Запуск тестовой базы данных
	if err := setupTestDB(); err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}

	// Запуск микросервисов
	if err := startMicroservices(); err != nil {
		log.Fatalf("Failed to start microservices: %v", err)
	}

	// Настройка gRPC клиентов
	if err := setupGRPCClients(); err != nil {
		log.Fatalf("Failed to setup gRPC clients: %v", err)
	}

	// Запуск тестов
	code := m.Run()

	// Очистка
	cleanup()
	os.Exit(code)
}

func setupTestDB() error {
	// Подключение к PostgreSQL для создания тестовой БД
	db, err := sql.Open("postgres", "postgres://user:password@localhost/postgres?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	// Создание тестовой БД
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		return err
	}

	// Подключение к тестовой БД
	testDB, err := sql.Open("postgres", fmt.Sprintf("postgres://user:password@localhost/%s?sslmode=disable", testDBName))
	if err != nil {
		return err
	}

	// Запуск миграций
	if err := runMigrations(testDB); err != nil {
		return err
	}

	db = testDB
	return nil
}

func runMigrations(db *sql.DB) error {
	// Здесь должны быть ваши SQL миграции
	migrations := []string{
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			author_id INTEGER REFERENCES users(id),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE comments (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			author_id INTEGER REFERENCES users(id),
			post_id INTEGER REFERENCES posts(id),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE chat_messages (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE refresh_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			token VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}

	return nil
}

func startMicroservices() error {
	// Здесь должен быть код для запуска ваших микросервисов
	// Можно использовать команды exec или docker-compose
	return nil
}

func setupGRPCClients() error {
	// Подключение к Auth Service
	authConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	authClient = proto.NewAuthServiceClient(authConn)

	// Подключение к Forum Service
	forumConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	forumClient = proto.NewForumServiceClient(forumConn)

	// Ожидание готовности сервисов
	if err := waitForServices(); err != nil {
		return err
	}

	return nil
}

func waitForServices() error {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		if isServiceReady("localhost:50051") && isServiceReady("localhost:50052") {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("services are not ready after %d seconds", maxRetries)
}

func isServiceReady(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func cleanup() {
	if db != nil {
		db.Close()
	}
}

// ================== ТЕСТЫ AUTH SERVICE ==================

func TestAuthService_CompleteFlow(t *testing.T) {
	ctx := context.Background()
	username := fmt.Sprintf("testuser_%d", time.Now().Unix())
	password := "testpassword123"

	// 1. Регистрация пользователя
	registerResp, err := authClient.Register(ctx, &proto.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if registerResp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if registerResp.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	if registerResp.ExpiresAt <= time.Now().Unix() {
		t.Error("Token should not be expired")
	}

	// 2. Валидация токена
	validateResp, err := authClient.ValidateToken(ctx, &proto.ValidateRequest{
		AccessToken: registerResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if !validateResp.IsValid {
		t.Error("Token should be valid")
	}
	if validateResp.Username != username {
		t.Errorf("Expected username %s, got %s", username, validateResp.Username)
	}

	// 3. Логин с теми же данными
	loginResp, err := authClient.Login(ctx, &proto.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	if loginResp.AccessToken == "" {
		t.Error("Access token should not be empty")
	}
	if loginResp.User.Username != username {
		t.Errorf("Expected username %s, got %s", username, loginResp.User.Username)
	}

	// 4. Обновление токена
	refreshResp, err := authClient.RefreshToken(ctx, &proto.RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Error("New access token should not be empty")
	}

	// 5. Получение профиля пользователя
	userResp, err := authClient.GetUserByID(ctx, &proto.GetUserRequest{
		UserId: validateResp.UserId,
	})
	if err != nil {
		t.Fatalf("Failed to get user profile: %v", err)
	}

	if userResp.Username != username {
		t.Errorf("Expected username %s, got %s", username, userResp.Username)
	}

	// 6. Логаут
	logoutResp, err := authClient.Logout(ctx, &proto.LogoutRequest{
		AccessToken: refreshResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("Failed to logout: %v", err)
	}

	if !logoutResp.Success {
		t.Error("Logout should be successful")
	}

	// 7. Проверка, что токен больше не валиден
	validateResp2, err := authClient.ValidateToken(ctx, &proto.ValidateRequest{
		AccessToken: refreshResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("Failed to validate token after logout: %v", err)
	}

	if validateResp2.IsValid {
		t.Error("Token should be invalid after logout")
	}
}

func TestAuthService_InvalidCredentials(t *testing.T) {
	ctx := context.Background()

	// Попытка входа с несуществующими данными
	_, err := authClient.Login(ctx, &proto.LoginRequest{
		Username: "nonexistentuser",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Error("Login with invalid credentials should fail")
	}
}

func TestAuthService_DuplicateRegistration(t *testing.T) {
	ctx := context.Background()
	username := fmt.Sprintf("duplicate_%d", time.Now().Unix())
	password := "testpassword123"

	// Первая регистрация
	_, err := authClient.Register(ctx, &proto.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Попытка повторной регистрации
	_, err = authClient.Register(ctx, &proto.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err == nil {
		t.Error("Duplicate registration should fail")
	}
}

// ================== ТЕСТЫ FORUM SERVICE ==================

func TestForumService_PostsFlow(t *testing.T) {
	ctx := context.Background()

	// Сначала создаем пользователя
	user := createTestUser(t)

	// 1. Создание поста
	createPostResp, err := forumClient.CreatePost(ctx, &proto.CreatePostRequest{
		Title:          "Test Post Title",
		Content:        "This is a test post content",
		AuthorId:       user.UserId,
		AuthorUsername: user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	postID := createPostResp.Post.Id
	if postID == 0 {
		t.Error("Post ID should not be zero")
	}
	if createPostResp.Post.Title != "Test Post Title" {
		t.Error("Post title mismatch")
	}

	// 2. Получение поста
	getPostResp, err := forumClient.GetPost(ctx, &proto.GetPostRequest{
		PostId: postID,
	})
	if err != nil {
		t.Fatalf("Failed to get post: %v", err)
	}

	if getPostResp.Post.Id != postID {
		t.Error("Retrieved post ID mismatch")
	}

	// 3. Обновление поста
	updatePostResp, err := forumClient.UpdatePost(ctx, &proto.UpdatePostRequest{
		PostId:  postID,
		Title:   stringPtr("Updated Post Title"),
		Content: stringPtr("Updated content"),
	})
	if err != nil {
		t.Fatalf("Failed to update post: %v", err)
	}

	if updatePostResp.Post.Title != "Updated Post Title" {
		t.Error("Post title was not updated")
	}

	// 4. Получение списка постов
	listPostsResp, err := forumClient.Posts(ctx, &proto.ListPostsRequest{})
	if err != nil {
		t.Fatalf("Failed to list posts: %v", err)
	}

	if len(listPostsResp.Posts) == 0 {
		t.Error("Should have at least one post")
	}

	// 5. Удаление поста
	_, err = forumClient.DeletePost(ctx, &proto.DeletePostRequest{
		PostId: postID,
	})
	if err != nil {
		t.Fatalf("Failed to delete post: %v", err)
	}

	// 6. Проверка, что пост удален
	_, err = forumClient.GetPost(ctx, &proto.GetPostRequest{
		PostId: postID,
	})
	if err == nil {
		t.Error("Getting deleted post should fail")
	}
}

func TestForumService_CommentsFlow(t *testing.T) {
	ctx := context.Background()

	// Создаем пользователя и пост
	user := createTestUser(t)
	post := createTestPost(t, user)

	// 1. Создание комментария
	createCommentResp, err := forumClient.CreateComment(ctx, &proto.CreateCommentRequest{
		Content:        "This is a test comment",
		AuthorId:       user.UserId,
		PostId:         post.Id,
		AuthorUsername: user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	commentID := createCommentResp.Comment.Id
	if commentID == 0 {
		t.Error("Comment ID should not be zero")
	}

	// 2. Получение комментария
	getCommentResp, err := forumClient.GetCommentByID(ctx, &proto.GetCommentRequest{
		CommentId: commentID,
	})
	if err != nil {
		t.Fatalf("Failed to get comment: %v", err)
	}

	if getCommentResp.Comment.Id != commentID {
		t.Error("Retrieved comment ID mismatch")
	}

	// 3. Получение комментариев по посту
	getCommentsByPostResp, err := forumClient.GetByPostID(ctx, &proto.GetCommentsByPostIDRequest{
		PostId: post.Id,
	})
	if err != nil {
		t.Fatalf("Failed to get comments by post: %v", err)
	}

	if len(getCommentsByPostResp.Comments) == 0 {
		t.Error("Should have at least one comment")
	}

	// 4. Обновление комментария
	updateCommentResp, err := forumClient.UpdateComment(ctx, &proto.UpdateCommentRequest{
		CommentId: commentID,
		Content:   stringPtr("Updated comment content"),
	})
	if err != nil {
		t.Fatalf("Failed to update comment: %v", err)
	}

	if updateCommentResp.Comment.Content != "Updated comment content" {
		t.Error("Comment content was not updated")
	}

	// 5. Удаление комментария
	_, err = forumClient.DeleteComment(ctx, &proto.DeleteCommentRequest{
		CommentId: commentID,
		UserId:    user.UserId,
	})
	if err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}

	// 6. Проверка, что комментарий удален
	_, err = forumClient.GetCommentByID(ctx, &proto.GetCommentRequest{
		CommentId: commentID,
	})
	if err == nil {
		t.Error("Getting deleted comment should fail")
	}
}

func TestForumService_ChatFlow(t *testing.T) {
	ctx := context.Background()

	// Создаем пользователя
	user := createTestUser(t)

	// 1. Отправка сообщения
	_, err := forumClient.SendMessage(ctx, &proto.ChatMessage{
		UserId:    user.UserId,
		Content:   "Hello from integration test!",
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// 2. Получение сообщений
	getMessagesResp, err := forumClient.GetMessages(ctx, &proto.GetMessagesRequest{})
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(getMessagesResp.Messages) == 0 {
		t.Error("Should have at least one message")
	}

	// Проверяем, что наше сообщение есть в списке
	found := false
	for _, msg := range getMessagesResp.Messages {
		if msg.UserId == user.UserId && msg.Content == "Hello from integration test!" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sent message not found in messages list")
	}
}

// ================== ИНТЕГРАЦИОННЫЕ ТЕСТЫ МЕЖДУ СЕРВИСАМИ ==================

func TestCrossService_PostWithComments(t *testing.T) {
	ctx := context.Background()

	// 1. Создаем пользователя через Auth Service
	user := createTestUser(t)

	// 2. Создаем пост через Forum Service
	post := createTestPost(t, user)

	// 3. Создаем несколько комментариев
	for i := 0; i < 3; i++ {
		_, err := forumClient.CreateComment(ctx, &proto.CreateCommentRequest{
			Content:        fmt.Sprintf("Comment %d", i+1),
			AuthorId:       user.UserId,
			PostId:         post.Id,
			AuthorUsername: user.Username,
		})
		if err != nil {
			t.Fatalf("Failed to create comment %d: %v", i+1, err)
		}
	}

	// 4. Проверяем количество комментариев
	commentsResp, err := forumClient.GetByPostID(ctx, &proto.GetCommentsByPostIDRequest{
		PostId: post.Id,
	})
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(commentsResp.Comments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(commentsResp.Comments))
	}

	// 5. Проверяем обновленный профиль пользователя
	userProfile, err := authClient.GetUserByID(ctx, &proto.GetUserRequest{
		UserId: user.UserId,
	})
	if err != nil {
		t.Fatalf("Failed to get user profile: %v", err)
	}

	if userProfile.PostCount != 1 {
		t.Errorf("Expected 1 post, got %d", userProfile.PostCount)
	}
	if userProfile.CommentCount != 3 {
		t.Errorf("Expected 3 comments, got %d", userProfile.CommentCount)
	}
}

func TestCrossService_AdminPermissions(t *testing.T) {
	ctx := context.Background()

	// 1. Создаем обычного пользователя
	user := createTestUser(t)

	// 2. Проверяем статус админа
	adminResp, err := authClient.CheckAdminStatus(ctx, &proto.CheckAdminRequest{
		UserId: user.UserId,
	})
	if err != nil {
		t.Fatalf("Failed to check admin status: %v", err)
	}

	if adminResp.IsAdmin {
		t.Error("New user should not be admin")
	}

}

// ================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==================

func createTestUser(t *testing.T) *proto.UserProfileResponse {
	ctx := context.Background()
	username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())
	password := "testpassword123"

	registerResp, err := authClient.Register(ctx, &proto.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	validateResp, err := authClient.ValidateToken(ctx, &proto.ValidateRequest{
		AccessToken: registerResp.AccessToken,
	})
	if err != nil {
		t.Fatalf("Failed to validate test user token: %v", err)
	}

	return &proto.UserProfileResponse{
		UserId:   validateResp.UserId,
		Username: validateResp.Username,
		IsAdmin:  validateResp.IsAdmin,
	}
}

func createTestPost(t *testing.T, user *proto.UserProfileResponse) *proto.Post {
	ctx := context.Background()

	createPostResp, err := forumClient.CreatePost(ctx, &proto.CreatePostRequest{
		Title:          fmt.Sprintf("Test Post %d", time.Now().UnixNano()),
		Content:        "Test content",
		AuthorId:       user.UserId,
		AuthorUsername: user.Username,
	})
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	return createPostResp.Post
}

func stringPtr(s string) *string {
	return &s
}

// ================== БЕНЧМАРКИ ==================

func BenchmarkAuthService_LoginFlow(b *testing.B) {
	ctx := context.Background()
	user := createTestUserForBench(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := authClient.Login(ctx, &proto.LoginRequest{
			Username: user.Username,
			Password: "testpassword123",
		})
		if err != nil {
			b.Fatalf("Login failed: %v", err)
		}
	}
}

func BenchmarkForumService_CreatePost(b *testing.B) {
	ctx := context.Background()
	user := createTestUserForBench(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := forumClient.CreatePost(ctx, &proto.CreatePostRequest{
			Title:          fmt.Sprintf("Benchmark Post %d", i),
			Content:        "Benchmark content",
			AuthorId:       user.UserId,
			AuthorUsername: user.Username,
		})
		if err != nil {
			b.Fatalf("Create post failed: %v", err)
		}
	}
}

func createTestUserForBench(b *testing.B) *proto.UserProfileResponse {
	ctx := context.Background()
	username := fmt.Sprintf("benchuser_%d", time.Now().UnixNano())
	password := "testpassword123"

	registerResp, err := authClient.Register(ctx, &proto.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		b.Fatalf("Failed to create bench user: %v", err)
	}

	validateResp, err := authClient.ValidateToken(ctx, &proto.ValidateRequest{
		AccessToken: registerResp.AccessToken,
	})
	if err != nil {
		b.Fatalf("Failed to validate bench user token: %v", err)
	}

	return &proto.UserProfileResponse{
		UserId:   validateResp.UserId,
		Username: validateResp.Username,
		IsAdmin:  validateResp.IsAdmin,
	}
}
