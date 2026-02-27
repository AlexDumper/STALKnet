package hub

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Client представляет подключённого клиента
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	userID   int
	roomID   int
	send     chan []byte
}

// Hub управляет всеми клиентскими соединениями
type Hub struct {
	clients    map[int]map[*Client]bool // roomID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

// Message сообщение для рассылки
type Message struct {
	RoomID  int             `json:"room_id"`
	UserID  int             `json:"user_id"`
	Content json.RawMessage `json:"content"`
	Type    string          `json:"type"` // message, task, system
}

// NewHub создаёт новый хаб
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

// Run запускает хаб
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.roomID]; !ok {
				h.clients[client.roomID] = make(map[*Client]bool)
			}
			h.clients[client.roomID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.roomID]; ok {
				delete(clients, client)
				close(client.send)
				if len(clients) == 0 {
					delete(h.clients, client.roomID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.RoomID]; ok {
				for client := range clients {
					select {
					case client.send <- message.Content:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем клиентам в комнате
func (h *Hub) Broadcast(roomID int, data []byte) {
	h.broadcast <- &Message{RoomID: roomID, Content: data}
}
