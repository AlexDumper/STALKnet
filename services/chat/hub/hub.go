package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Contact представляет контакт в личном сообщении
type Contact struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Client представляет подключённого клиента
type Client struct {
	Hub       *Hub
	Conn      *websocket.Conn
	UserID    int
	Username  string
	SessionID string  // Session ID для отслеживания сессий
	RoomID    int
	Send      chan []byte
}

// Hub управляет всеми клиентскими соединениями
type Hub struct {
	Clients     map[int]map[*Client]bool // roomID -> clients
	UserOnline  map[int]bool             // userID -> online статус
	Register    chan *Client
	Unregister  chan *Client
	MessageChan chan *Message
	Mu          sync.RWMutex
}

// Message сообщение для рассылки
type Message struct {
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // message, task, system
	Timestamp time.Time `json:"timestamp"`
}

// NewHub создаёт новый хаб
func NewHub() *Hub {
	return &Hub{
		Clients:     make(map[int]map[*Client]bool),
		UserOnline:  make(map[int]bool),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		MessageChan: make(chan *Message, 256),
	}
}

// Run запускает хаб
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			if _, ok := h.Clients[client.RoomID]; !ok {
				h.Clients[client.RoomID] = make(map[*Client]bool)
			}
			h.Clients[client.RoomID][client] = true
			h.UserOnline[client.UserID] = true
			h.Mu.Unlock()

		case client := <-h.Unregister:
			h.Mu.Lock()
			if clients, ok := h.Clients[client.RoomID]; ok {
				delete(clients, client)
				close(client.Send)
				if len(clients) == 0 {
					delete(h.Clients, client.RoomID)
				}
			}
			// Проверяем, есть ли другие клиенты этого пользователя
			hasOther := false
			for _, clients := range h.Clients {
				for c := range clients {
					if c.UserID == client.UserID && c != client {
						hasOther = true
						break
					}
				}
				if hasOther {
					break
				}
			}
			if !hasOther {
				delete(h.UserOnline, client.UserID)
			}
			h.Mu.Unlock()

		case message := <-h.MessageChan:
			h.Mu.RLock()
			if clients, ok := h.Clients[message.RoomID]; ok {
				// Сериализуем сообщение
				msgData, _ := json.Marshal(message)
				for client := range clients {
					select {
					case client.Send <- msgData:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.Mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем клиентам в комнате
func (h *Hub) Broadcast(roomID, userID int, username, content, msgType string) {
	h.MessageChan <- &Message{
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		Type:      msgType,
		Timestamp: time.Now(),
	}
}

// BroadcastPrivate отправляет личное сообщение только отправителю и получателю
func (h *Hub) BroadcastPrivate(roomID, senderID int, senderUsername, content, msgType string, contacts []Contact) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if clients, ok := h.Clients[roomID]; ok {
		// Сериализуем сообщение
		msgData := map[string]interface{}{
			"type":              "private_message",
			"sender_id":         senderID,
			"sender_username":   senderUsername,
			"content":           content,
			"message_type":      msgType,
			"contacts":          contacts,
			"timestamp":         time.Now().Format(time.RFC3339),
		}
		
		// Добавляем recipient_username из contacts
		if len(contacts) > 1 {
			msgData["recipient_username"] = contacts[1].Name
			msgData["recipient_id"] = contacts[1].ID
		}

		jsonData, err := json.Marshal(msgData)
		if err != nil {
			return
		}

		// Отправляем только клиентам из contacts
		for client := range clients {
			// Проверяем, является ли клиент отправителем или получателем
			canSee := false
			for _, contact := range contacts {
				if contact.ID == client.UserID {
					canSee = true
					break
				}
			}

			if canSee {
				select {
				case client.Send <- jsonData:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
		}
	}
}

// GetClientsInRoom возвращает всех клиентов в комнате
func (h *Hub) GetClientsInRoom(roomID int) []*Client {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if clients, ok := h.Clients[roomID]; ok {
		result := make([]*Client, 0, len(clients))
		for client := range clients {
			result = append(result, client)
		}
		return result
	}
	return nil
}

// GetClientCountInRoom возвращает количество клиентов в комнате
func (h *Hub) GetClientCountInRoom(roomID int) int {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if clients, ok := h.Clients[roomID]; ok {
		return len(clients)
	}
	return 0
}

// IsUserOnline проверяет, онлайн ли пользователь
func (h *Hub) IsUserOnline(userID int) bool {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	online, exists := h.UserOnline[userID]
	return exists && online
}

// GetOnlineUsers возвращает список онлайн-пользователей
func (h *Hub) GetOnlineUsers() []int {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	users := make([]int, 0, len(h.UserOnline))
	for userID, online := range h.UserOnline {
		if online {
			users = append(users, userID)
		}
	}
	return users
}
