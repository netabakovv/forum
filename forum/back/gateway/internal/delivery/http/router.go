package http

import (
	"back/gateway/internal/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handler.Handler) {
	protected := r.Group("/api")
	protected.Use(AuthMiddleware(h.Auth))

	protected.GET("/profile", func(c *gin.Context) {
		userID := c.GetString("userID")
		username := c.GetString("username")
		isAdmin := c.GetBool("isAdmin")

		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
			"isAdmin":  isAdmin,
		})
	})

	// Регистрация и аутентификация
	r.POST("/register", h.Register())
	r.POST("/login", h.Login())
	r.POST("/refresh", h.RefreshToken())
	protected.POST("/logout", h.Logout())

	// Посты
	protected.POST("/posts", h.CreatePost())       // создать пост
	r.GET("/posts", h.GetPosts())                  // получить список постов
	r.GET("/posts/:id", h.GetPost())               // получить пост по ID
	protected.DELETE("/posts/:id", h.DeletePost()) // удалить пост

	// Комментарии
	protected.POST("/comments", h.CreateComment())           // создать комментарий
	r.GET("/comments/:id", h.GetCommentByID())               // получить комментарий по ID
	protected.DELETE("/comments/:id", h.DeleteComment())     // удалить комментарий
	r.GET("/comments/post/:postID", h.GetCommentsByPostID()) // получить комментарии по посту

	// Чат
	protected.POST("/chat", h.SendMessage()) // отправить сообщение
	r.GET("/chat", h.GetMessages())          // получить все сообщения

	// Прокси WebSocket чат
	r.GET("/ws/chat", func(c *gin.Context) {
		// Прокидываем на форум-сервис напрямую
		target := "ws://localhost:8080/ws/chat"
		http.Redirect(c.Writer, c.Request, target, http.StatusTemporaryRedirect)
	})
}
