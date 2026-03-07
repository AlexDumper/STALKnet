package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/stalknet/services/chat/repository"
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
	sessionID := c.Query("session_id") // Получаем session_id

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
		Hub:       h.hub,
		Conn:      conn,
		UserID:    userID,
		Username:  username,
		SessionID: sessionID, // Сохраняем session_id
		RoomID:    roomID,
		Send:      make(chan []byte, 256),
	}

	// Регистрируем клиента
	client.Hub.Register <- client

	// Загружаем последние 50 сообщений из оперативной таблицы messages
	messages, err := h.repo.GetRecentMessages(c.Request.Context(), roomID, 50)
	if err != nil {
		log.Printf("Failed to load recent messages: %v", err)
		// Продолжаем подключение даже если загрузка не удалась
	} else {
		log.Printf("Loaded %d recent messages for room %d", len(messages), roomID)
		// Отправляем историю сообщений клиенту
		for _, msg := range messages {
			msgData := map[string]interface{}{
				"type":       "message",
				"room_id":    msg.RoomID,
				"user_id":    msg.UserID,
				"username":   msg.Username,
				"content":    msg.Content,
				"timestamp":  msg.Timestamp.Format(time.RFC3339),
				"from_history": true, // Флаг, что это сообщение из истории
			}
			jsonData, err := json.Marshal(msgData)
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, jsonData)
				if err != nil {
					log.Printf("Failed to send message history: %v", err)
					break
				}
			}
		}
		log.Printf("Sent %d messages to client", len(messages))
	}

	// Отправляем приветственное сообщение о подключении
	h.hub.Broadcast(roomID, 0, "system", username+" присоединился к чату", "system")

	// Запускаем горутину для записи
	go h.writePump(client)

	// Запускаем горутину для чтения
	go h.readPump(client, sessionID, clientIP, clientPort)
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

func (h *ChatHandler) readPump(client *hub.Client, sessionID, clientIP string, clientPort int) {
	defer func() {
		// Отправляем событие DISCONNECT при разрыве соединения
		go sendDisconnectEventToCompliance(client.UserID, client.Username, sessionID, clientIP, clientPort)

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
		var rawMsg map[string]interface{}
		if err := json.Unmarshal(message, &rawMsg); err != nil {
			log.Printf("JSON parse error: %v", err)
			continue
		}

		msgType, ok := rawMsg["type"].(string)
		if !ok {
			log.Printf("Invalid message type")
			continue
		}

		content, _ := rawMsg["content"].(string)

		// Обработка обычного сообщения
		if msgType == "message" {
			chatMsg := &repository.ChatMessage{
				RoomID:     client.RoomID,
				UserID:     client.UserID,
				Username:   client.Username,
				Content:    content,
				ClientIP:   clientIP,
				ClientPort: clientPort,
				MessageType: msgType,
				Timestamp:  time.Now(),
			}

			// Асинхронное сохранение в базу
			go func(m *repository.ChatMessage) {
				if err := h.repo.SaveMessage(context.Background(), m); err != nil {
					log.Printf("Failed to save message: %v", err)
				}
			}(chatMsg)

			// Отправляем в Compliance Service для ФЗ-374
			go sendToComplianceService(client.RoomID, client.UserID, client.Username, content, clientIP, clientPort, msgType)

			// Отправляем сообщение всем в комнате
			h.hub.Broadcast(client.RoomID, client.UserID, client.Username, content, msgType)
		}

		// Обработка личного сообщения
		if msgType == "private_message" {
			recipientUsername, _ := rawMsg["recipient_username"].(string)

			if recipientUsername == "" {
				sendError(client.Conn, "recipient_username is required")
				continue
			}

			if recipientUsername == client.Username {
				sendError(client.Conn, "Cannot send private message to yourself")
				continue
			}

			// Ищем получателя через Auth Service
			recipientID, recipientName, err := h.findUserByUsername(context.Background(), recipientUsername)
			if err != nil {
				sendError(client.Conn, fmt.Sprintf("User '%s' not found", recipientUsername))
				continue
			}

			// Создаём контакты: [отправитель, получатель]
			contacts := []repository.Contact{
				{ID: client.UserID, Name: client.Username},
				{ID: recipientID, Name: recipientName},
			}

			// Создаём сообщение для БД
			privateMsg := &repository.ChatMessage{
				RoomID:        client.RoomID,
				UserID:        client.UserID,
				Username:      client.Username,
				Content:       content,
				ClientIP:      clientIP,
				ClientPort:    clientPort,
				MessageType:   "private",
				Contacts:      contacts,
				RecipientID:   recipientID,
				RecipientName: recipientName,
				Timestamp:     time.Now(),
			}

			// Асинхронное сохранение в chat_messages (для ФЗ-374)
			go func(m *repository.ChatMessage) {
				if err := h.repo.SavePrivateMessage(context.Background(), m); err != nil {
					log.Printf("Failed to save private message: %v", err)
				}
			}(privateMsg)

			// Отправляем в Compliance Service
			go sendPrivateMessageToCompliance(client.RoomID, client.UserID, client.Username, recipientID, recipientName, content, clientIP, clientPort)

			// Проверяем: получатель онлайн?
			isOnline := h.hub.IsUserOnline(recipientID)

			if isOnline {
				// Преобразуем contacts в тип hub.Contact
				hubContacts := make([]hub.Contact, len(contacts))
				for i, c := range contacts {
					hubContacts[i] = hub.Contact{ID: c.ID, Name: c.Name}
				}

				// Отправляем сообщение через Broadcast с фильтрацией
				h.hub.BroadcastPrivate(client.RoomID, client.UserID, client.Username, content, "private", hubContacts)
			} else {
				// Получатель офлайн - сохраняем в private_messages_offline
				offlineMsg := &repository.OfflinePrivateMessage{
					SenderID:       client.UserID,
					SenderUsername: client.Username,
					RecipientID:    recipientID,
					Content:        content,
				}
				go func(m *repository.OfflinePrivateMessage) {
					if err := h.repo.SaveOfflinePrivateMessage(context.Background(), m); err != nil {
						log.Printf("Failed to save offline private message: %v", err)
					}
				}(offlineMsg)

				log.Printf("Offline private message saved: from=%s, to=%s (ID=%d)",
					client.Username, recipientName, recipientID)
			}

			// Подтверждение отправителю
			sendPrivateMessageSent(client.Conn, recipientName, content)
		}
	}
}

// sendError отправляет ошибку клиенту
func sendError(conn *websocket.Conn, errorMsg string) {
	resp := map[string]interface{}{
		"type":  "error",
		"error": errorMsg,
	}
	data, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, data)
}

// sendPrivateMessageSent отправляет подтверждение об отправке личного сообщения
func sendPrivateMessageSent(conn *websocket.Conn, recipientName, content string) {
	resp := map[string]interface{}{
		"type":             "private_message_sent",
		"recipient_username": recipientName,
		"content":          content,
		"timestamp":        time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, data)
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

// sendPrivateMessageToCompliance отправляет личное сообщение в Compliance Service
func sendPrivateMessageToCompliance(roomID, senderID int, senderUsername string, recipientID int, recipientUsername, content, clientIP string, clientPort int) {
	complianceMsg := struct {
		RoomID           int       `json:"room_id"`
		SenderID         int       `json:"sender_id"`
		SenderUsername   string    `json:"sender_username"`
		RecipientID      int       `json:"recipient_id"`
		RecipientUsername string   `json:"recipient_username"`
		Content          string    `json:"content"`
		ClientIP         string    `json:"client_ip"`
		ClientPort       int       `json:"client_port"`
		MessageType      string    `json:"message_type"`
		Contacts         []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"contacts"`
		Timestamp time.Time `json:"timestamp"`
	}{
		RoomID:            roomID,
		SenderID:          senderID,
		SenderUsername:    senderUsername,
		RecipientID:       recipientID,
		RecipientUsername: recipientUsername,
		Content:           content,
		ClientIP:          clientIP,
		ClientPort:        clientPort,
		MessageType:       "private",
		Timestamp:         time.Now(),
	}
	
	// Добавляем контакты
	complianceMsg.Contacts = []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		{ID: senderID, Name: senderUsername},
		{ID: recipientID, Name: recipientUsername},
	}

	jsonData, err := json.Marshal(complianceMsg)
	if err != nil {
		log.Printf("Failed to marshal private compliance message: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", complianceServiceURL+"/api/compliance/messages", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Failed to create private compliance request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send private message to compliance service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Compliance service returned status: %d", resp.StatusCode)
	} else {
		log.Printf("Private message sent to compliance: from=%s, to=%s", senderUsername, recipientUsername)
	}
}

// sendDisconnectEventToCompliance отправляет событие DISCONNECT в Compliance Service
func sendDisconnectEventToCompliance(userID int, username, sessionID, clientIP string, clientPort int) {
	event := struct {
		EventType  string    `json:"event_type"`
		UserID     int       `json:"user_id"`
		Username   string    `json:"username"`
		SessionID  string    `json:"session_id"`
		ClientIP   string    `json:"client_ip"`
		ClientPort int       `json:"client_port"`
		LoginTime  time.Time `json:"login_time"`
	}{
		EventType:  "DISCONNECT",
		UserID:     userID,
		Username:   username,
		SessionID:  sessionID,
		ClientIP:   clientIP,
		ClientPort: clientPort,
		LoginTime:  time.Now(),
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal disconnect event: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", complianceServiceURL+"/api/compliance/sessions", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Failed to create compliance request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send disconnect event to compliance service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("Compliance service returned status: %d", resp.StatusCode)
	} else {
		log.Printf("Disconnect event sent to compliance: user=%s, ip=%s", username, clientIP)
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
