package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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
