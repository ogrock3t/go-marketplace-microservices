package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/domain"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/dto"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailAlreadyExists = errors.New("email already exists")

var refreshTokenTTL = time.Hour * 24 * 30

type AuthService struct {
	userRepository repository.UserRepository
	hasher         repository.Hasher
	jwt            repository.TokenGenerator
}

func NewAuthService(repository repository.UserRepository, hasher repository.Hasher, jwt repository.TokenGenerator) *AuthService {
	return &AuthService{
		userRepository: repository,
		hasher:         hasher,
		jwt:            jwt,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	passwordHash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userRepository.CreateUser(ctx, user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}

	response, err := s.jwt.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	if err := s.userRepository.SaveRefreshToken(
		ctx,
		user.ID,
		user.Email,
		response.RefreshToken,
		time.Now().Add(refreshTokenTTL)); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if err := s.hasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	response, err := s.jwt.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	if err := s.userRepository.SaveRefreshToken(
		ctx,
		user.ID,
		user.Email,
		response.RefreshToken,
		time.Now().Add(refreshTokenTTL)); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	rt, err := s.userRepository.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidCredentials
	}

	if err = s.userRepository.DeleteRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

	response, err := s.jwt.GenerateTokenPair(rt.UserID, rt.Email)
	if err != nil {
		return nil, err
	}

	err = s.userRepository.SaveRefreshToken(
		ctx,
		rt.UserID,
		rt.Email,
		response.RefreshToken,
		time.Now().Add(refreshTokenTTL))
	if err != nil {
		return nil, err
	}

	return response, nil
}
