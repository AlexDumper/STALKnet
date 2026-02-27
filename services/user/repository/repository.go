package repository

import "database/sql"

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID находит пользователя по ID
func (r *UserRepository) GetUserByID(id int) (*User, error) {
	query := `SELECT id, username, email, status, created_at, last_seen FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Status, &user.CreatedAt, &user.LastSeen)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateStatus обновляет статус пользователя
func (r *UserRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE users SET status = $1, last_seen = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

// GetOnlineUsers возвращает список онлайн пользователей
func (r *UserRepository) GetOnlineUsers() ([]User, error) {
	query := `SELECT id, username, email, status, created_at, last_seen FROM users WHERE status = 'online'`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Status, &user.CreatedAt, &user.LastSeen)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// User модель
type User struct {
	ID       int
	Username string
	Email    string
	Status   string
}
