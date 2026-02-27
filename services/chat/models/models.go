package models

import "time"

// Room комната чата
type Room struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   int       `json:"created_by"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}

// Message сообщение в чате
type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// RoomMember участник комнаты
type RoomMember struct {
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	JoinedAt  time.Time `json:"joined_at"`
}

// CreateRoomRequest запрос создания комнаты
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

// SendMessageRequest запрос отправки сообщения
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}
