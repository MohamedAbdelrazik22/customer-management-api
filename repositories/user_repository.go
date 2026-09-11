package repositories

import (
	"database/sql"
	"errors"
)

// ErrInvalidCredentials is returned when username or password is wrong.
var ErrInvalidCredentials = errors.New("invalid username or password")

// UserRepository handles database operations related to users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetHashedPassword returns the bcrypt-hashed password for the given username.
// Returns ErrInvalidCredentials if the username does not exist.
func (r *UserRepository) GetHashedPassword(username string) (string, error) {
	var hashedPassword string

	err := r.db.QueryRow(
		"SELECT password FROM users WHERE username = ?",
		username,
	).Scan(&hashedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	return hashedPassword, nil
}
