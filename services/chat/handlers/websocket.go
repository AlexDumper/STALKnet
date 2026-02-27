package handlers

import (
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

	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		conn.Close()
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		conn.Close()
		return
	}

	client := &hub.Client{
		hub:    h.hub,
		conn:   conn,
		userID: userID,
		roomID: roomID,
		send:   make(chan []byte, 256),
	}

	client.hub.register <- client

	// Запускаем горутину для записи
	go client.writePump()

	// Запускаем горутину для чтения
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Обработка входящего сообщения
		c.hub.Broadcast(c.roomID, message)
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
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
