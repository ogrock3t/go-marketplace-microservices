package security

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/dto"
)

type JWTService struct {
	secretKey           *rsa.PrivateKey
	accessTokenDuration time.Duration
}

type AccessClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKeyPath string, accessTokenDuration time.Duration) (*JWTService, error) {
	keyBytes, err := os.ReadFile(secretKeyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, err
	}

	return &JWTService{
		secretKey:           privateKey,
		accessTokenDuration: accessTokenDuration,
	}, nil
}

func (s *JWTService) GenerateTokenPair(userID int64, email string) (*dto.AuthResponse, error) {
	accessToken, err := s.generateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.accessTokenDuration),
	}, nil
}

func (s *JWTService) generateAccessToken(userID int64, email string) (string, error) {
	now := time.Now()

	claims := AccessClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenDuration)),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
