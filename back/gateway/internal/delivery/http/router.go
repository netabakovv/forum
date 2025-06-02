package http

import (
	"net/http"

	"github.com/netabakovv/forum/back/gateway/internal/handler"

	_ "github.com/netabakovv/forum/back/gateway/cmd/docs" // Импорт сгенерированной документации

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, h *handler.Handler) {
	// Группа защищенных маршрутов
	protected := r.Group("/api")
	protected.Use(AuthMiddleware(h.Auth))

	// Профиль пользователя
	protected.GET("/profile", func(c *gin.Context) {
		userID := c.GetString("userID")
		username := c.GetString("username")
		isAdmin := c.GetBool("is_admin")

		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
			"isAdmin":  isAdmin,
		})
	})

	// Аутентификация
	r.POST("/register", h.Register())
	r.POST("/login", h.Login())
	r.POST("/refresh", h.RefreshToken())
	protected.POST("/logout", h.Logout())

	// Посты
	r.GET("/posts", h.GetPosts())
	r.GET("/posts/:id", h.GetPost())
	protected.POST("/posts", h.CreatePost())
	protected.DELETE("/posts/:id", h.DeletePost())

	// Комментарии
	r.GET("/comments/:id", h.GetCommentByID())
	r.GET("/comments/post/:postID", h.GetCommentsByPostID())
	protected.POST("/comments", h.CreateComment())
	protected.DELETE("/comments/:id", h.DeleteComment())

	// Чат
	protected.POST("/chat", h.SendMessage())
	r.GET("/chat", h.GetMessages())

	// WebSocket
	r.GET("/ws/chat", func(c *gin.Context) {
		target := "ws://localhost:8080/ws/chat"
		http.Redirect(c.Writer, c.Request, target, http.StatusTemporaryRedirect)
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
