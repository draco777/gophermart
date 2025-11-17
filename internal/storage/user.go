package storage

import (
	"database/sql"
	"fmt"

	"github.com/draco777/gophermart/internal/models"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) CreateUser(login, passwordHash string) (*models.User, error) {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`

	var user models.User
	err := s.db.QueryRow(query, login, passwordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.Login = login
	user.PasswordHash = passwordHash

	// Создаем баланс для пользователя
	_, err = s.db.Exec(`INSERT INTO balances (user_id) VALUES ($1)`, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance for user: %w", err)
	}

	return &user, nil
}

func (s *UserStorage) GetUserByLogin(login string) (*models.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	var user models.User
	err := s.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *UserStorage) UserExists(login string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE login = $1`

	var count int
	err := s.db.QueryRow(query, login).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}
