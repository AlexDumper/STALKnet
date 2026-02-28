package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stalknet/services/chat/hub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все origin для разработки
	},
}

// HandleWebSocket обрабатывает WebSocket соединения
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Получаем room_id и user_id из query параметров
	roomIDStr := c.Query("room_id")
	userIDStr := c.Query("user_id")
	username := c.Query("username")

	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		log.Printf("Invalid room_id: %v", err)
		conn.Close()
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Invalid user_id: %v", err)
		conn.Close()
		return
	}

	if username == "" {
		log.Printf("Username required")
		conn.Close()
		return
	}

	client := &hub.Client{
		Hub:      h.hub,
		Conn:     conn,
		UserID:   userID,
		Username: username,
		RoomID:   roomID,
		Send:     make(chan []byte, 256),
	}

	// Регистрируем клиента
	client.Hub.Register <- client

	// Отправляем приветственное сообщение
	h.hub.Broadcast(roomID, 0, "system", username+" присоединился к чату", "system")

	// Запускаем горутину для записи
	go h.writePump(client)

	// Запускаем горутину для чтения
	go h.readPump(client)
}

func (h *ChatHandler) readPump(client *hub.Client) {
	defer func() {
		client.Hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Обрабатываем JSON сообщение
		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("JSON parse error: %v", err)
			continue
		}

		// Отправляем сообщение всем в комнате
		h.hub.Broadcast(client.RoomID, client.UserID, client.Username, msg.Content, msg.Type)
	}
}

func (h *ChatHandler) writePump(client *hub.Client) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
