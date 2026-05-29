package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/domain"
)

type UserStorage struct {
	connect *Connection
}

func NewUserStorage(connect *Connection) *UserStorage {
	return &UserStorage{connect: connect}
}

func (s *UserStorage) CreateUser(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := s.connect.pool.QueryRow(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user on postgresql: %w", err)
	}

	return nil
}

func (s *UserStorage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, first_name, last_name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User

	err := s.connect.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	return &user, nil
}

func (s *UserStorage) SaveRefreshToken(ctx context.Context, userID int64, email, token string, expiresAt time.Time) error {
	const query = `
		INSERT INTO refresh_tokens (token, user_id, email, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.connect.pool.Exec(ctx, query, token, userID, email, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token on postgresql: %w", err)
	}

	return nil
}

func (s *UserStorage) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	const query = `
		SELECT token, user_id, email, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`

	var refreshToken domain.RefreshToken
	err := s.connect.pool.QueryRow(ctx, query, token).Scan(
		&refreshToken.Token,
		&refreshToken.UserID,
		&refreshToken.Email,
		&refreshToken.ExpiresAt,
	)
	if err != nil {
		return nil, domain.ErrRefreshTokenNotFound
	}

	return &refreshToken, nil
}

func (s *UserStorage) DeleteRefreshToken(ctx context.Context, token string) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE token = $1
	`

	_, err := s.connect.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token on postgresql: %w", err)
	}
	
	return nil
}
