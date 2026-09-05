package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/domain"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/dto"
)

type mockRepository struct {
	user         *domain.User
	err          error
	refreshToken *domain.RefreshToken
}

func (m *mockRepository) CreateUser(_ context.Context, user *domain.User) error {
	return m.err
}

func (m *mockRepository) GetUserByEmail(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func (m *mockRepository) SaveRefreshToken(_ context.Context, _ int64, _, _ string, _ time.Time) error {
	return nil
}

func (m *mockRepository) GetRefreshToken(_ context.Context, _ string) (*domain.RefreshToken, error) {
	return m.refreshToken, m.err
}

func (m *mockRepository) DeleteRefreshToken(_ context.Context, _ string) error {
	return nil
}

type mockHasher struct {
	compareErr error
	hashErr    error
}

func (m *mockHasher) Hash(p string) (string, error) {
	return "hashed:" + p, m.hashErr
}

func (m *mockHasher) Compare(_, _ string) error {
	return m.compareErr
}

type mockTokenGenerator struct{}

func (m *mockTokenGenerator) GenerateTokenPair(userID int64, email string) (*dto.AuthResponse, error) {
	return &dto.AuthResponse{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}, nil
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		err: &pgconn.PgError{Code: "23505"}},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "test@gmail.com",
		Password: "passwordHash",
	})

	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(&mockRepository{},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
		FirstName: "John",
		LastName:  "Test",
		Email:     "test@gmail.com",
		Password:  "passwordHash",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestRregister_HashError(t *testing.T) {
	svc := NewAuthService(&mockRepository{},
		&mockHasher{hashErr: errors.New("hash error")},
		&mockTokenGenerator{},
	)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		FirstName: "John",
		LastName:  "Test",
		Email:     "test@gmail.com",
		Password:  "passwordHash",
	})

	if err == nil || err.Error() != "hash error" {
		t.Fatalf("expected hash error, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		err: domain.ErrUserNotFound},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "test@gmail.com",
		Password: "password",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		user: &domain.User{
			ID:           1,
			Email:        "test@gmail.com",
			PasswordHash: "passwordHash"}},
		&mockHasher{compareErr: errors.New("password mismatch")},
		&mockTokenGenerator{},
	)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "test@gmail.com",
		Password: "wrongPassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		user: &domain.User{
			ID:           1,
			Email:        "test@gmail.com",
			PasswordHash: "passwordHash"}},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "test@gmail.com",
		Password: "passwordHash",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		refreshToken: &domain.RefreshToken{
			Token:     "validRefresh",
			UserID:    1,
			Email:     "test@gmail.com",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	resp, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "validRefresh",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := NewAuthService(
		&mockRepository{err: errors.New("refresh token not found")},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	_, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "notvalidRefresh",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	svc := NewAuthService(&mockRepository{
		refreshToken: &domain.RefreshToken{
			Token:     "validRefresh",
			UserID:    1,
			Email:     "test@gmail.com",
			ExpiresAt: time.Now(),
		}},
		&mockHasher{},
		&mockTokenGenerator{},
	)

	_, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{
		RefreshToken: "validRefresh",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
