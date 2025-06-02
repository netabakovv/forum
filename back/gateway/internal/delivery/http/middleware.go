package http

import (
	"net/http"

	pb "github.com/netabakovv/forum/back/proto"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authClient pb.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Получаем access token из заголовка Authorization
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "токен не найден"})
			return
		}

		// 2. Удаляем префикс "Bearer "
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 3. Валидируем токен через gRPC Auth-сервис
		resp, err := authClient.ValidateToken(c.Request.Context(), &pb.ValidateRequest{
			AccessToken: token,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
			return
		}

		// 4. Сохраняем userID и username в контекст
		c.Set("userID", resp.UserId)
		c.Set("username", resp.Username)
		c.Set("isAdmin", resp.IsAdmin)

		c.Next()
	}
}
