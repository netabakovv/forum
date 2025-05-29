package handler

import (
	"back/pkg/logger"
	pb "back/proto"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
			"user":          resp.User,
		})

	}
}

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

func (h *Handler) CreateComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req pb.CreateCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		userID, _ := c.Get("userID")
		req.AuthorId = userID.(int64)

		resp, err := h.Forum.CreateComment(context.Background(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("не удалось создать комментарий: %v", err)})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// --- Post operations ---

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

func (h *Handler) DeletePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("id")
		fmt.Printf(postIDStr, "\n")
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный id поста %v", err)})
			return
		}
		req := pb.DeletePostRequest{PostId: postID}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("неверный формат запроса %v", err)})
			return
		}
		_, err = h.Forum.DeletePost(c, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка удаления поста: %v", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "пост удален"})
	}
}

// --- Comment operations ---

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

func (h *Handler) DeleteComment() gin.HandlerFunc {
	return func(c *gin.Context) {
		commentIDstr := c.Param("id")
		commentID, err := strconv.ParseInt(commentIDstr, 10, 64)
		if commentID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "идентификатор комментария обязателен"})
			return
		}

		userIDstr := c.GetString("userID")
		userID, err := strconv.ParseInt(userIDstr, 10, 64)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "не авторизован"})
			return
		}

		req := pb.DeleteCommentRequest{
			CommentId: commentID,
			UserId:    userID,
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
