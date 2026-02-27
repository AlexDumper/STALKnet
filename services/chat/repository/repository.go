package repository

import "database/sql"

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// GetRooms возвращает все комнаты
func (r *ChatRepository) GetRooms() ([]Room, error) {
	query := `SELECT id, name, description, created_by, is_private, created_at FROM rooms`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatedBy, &room.IsPrivate, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// CreateRoom создаёт новую комнату
func (r *ChatRepository) CreateRoom(name, description string, createdBy int, isPrivate bool) (int, error) {
	var id int
	query := `INSERT INTO rooms (name, description, created_by, is_private) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(query, name, description, createdBy, isPrivate).Scan(&id)
	return id, err
}

// GetMessages возвращает сообщения комнаты
func (r *ChatRepository) GetMessages(roomID int, limit, offset int) ([]Message, error) {
	query := `
		SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// SaveMessage сохраняет сообщение
func (r *ChatRepository) SaveMessage(roomID, userID int, content string) (int, error) {
	var id int
	query := `INSERT INTO messages (room_id, user_id, content) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(query, roomID, userID, content).Scan(&id)
	return id, err
}

// Room модель
type Room struct {
	ID          int
	Name        string
	Description string
	CreatedBy   int
	IsPrivate   bool
}

// Message модель
type Message struct {
	ID       int
	RoomID   int
	UserID   int
	Username string
	Content  string
}
