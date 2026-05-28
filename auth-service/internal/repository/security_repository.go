package repository

import (
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/dto"
)

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type TokenGenerator interface {
	GenerateTokenPair(userID int64, email string) (*dto.AuthResponse, error)
}
