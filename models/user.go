package models

import "time"

// User represents a user that can log in and get a JWT token.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginInput holds the credentials sent to POST /login.
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
