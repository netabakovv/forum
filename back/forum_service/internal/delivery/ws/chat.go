package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/netabakovv/forum/back/forum_service/internal/entities"
	"github.com/netabakovv/forum/back/forum_service/internal/usecase"
	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"

	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	upgrader   websocket.Upgrader
	clients    sync.Map
	chatUC     *usecase.ChatUsecase
	authClient pb.AuthServiceClient
	logger     logger.Logger
	config     *pb.ChatConfig
}

func NewChatHandler(chatUC *usecase.ChatUsecase, logger logger.Logger, config *pb.ChatConfig, authClient pb.AuthServiceClient) *ChatHandler {
	return &ChatHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:    sync.Map{},
		chatUC:     chatUC,
		logger:     logger,
		config:     config,
		authClient: authClient,
	}
}

func (h *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Разрешить подключение с любого origin (на проде — поаккуратнее)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("не удалось апгрейдить соединение", logger.NewField("error", err))
		return
	}
	defer conn.Close()

	// Ожидаем первое сообщение: авторизация
	_, authMsg, err := conn.ReadMessage()
	if err != nil {
		h.logger.Error("ошибка чтения авторизационного сообщения", logger.NewField("error", err))
		return
	}

	var authData struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}

	if err := json.Unmarshal(authMsg, &authData); err != nil || authData.Type != "auth" || authData.Token == "" {
		h.logger.Error("невалидное авторизационное сообщение")
		conn.WriteJSON(map[string]string{"error": "unauthorized"})
		return
	}

	// Проверка токена через AuthService
	resp, err := h.authClient.ValidateToken(context.Background(), &pb.ValidateRequest{
		AccessToken: authData.Token,
	})
	if err != nil {
		h.logger.Error("невалидный токен", logger.NewField("error", err))
		conn.WriteJSON(map[string]string{"error": "unauthorized"})
		return
	}

	userID := resp.UserId
	username := resp.Username

	h.logger.Info("авторизация успешна", logger.NewField("userID", userID))

	// Дальнейшая обработка: чтение/запись сообщений
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			h.logger.Info("пользователь отключился", logger.NewField("userID", userID))
			break
		}

		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}

		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		if msg.Type == "message" {
			chatMsg := &entities.ChatMessage{
				UserID:    userID,
				Username:  username,
				Content:   msg.Content,
				CreatedAt: time.Now(),
			}

			if err := h.chatUC.SendMessage(context.Background(), chatMsg); err != nil {
				h.logger.Error("не удалось отправить сообщение", logger.NewField("error", err))
				continue
			}

			// Отправляем обратно клиенту подтверждённое сообщение
			conn.WriteJSON(chatMsg)
		}
		if msg.Type == "history" {
			h.logger.Error("HISTORY ADDDDDD")
			historyMessages, err := h.chatUC.GetMessages(context.Background())
			if err != nil {
				h.logger.Error("не удалось получить историю сообщений", logger.NewField("error", err))
				continue
			}

			// Отправим массив сообщений клиенту в формате:
			resp := struct {
				Type     string                  `json:"type"`
				Messages []*entities.ChatMessage `json:"messages"`
			}{
				Type:     "history",
				Messages: historyMessages,
			}

			conn.WriteJSON(resp)
		}
	}

}

func (h *ChatHandler) handleMessage(msg *entities.ChatMessage) error {
	if msg.Content == "" {
		return fmt.Errorf("пустое сообщение")
	}

	ctx := context.Background()
	if err := h.chatUC.SendMessage(ctx, msg); err != nil {
		return err
	}

	h.broadcast(msg)
	return nil
}

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
