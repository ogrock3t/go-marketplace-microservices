package domain

import (
	"fmt"
	"time"
)

var (
	ErrUserNotFound         = fmt.Errorf("user not found")
	ErrRefreshTokenNotFound = fmt.Errorf("refresh token not found")
)

type User struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	Token     string
	UserID    int64
	Email     string
	ExpiresAt time.Time
}
