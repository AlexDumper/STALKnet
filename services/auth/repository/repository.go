package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewAuthRepository(db *sql.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		db:    db,
		redis: redis,
	}
}

// User модель пользователя
type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Session модель сессии
type Session struct {
	UserID       int       `json:"user_id"`
	Username     string    `json:"username"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUser создаёт нового пользователя
func (r *AuthRepository) CreateUser(ctx context.Context, username, passwordHash, email string) (int, error) {
	var id int
	query := `INSERT INTO users (username, password_hash, email, status) VALUES ($1, $2, $3, 'offline') RETURNING id`
	err := r.db.QueryRowContext(ctx, query, username, passwordHash, email).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUserByUsername находит пользователя по имени
func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password_hash, email, status, created_at FROM users WHERE username = $1`
	row := r.db.QueryRowContext(ctx, query, username)

	var user User
	var createdAt sql.NullTime
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}

	return &user, nil
}

// GetUserByID находит пользователя по ID
func (r *AuthRepository) GetUserByID(ctx context.Context, userID int) (*User, error) {
	query := `SELECT id, username, password_hash, email, status, created_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, userID)

	var user User
	var createdAt sql.NullTime
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Status, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		user.CreatedAt = createdAt.Time
	}

	return &user, nil
}

// UpdateUserStatus обновляет статус пользователя
func (r *AuthRepository) UpdateUserStatus(ctx context.Context, userID int, status string) error {
	query := `UPDATE users SET status = $1, last_seen = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, userID)
	return err
}

// CreateSession создаёт сессию в Redis
func (r *AuthRepository) CreateSession(ctx context.Context, session *Session) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// Ключ сессии
	sessionKey := fmt.Sprintf("session:%s", session.Token)
	refreshKey := fmt.Sprintf("refresh:%s", session.RefreshToken)

	// Сохраняем сессию
	err = r.redis.Set(ctx, sessionKey, sessionData, time.Until(session.ExpiresAt)).Err()
	if err != nil {
		return err
	}

	// Сохраняем refresh token
	err = r.redis.Set(ctx, refreshKey, session.Token, 7*24*time.Hour).Err()
	if err != nil {
		return err
	}

	// Индекс сессий по пользователю
	userSessionsKey := fmt.Sprintf("user_sessions:%d", session.UserID)
	err = r.redis.LPush(ctx, userSessionsKey, session.Token).Err()
	if err != nil {
		return err
	}

	return nil
}

// GetSession получает сессию по токену
func (r *AuthRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	sessionKey := fmt.Sprintf("session:%s", token)

	data, err := r.redis.Get(ctx, sessionKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetSessionByRefreshToken получает сессию по refresh токену
func (r *AuthRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	refreshKey := fmt.Sprintf("refresh:%s", refreshToken)

	token, err := r.redis.Get(ctx, refreshKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return r.GetSession(ctx, token)
}

// DeleteSession удаляет сессию
func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	session, err := r.GetSession(ctx, token)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	sessionKey := fmt.Sprintf("session:%s", token)
	refreshKey := fmt.Sprintf("refresh:%s", session.RefreshToken)
	userSessionsKey := fmt.Sprintf("user_sessions:%d", session.UserID)

	// Удаляем сессию
	_ = r.redis.Del(ctx, sessionKey).Err()
	_ = r.redis.Del(ctx, refreshKey).Err()

	// Удаляем из списка сессий пользователя
	_ = r.redis.LRem(ctx, userSessionsKey, 0, token).Err()

	return nil
}

// DeleteUserSessions удаляет все сессии пользователя
func (r *AuthRepository) DeleteUserSessions(ctx context.Context, userID int) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	tokens, err := r.redis.LRange(ctx, userSessionsKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, token := range tokens {
		_ = r.DeleteSession(ctx, token)
	}

	return nil
}

// GetUserSessions получает все активные сессии пользователя
func (r *AuthRepository) GetUserSessions(ctx context.Context, userID int) ([]*Session, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	tokens, err := r.redis.LRange(ctx, userSessionsKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	for _, token := range tokens {
		session, err := r.GetSession(ctx, token)
		if err == nil && session != nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}
