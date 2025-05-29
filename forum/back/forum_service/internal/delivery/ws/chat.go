// forum_service/internal/delivery/ws/chat.go
package ws

import (
	"back/forum_service/internal/entities"
	"back/forum_service/internal/usecase"
	"back/pkg/logger"
	pb "back/proto"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	upgrader websocket.Upgrader
	clients  sync.Map
	chatUC   *usecase.ChatUsecase
	logger   logger.Logger
	config   *pb.ChatConfig
}

func NewChatHandler(chatUC *usecase.ChatUsecase, logger logger.Logger, config *pb.ChatConfig) *ChatHandler {
	return &ChatHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Разрешаем все подключения для учебного проекта
			},
		},
		clients: sync.Map{},
		chatUC:  chatUC,
		logger:  logger,
		config:  config,
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из токена (для учебного проекта можно хардкодить)
	userID := int64(1) // Для теста

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ошибка подключения websocket",
			logger.NewField("error", err))
		return
	}
	defer conn.Close()

	// Добавляем клиента
	h.clients.Store(conn, true)
	defer h.clients.Delete(conn)

	// Чтение сообщений
	for {
		var msg entities.ChatMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break // Клиент отключился
		}

		msg.UserID = userID
		msg.CreatedAt = time.Now()

		// Сохраняем и рассылаем сообщение
		if err := h.handleMessage(&msg); err != nil {
			h.logger.Error("ошибка обработки сообщения",
				logger.NewField("error", err))
		}
	}
}

// handleMessage обрабатывает новое сообщение
func (h *ChatHandler) handleMessage(msg *entities.ChatMessage) error {
	if msg.Content == "" {
		return fmt.Errorf("пустое сообщение")
	}

	// Сохраняем
	ctx := context.Background()
	if err := h.chatUC.SendMessage(ctx, msg); err != nil {
		return err
	}

	// Рассылаем всем
	h.broadcast(msg)
	return nil
}

// broadcast рассылает сообщение всем подключенным клиентам
func (h *ChatHandler) broadcast(msg *entities.ChatMessage) {
	h.clients.Range(func(key, _ interface{}) bool {
		client := key.(*websocket.Conn)
		if err := client.WriteJSON(msg); err != nil {
			client.Close()
			h.clients.Delete(client)
		}
		return true
	})
}
