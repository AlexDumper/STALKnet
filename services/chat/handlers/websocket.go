package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stalknet/services/chat/hub"
)

var complianceServiceURL = os.Getenv("COMPLIANCE_SERVICE_URL")

func init() {
	if complianceServiceURL == "" {
		complianceServiceURL = "http://localhost:8086"
	}
}

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

	// Получаем IP и порт клиента
	clientIP, clientPort := getClientIPAndPort(c.Request)

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
	go h.readPump(client, clientIP, clientPort)
}

// getClientIPAndPort извлекает IP адрес и порт клиента из запроса
func getClientIPAndPort(r *http.Request) (string, int) {
	// Проверяем заголовок X-Forwarded-For (для reverse proxy)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			return ip, 0 // Порт неизвестен через proxy
		}
	}

	// Проверяем заголовок X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri, 0
	}

	// Получаем из RemoteAddr
	host, portStr, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Если RemoteAddr без порта
		return r.RemoteAddr, 0
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return host, 0
	}

	return host, port
}

func (h *ChatHandler) readPump(client *hub.Client, clientIP string, clientPort int) {
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

		// Отправляем сообщение в Compliance Service для сохранения
		if msg.Type == "message" {
			go sendToComplianceService(client.RoomID, client.UserID, client.Username, msg.Content, clientIP, clientPort, msg.Type)
		}

		// Отправляем сообщение всем в комнате
		h.hub.Broadcast(client.RoomID, client.UserID, client.Username, msg.Content, msg.Type)
	}
}

// sendToComplianceService отправляет сообщение в Compliance Service для сохранения
func sendToComplianceService(roomID, userID int, username, content, clientIP string, clientPort int, msgType string) {
	complianceMsg := struct {
		RoomID      int       `json:"room_id"`
		UserID      int       `json:"user_id"`
		Username    string    `json:"username"`
		Content     string    `json:"content"`
		ClientIP    string    `json:"client_ip"`
		ClientPort  int       `json:"client_port"`
		MessageType string    `json:"message_type"`
		Timestamp   time.Time `json:"timestamp"`
	}{
		RoomID:      roomID,
		UserID:      userID,
		Username:    username,
		Content:     content,
		ClientIP:    clientIP,
		ClientPort:  clientPort,
		MessageType: msgType,
		Timestamp:   time.Now(),
	}

	jsonData, err := json.Marshal(complianceMsg)
	if err != nil {
		log.Printf("Failed to marshal compliance message: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", complianceServiceURL+"/api/compliance/messages", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Failed to create compliance request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send message to compliance service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Compliance service returned status: %d", resp.StatusCode)
	} else {
		log.Printf("Message sent to compliance service: user=%s, room=%d", username, roomID)
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
