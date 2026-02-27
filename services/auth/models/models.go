package models

import "time"

// User модель пользователя
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

// Session сессия пользователя
type Session struct {
	UserID     int       `json:"user_id"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}
