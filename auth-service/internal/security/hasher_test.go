package security

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHash_ReturnsValidBcriptHash(t *testing.T) {
	h := NewBcryptHasher(bcrypt.MinCost)

	hash, err := h.Hash("password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("password")); err != nil {
		t.Fatalf("expected valid bcrypt hash, got error: %v", err)
	}
}

func TestCompare_ReturnsNilForMatchingPassword(t *testing.T) {
	h := NewBcryptHasher(bcrypt.MinCost)
	
	hash, err := h.Hash("password")	
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if err := h.Compare(hash, "password"); err != nil {
		t.Fatalf("expected no error for matching password, got %v", err)
	}
}

func TestCompare_ReturnsErrorForNonMatchingPassword(t *testing.T) {
	h := NewBcryptHasher(bcrypt.MinCost)
	
	hash, err := h.Hash("password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if err := h.Compare(hash, "wrongPassword"); err == nil {
		t.Fatalf("expected error for non-matching password, got nil")
	}
}