package security

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestJWTService(t *testing.T, duration time.Duration) *JWTService {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return &JWTService{secretKey: key, accessTokenDuration: duration}
}

func TestGenerateTokenPair_ReturnsNonEmptyTokens(t *testing.T) {
	svc := newTestJWTService(t, time.Hour)
	resp, err := svc.GenerateTokenPair(1, "test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestGenerateTokenPair_AccessTokenContainsClaims(t *testing.T) {
	svc := newTestJWTService(t, time.Hour)
	resp, err := svc.GenerateTokenPair(42, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := jwt.ParseWithClaims(resp.AccessToken, &AccessClaims{}, func(_ *jwt.Token) (any, error) {
		return &svc.secretKey.PublicKey, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	claims := token.Claims.(*AccessClaims)
	if claims.UserID != 42 {
		t.Errorf("expected userID 42, got %d", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", claims.Email)
	}
}

func TestGenerateTokenPair_AccessTokenExpiry(t *testing.T) {
	duration := 15 * time.Minute
	svc := newTestJWTService(t, duration)
	resp, err := svc.GenerateTokenPair(1, "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := jwt.ParseWithClaims(resp.AccessToken, &AccessClaims{}, func(_ *jwt.Token) (any, error) {
		return &svc.secretKey.PublicKey, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	claims := token.Claims.(*AccessClaims)
	expectedExpiry := time.Now().Add(duration)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry).Abs()
	
	if diff > 2*time.Second {
		t.Errorf("expiry off by %v", diff)
	}
}

func TestNewJWTService_InvalidKeyPath(t *testing.T) {
	_, err := NewJWTService("/nonexistent/key.pem", time.Hour)
	if !os.IsNotExist(err) {
		t.Errorf("expected file not found error, got %v", err)
	}
}
