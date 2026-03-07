package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

// Contact представляет контакт в личном сообщении
type Contact struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// OfflinePrivateMessage представляет офлайн-приватное сообщение
type OfflinePrivateMessage struct {
	ID             int       `json:"id"`
	SenderID       int       `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	RecipientID    int       `json:"recipient_id"`
	Content        string    `json:"content"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// ChatMessage представляет сообщение чата для сохранения в БД
type ChatMessage struct {
	ID              int       `json:"id"`
	RoomID          int       `json:"room_id"`
	UserID          int       `json:"user_id"`
	Username        string    `json:"username"`
	Content         string    `json:"content"`
	ClientIP        string    `json:"client_ip"`
	ClientPort      int       `json:"client_port"`
	Timestamp       time.Time `json:"timestamp"`
	MessageType     string    `json:"message_type"`  // "message", "system", "task", "private"
	Contacts        []Contact `json:"contacts,omitempty"`  // Для личных сообщений
	RecipientID     int       `json:"recipient_id,omitempty"`
	RecipientName   string    `json:"recipient_username,omitempty"`
}

// ChatRepository репозиторий для работы с сообщениями чата
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository создаёт новый репозиторий
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

// SaveMessage сохраняет сообщение в базу данных
// Запись производится в обе таблицы: messages (оперативная) и chat_messages (ФЗ-374)
func (r *ChatRepository) SaveMessage(ctx context.Context, msg *ChatMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// 1. Вставляем в messages (оперативная таблица для быстрой загрузки истории)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO messages (room_id, user_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
	`, msg.RoomID, msg.UserID, msg.Content)
	if err != nil {
		log.Printf("Failed to insert into messages: %v", err)
		return err
	}

	// 2. Вставляем в chat_messages (для соблюдения ФЗ-374)
	query := `
		INSERT INTO chat_messages (room_id, user_id, username, content, client_ip, client_port, message_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = tx.QueryRowContext(ctx, query,
		msg.RoomID,
		msg.UserID,
		msg.Username,
		msg.Content,
		msg.ClientIP,
		msg.ClientPort,
		msg.MessageType,
		msg.Timestamp,
	).Scan(&msg.ID)
	if err != nil {
		log.Printf("Failed to insert into chat_messages: %v", err)
		return err
	}

	// 3. Удаляем старые сообщения из messages (оставляем последние 50 на комнату)
	_, err = tx.ExecContext(ctx, `
		DELETE FROM messages
		WHERE room_id = $1
		  AND id NOT IN (
		    SELECT id FROM messages
		    WHERE room_id = $1
		    ORDER BY created_at DESC
		    LIMIT 50
		  )
	`, msg.RoomID)
	if err != nil {
		log.Printf("Failed to cleanup old messages: %v", err)
		return err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Printf("Message saved: room=%d, user=%s, ip=%s:%d", msg.RoomID, msg.Username, msg.ClientIP, msg.ClientPort)
	return nil
}

// SavePrivateMessage сохраняет личное сообщение в базу данных
// Личные сообщения сохраняются только в chat_messages (не в messages)
func (r *ChatRepository) SavePrivateMessage(ctx context.Context, msg *ChatMessage) error {
	// Сериализация contacts в JSON
	contactsJSON, err := json.Marshal(msg.Contacts)
	if err != nil {
		log.Printf("Failed to marshal contacts: %v", err)
		return err
	}

	query := `
		INSERT INTO chat_messages (room_id, user_id, username, content, client_ip, client_port, message_type, contacts, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = r.db.QueryRowContext(ctx, query,
		msg.RoomID,
		msg.UserID,
		msg.Username,
		msg.Content,
		msg.ClientIP,
		msg.ClientPort,
		"private",
		contactsJSON,
		msg.Timestamp,
	).Scan(&msg.ID)
	if err != nil {
		log.Printf("Failed to insert private message into chat_messages: %v", err)
		return err
	}

	log.Printf("Private message saved: room=%d, from=%s, to=%s, ip=%s:%d",
		msg.RoomID, msg.Username, msg.RecipientName, msg.ClientIP, msg.ClientPort)
	return nil
}

// GetMessagesByRoom получает сообщения для указанной комнаты (из chat_messages)
// Используется для API запросов истории
func (r *ChatRepository) GetMessagesByRoom(ctx context.Context, roomID int, limit, offset int) ([]ChatMessage, error) {
	query := `
		SELECT id, room_id, user_id, username, content, client_ip, client_port, timestamp, message_type, contacts
		FROM chat_messages
		WHERE room_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		var contactsJSON sql.NullString
		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.ClientIP,
			&msg.ClientPort,
			&msg.Timestamp,
			&msg.MessageType,
			&contactsJSON,
		)
		if err != nil {
			return nil, err
		}
		
		// Парсим contacts если есть
		if contactsJSON.Valid && contactsJSON.String != "" {
			json.Unmarshal([]byte(contactsJSON.String), &msg.Contacts)
		}
		
		messages = append(messages, msg)
	}

	// Реверсируем порядок (новые сверху)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetRecentMessages получает последние сообщения для отображения при подключении
// Загружает из оперативной таблицы messages (быстрый доступ)
// Для личных сообщений используется chat_messages с contacts
func (r *ChatRepository) GetRecentMessages(ctx context.Context, roomID int, limit int) ([]ChatMessage, error) {
	query := `
		SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		var createdAt time.Time
		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}
		msg.Timestamp = createdAt
		msg.MessageType = "message"
		messages = append(messages, msg)
	}

	// Реверсируем порядок (старые сообщения первыми, новые последними)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessagesByUser получает сообщения от указанного пользователя
func (r *ChatRepository) GetMessagesByUser(ctx context.Context, userID int, limit int) ([]ChatMessage, error) {
	query := `
		SELECT id, room_id, user_id, username, content, client_ip, client_port, timestamp, message_type
		FROM chat_messages
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Username,
			&msg.Content,
			&msg.ClientIP,
			&msg.ClientPort,
			&msg.Timestamp,
			&msg.MessageType,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteOldMessages удаляет сообщения старше указанного периода
func (r *ChatRepository) DeleteOldMessages(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM chat_messages
		WHERE timestamp < NOW() - $1::interval
	`

	result, err := r.db.ExecContext(ctx, query, olderThan.String())
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	log.Printf("Deleted %d old chat messages", rowsAffected)
	return rowsAffected, nil
}

// GetTotalMessages возвращает общее количество сообщений
func (r *ChatRepository) GetTotalMessages(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chat_messages`).Scan(&count)
	return count, err
}

// GetMessagesCountByRoom возвращает количество сообщений в комнате
func (r *ChatRepository) GetMessagesCountByRoom(ctx context.Context, roomID int) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chat_messages WHERE room_id = $1`, roomID).Scan(&count)
	return count, err
}

// SaveOfflinePrivateMessage сохраняет офлайн-приватное сообщение
func (r *ChatRepository) SaveOfflinePrivateMessage(ctx context.Context, msg *OfflinePrivateMessage) error {
	query := `
		INSERT INTO private_messages_offline 
			(sender_id, sender_username, recipient_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, expires_at`
	
	return r.db.QueryRowContext(ctx, query,
		msg.SenderID, msg.SenderUsername, msg.RecipientID, msg.Content,
	).Scan(&msg.ID, &msg.CreatedAt, &msg.ExpiresAt)
}

// GetUnreadOfflineMessages получает непрочитанные офлайн-сообщения для пользователя
func (r *ChatRepository) GetUnreadOfflineMessages(ctx context.Context, recipientID int) ([]OfflinePrivateMessage, error) {
	query := `
		SELECT id, sender_id, sender_username, recipient_id, content, is_read, created_at, expires_at
		FROM private_messages_offline
		WHERE recipient_id = $1 AND is_read = FALSE
		ORDER BY created_at ASC`
	
	rows, err := r.db.QueryContext(ctx, query, recipientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []OfflinePrivateMessage
	for rows.Next() {
		var msg OfflinePrivateMessage
		err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.SenderUsername, &msg.RecipientID,
			&msg.Content, &msg.IsRead, &msg.CreatedAt, &msg.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	
	return messages, rows.Err()
}

// MarkOfflineMessagesAsRead помечает все офлайн-сообщения пользователя как прочитанные
func (r *ChatRepository) MarkOfflineMessagesAsRead(ctx context.Context, recipientID int) error {
	query := `
		UPDATE private_messages_offline
		SET is_read = TRUE
		WHERE recipient_id = $1 AND is_read = FALSE`
	
	_, err := r.db.ExecContext(ctx, query, recipientID)
	return err
}
