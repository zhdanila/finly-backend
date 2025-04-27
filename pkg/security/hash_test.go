package security

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if password == hashedPassword {
		t.Errorf("expected hashed password to be different from the original")
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		t.Fatalf("expected passwords to match, got error %v", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testPassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error while hashing password, got %v", err)
	}

	if !CheckPasswordHash(password, hashedPassword) {
		t.Errorf("expected password to match hashed password")
	}

	incorrectPassword := "wrongPassword"
	if CheckPasswordHash(incorrectPassword, hashedPassword) {
		t.Errorf("expected incorrect password to not match")
	}
}
