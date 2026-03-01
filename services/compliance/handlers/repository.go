package handlers

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// ComplianceRepository репозиторий для работы с сообщениями чата
type ComplianceRepository struct {
	db *sql.DB
}

// NewComplianceRepository создаёт новый репозиторий
func NewComplianceRepository(db *sql.DB) *ComplianceRepository {
	return &ComplianceRepository{
		db: db,
	}
}

// SaveMessage сохраняет сообщение в базу данных
func (r *ComplianceRepository) SaveMessage(ctx context.Context, msg *ChatMessage) error {
	query := `
		INSERT INTO chat_messages (room_id, user_id, username, content, client_ip, client_port, message_type, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
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
		log.Printf("Failed to save chat message: %v", err)
		return err
	}

	return nil
}

// GetMessagesByRoom получает сообщения для указанной комнаты
func (r *ComplianceRepository) GetMessagesByRoom(ctx context.Context, roomID int, limit, offset int) ([]ChatMessage, error) {
	query := `
		SELECT id, room_id, user_id, username, content, client_ip, client_port, timestamp, message_type
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

	// Реверсируем порядок (новые сверху)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessagesByUser получает сообщения от указанного пользователя
func (r *ComplianceRepository) GetMessagesByUser(ctx context.Context, userID int, limit int) ([]ChatMessage, error) {
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
func (r *ComplianceRepository) DeleteOldMessages(ctx context.Context, olderThan time.Duration) (int64, error) {
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

	log.Printf("Deleted %d old chat messages (older than %v)", rowsAffected, olderThan)
	return rowsAffected, nil
}

// GetTotalMessages возвращает общее количество сообщений
func (r *ComplianceRepository) GetTotalMessages(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chat_messages`).Scan(&count)
	return count, err
}

// SaveUserEvent сохраняет событие пользователя
func (r *ComplianceRepository) SaveUserEvent(ctx context.Context, event *UserEvent) error {
	query := `
		INSERT INTO user_events (event_type, user_id, username, client_ip, client_port, old_username, new_username, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE(NULLIF($8, ''), '{}')::jsonb, $9)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		event.EventType,
		event.UserID,
		event.Username,
		event.ClientIP,
		event.ClientPort,
		event.OldUsername,
		event.NewUsername,
		event.Metadata,
		event.Timestamp,
	).Scan(&event.ID)

	if err != nil {
		log.Printf("Failed to save user event: %v", err)
		return err
	}

	return nil
}

// GetUserEvents получает все события пользователей
func (r *ComplianceRepository) GetUserEvents(ctx context.Context, eventType string, limit, offset int) ([]UserEvent, error) {
	var query string
	var rows *sql.Rows
	var err error

	if eventType != "" {
		query = `
			SELECT id, event_type, user_id, username, client_ip, client_port, old_username, new_username, metadata, timestamp
			FROM user_events
			WHERE event_type = $1
			ORDER BY timestamp DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.QueryContext(ctx, query, eventType, limit, offset)
	} else {
		query = `
			SELECT id, event_type, user_id, username, client_ip, client_port, old_username, new_username, metadata, timestamp
			FROM user_events
			ORDER BY timestamp DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []UserEvent
	for rows.Next() {
		var event UserEvent
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&event.Username,
			&event.ClientIP,
			&event.ClientPort,
			&event.OldUsername,
			&event.NewUsername,
			&event.Metadata,
			&event.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// GetUserEventsByUsername получает события по имени пользователя
func (r *ComplianceRepository) GetUserEventsByUsername(ctx context.Context, username string, limit int) ([]UserEvent, error) {
	query := `
		SELECT id, event_type, user_id, username, client_ip, client_port, old_username, new_username, metadata, timestamp
		FROM user_events
		WHERE username = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, username, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []UserEvent
	for rows.Next() {
		var event UserEvent
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.UserID,
			&event.Username,
			&event.ClientIP,
			&event.ClientPort,
			&event.OldUsername,
			&event.NewUsername,
			&event.Metadata,
			&event.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// SaveSession сохраняет сессию пользователя (LOGIN)
func (r *ComplianceRepository) SaveSession(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (event_type, user_id, username, session_id, client_ip, client_port, user_agent, login_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		session.EventType,
		session.UserID,
		session.Username,
		session.SessionID,
		session.ClientIP,
		session.ClientPort,
		session.UserAgent,
		session.LoginTime,
	).Scan(&session.ID)

	if err != nil {
		log.Printf("Failed to save user session: %v", err)
		return err
	}

	return nil
}

// GetSessions получает все сессии
func (r *ComplianceRepository) GetSessions(ctx context.Context, eventType string, limit, offset int) ([]UserSession, error) {
	var query string
	var rows *sql.Rows
	var err error

	if eventType != "" {
		query = `
			SELECT id, event_type, user_id, username, session_id, client_ip, client_port, user_agent, login_time, logout_time, duration_seconds
			FROM user_sessions
			WHERE event_type = $1
			ORDER BY login_time DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.QueryContext(ctx, query, eventType, limit, offset)
	} else {
		query = `
			SELECT id, event_type, user_id, username, session_id, client_ip, client_port, user_agent, login_time, logout_time, duration_seconds
			FROM user_sessions
			ORDER BY login_time DESC
			LIMIT $1 OFFSET $2
		`
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		var logoutTime sql.NullTime
		var duration sql.NullInt64
		err := rows.Scan(
			&session.ID,
			&session.EventType,
			&session.UserID,
			&session.Username,
			&session.SessionID,
			&session.ClientIP,
			&session.ClientPort,
			&session.UserAgent,
			&session.LoginTime,
			&logoutTime,
			&duration,
		)
		if err != nil {
			return nil, err
		}
		if logoutTime.Valid {
			session.LogoutTime = logoutTime.Time
		}
		if duration.Valid {
			session.DurationSeconds = int(duration.Int64)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetActiveSessions получает активные сессии (без logout_time)
func (r *ComplianceRepository) GetActiveSessions(ctx context.Context) ([]UserSession, error) {
	query := `
		SELECT id, event_type, user_id, username, session_id, client_ip, client_port, user_agent, login_time, logout_time, duration_seconds
		FROM user_sessions
		WHERE logout_time IS NULL
		ORDER BY login_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		var logoutTime sql.NullTime
		var duration sql.NullInt64
		err := rows.Scan(
			&session.ID,
			&session.EventType,
			&session.UserID,
			&session.Username,
			&session.SessionID,
			&session.ClientIP,
			&session.ClientPort,
			&session.UserAgent,
			&session.LoginTime,
			&logoutTime,
			&duration,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetUserSessions получает сессии конкретного пользователя
func (r *ComplianceRepository) GetUserSessions(ctx context.Context, userID int, limit int) ([]UserSession, error) {
	query := `
		SELECT id, event_type, user_id, username, session_id, client_ip, client_port, user_agent, login_time, logout_time, duration_seconds
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY login_time DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		var logoutTime sql.NullTime
		var duration sql.NullInt64
		err := rows.Scan(
			&session.ID,
			&session.EventType,
			&session.UserID,
			&session.Username,
			&session.SessionID,
			&session.ClientIP,
			&session.ClientPort,
			&session.UserAgent,
			&session.LoginTime,
			&logoutTime,
			&duration,
		)
		if err != nil {
			return nil, err
		}
		if logoutTime.Valid {
			session.LogoutTime = logoutTime.Time
		}
		if duration.Valid {
			session.DurationSeconds = int(duration.Int64)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateLogout обновляет сессию при LOGOUT
func (r *ComplianceRepository) UpdateLogout(ctx context.Context, sessionID int) error {
	query := `
		UPDATE user_sessions
		SET logout_time = NOW(),
		    duration_seconds = EXTRACT(EPOCH FROM (NOW() - login_time))::INTEGER
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// UpdateSessionLogout обновляет сессию при LOGOUT/DISCONNECT по session_id
func (r *ComplianceRepository) UpdateSessionLogout(ctx context.Context, session *UserSession) error {
	query := `
		UPDATE user_sessions
		SET logout_time = NOW(),
		    duration_seconds = EXTRACT(EPOCH FROM (NOW() - login_time))::INTEGER
		WHERE user_id = $1 AND session_id = $2
	`
	result, err := r.db.ExecContext(ctx, query, session.UserID, session.SessionID)
	if err != nil {
		log.Printf("Failed to update session logout: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Если сессия не найдена, создаём новую запись (для DISCONNECT без предварительного LOGIN)
	if rowsAffected == 0 {
		log.Printf("Session not found for user=%d, session_id=%s, creating new record", session.UserID, session.SessionID)
		return r.SaveSession(ctx, session)
	}

	log.Printf("Updated %d session(s) for user=%d, session_id=%s", rowsAffected, session.UserID, session.SessionID)
	return nil
}
