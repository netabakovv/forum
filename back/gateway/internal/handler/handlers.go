package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"

	"github.com/gin-gonic/gin"
)

type EmptyMessage struct{}

type Handler struct {
	Forum pb.ForumServiceClient
	Auth  pb.AuthServiceClient
	log   logger.Logger
}

func NewHandler(forumClient pb.ForumServiceClient, authClient pb.AuthServiceClient, log logger.Logger) *Handler {
	return &Handler{
		Forum: forumClient,
		Auth:  authClient,
		log:   log,
	}
}

// --- Auth ---

// @Summary Логин пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body pb.LoginRequest true "Данные для входа"
// @Success 200 {object} pb.LoginResponse "Токены и данные пользователя"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 401 {object} map[string]string "Ошибка авторизации"
// @Router /login [post]
func (h *Handler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		resp, err := h.Auth.Login(c, &pb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("ошибка авторизации: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_at":    resp.ExpiresAt,
			"is_admin":      resp.User.IsAdmin,
			"user":          resp.User,
		})

	}
}

// @Summary Регистрация пользователя
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body pb.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} pb.RegisterResponse "Токены после регистрации"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /register [post]
func (h *Handler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		_, err := h.Auth.Register(c, &pb.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка регистрации: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "успешно зарегистрирован"})
	}
}

// @Summary Обновить токен
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body pb.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} pb.RefreshTokenResponse "Новые токены"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка обновления токена"
// @Router /refresh [post]
func (h *Handler) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		resp, err := h.Auth.RefreshToken(c, &pb.RefreshTokenRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка обновления токена: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_at":    resp.ExpiresAt, // resp.ExpiresAt уже Unix-секунды из protobuf
		})
	}
}

func (h *Handler) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			AccessToken string `json:"access_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		_, err := h.Auth.ValidateToken(c, &pb.ValidateRequest{
			AccessToken: req.AccessToken,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка валидации токена: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "токен валиден"})
	}
}

// @Summary Выход из системы (Logout)
// @Tags Auth
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body pb.LogoutRequest true "Access token"
// @Success 200 {object} pb.LogoutResponse "Успешный выход"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка выхода"
// @Router /api/logout [post]
func (h *Handler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			AccessToken string `json:"access_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		_, err := h.Auth.Logout(c, &pb.LogoutRequest{
			AccessToken: req.AccessToken,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка выхода: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "выход успешен"})
	}
}

func (h *Handler) CheckAdminStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserId int64 `json:"user_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}

		_, err := h.Auth.CheckAdminStatus(c, &pb.CheckAdminRequest{
			UserId: req.UserId,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка проверки статуса админа: %v", err)})
			return
		}
	}
}

// --- Forum ---

// @Summary Получить список постов
// @Tags Posts
// @Produce json
// @Success 200 {object} pb.PostResponse "Созданный пост"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /posts [get]
func (h *Handler) GetPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := h.Forum.Posts(c, &pb.ListPostsRequest{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("не удалось получить посты: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp.Posts)
	}
}

// @Summary Создать пост
// @Tags Posts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param СreatePostRequest body pb.CreatePostRequest true "Данные нового поста"
// @Success 200 {object} pb.PostResponse "Созданный пост"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/posts [post]
func (h *Handler) CreatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		userID, _ := c.Get("userID")
		req.AuthorId = userID.(int64)

		resp, err := h.Forum.CreatePost(context.Background(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("не удалось создать пост: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Получить пост по ID
// @Tags Posts
// @Produce json
// @Param id path int true "ID поста"
// @Success 200 {object} pb.PostResponse "Пост"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /posts/{id} [get]
func (h *Handler) GetPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		postID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
			return
		}

		req := &pb.GetPostRequest{PostId: postID}
		resp, err := h.Forum.GetPost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка получения поста: %v", err)})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (h *Handler) UpdatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.UpdatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		resp, err := h.Forum.UpdatePost(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка обновления поста: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Удалить пост по ID
// @Tags Posts
// @Security ApiKeyAuth
// @Param id path int true "ID поста"
// @Success 200 {object} pb.EmptyMessage "Пустое сообщение"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/posts/{id} [delete]
func (h *Handler) DeletePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("id")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный id поста: %v", err)})
			return
		}

		req := &pb.DeletePostRequest{
			PostId: postID,
		}

		_, err = h.Forum.DeletePost(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка удаления поста: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "пост удален"})
	}
}

// --- Comment operations ---

// @Summary Создать комментарий
// @Tags Comments
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param createCommentRequest body pb.CreateCommentRequest true "Данные нового комментария"
// @Success 200 {object} pb.CommentResponse "Созданный комментарий"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/comments [post]
func (h *Handler) CreateComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.CreateCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		userID, _ := c.Get("userID")
		req.AuthorId = userID.(int64)

		resp, err := h.Forum.CreateComment(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("не удалось создать комментарий: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Получить комментарий по ID
// @Tags Comments
// @Produce json
// @Param id path int true "ID комментария"
// @Success 200 {object} pb.CommentResponse "Комментарий"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /comments/{id} [get]
func (h *Handler) GetCommentByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		commentID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID комментария"})
			return
		}

		req := &pb.GetCommentRequest{CommentId: commentID}
		resp, err := h.Forum.GetCommentByID(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка получения комментария: %v", err)})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Получить комментарии по ID поста
// @Tags Comments
// @Produce json
// @Param postID path int true "ID поста"
// @Success 200 {object} pb.ListCommentsResponse "Список комментариев"
// @Failure 400 {object} map[string]string "Неверный ID поста"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /comments/post/{postID} [get]
func (h *Handler) GetCommentsByPostID() gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("postID")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
			return
		}

		req := &pb.GetCommentsByPostIDRequest{PostId: postID}
		resp, err := h.Forum.GetByPostID(c, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка получения комментариев: %v", err)})
			return
		}

		c.JSON(http.StatusOK, resp.Comments)
	}
}

func (h *Handler) UpdateComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.UpdateCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		resp, err := h.Forum.UpdateComment(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка обновления комментария: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// @Summary Удалить комментарий по ID
// @Tags Comments
// @Security ApiKeyAuth
// @Param id path int true "ID комментария"
// @Success 200 {object} pb.EmptyMessage "Пустое сообщение"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/comments/{id} [delete]
func (h *Handler) DeleteComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		commentIDstr := c.Param("id")
		commentID, err := strconv.ParseInt(commentIDstr, 10, 64)
		if commentID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "идентификатор комментария обязателен"})
			return
		}

		// userIDstr := c.GetString("userID")
		// userID, err := strconv.ParseInt(userIDstr, 10, 64)
		// if userID == 0 {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "не авторизован"})
		// 	return
		// }

		req := pb.DeleteCommentRequest{
			CommentId: commentID,
		}

		_, err = h.Forum.DeleteComment(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка удаления комментария: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "комментарий удален"})
	}
}

func (h *Handler) Comments() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.ListCommentsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		resp, err := h.Forum.Comments(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка получения комментариев: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp.Comments)
	}
}

// --- Chat operations ---

// @Summary Отправить сообщение в чат
// @Tags Chat
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param message body pb.ChatMessage true "Сообщение для отправки"
// @Success 200 {object} pb.EmptyMessage "Пустой ответ"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/chat [post]
func (h *Handler) SendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg pb.ChatMessage
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		_, err := h.Forum.SendMessage(c, &msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка отправки сообщений %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "сообщение отправлено"})
	}
}

// @Summary Получить все сообщения чата
// @Tags Chat
// @Produce json
// @Success 200 {object} pb.GetMessagesResponse "Список сообщений чата"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /chat [get]
func (h *Handler) GetMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.GetMessagesRequest
		resp, err := h.Forum.GetMessages(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка получения сообщений %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp.Messages)
	}
}
