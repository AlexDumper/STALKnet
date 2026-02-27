package models

import "time"

// User модель пользователя
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// StatusUpdate запрос обновления статуса
type StatusUpdate struct {
	Status string `json:"status"`
}
