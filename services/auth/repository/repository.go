package repository

import (
	"database/sql"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser создаёт нового пользователя
func (r *AuthRepository) CreateUser(username, passwordHash, email string) (int, error) {
	var id int
	query := `INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(query, username, passwordHash, email).Scan(&id)
	return id, err
}

// GetUserByUsername находит пользователя по имени
func (r *AuthRepository) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password_hash, email, status FROM users WHERE username = $1`
	row := r.db.QueryRow(query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Status)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// User модель для репозитория
type User struct {
	ID       int
	Username string
	Password string
	Email    string
	Status   string
}
