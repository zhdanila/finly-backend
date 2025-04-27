package security

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	userID := "12345"
	email := "user@example.com"

	token, err := GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Errorf("expected non-empty token")
	}

	claims, err := GetUserFromToken("Bearer " + token)
	if err != nil {
		t.Fatalf("expected no error while parsing token, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %v, got %v", email, claims.Email)
	}
}

func TestGetUserFromToken(t *testing.T) {
	userID := "12345"
	email := "user@example.com"

	token, err := GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := GetUserFromToken("Bearer " + token)
	if err != nil {
		t.Fatalf("expected no error while parsing token, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %v, got %v", email, claims.Email)
	}

	invalidToken := "Bearer invalid_token"
	_, err = GetUserFromToken(invalidToken)
	if err == nil {
		t.Errorf("expected error for invalid token, got nil")
	}
}

func TestVerify(t *testing.T) {
	userID := "12345"
	email := "user@example.com"

	token, err := GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	claims, err := Verify("Bearer " + token)
	if err != nil {
		t.Fatalf("expected no error while verifying token, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %v, got %v", email, claims.Email)
	}

	oldJWTSecret := jwtSecret
	jwtSecret = []byte("temporary_secret_key")
	defer func() {
		jwtSecret = oldJWTSecret
	}()

	// Create a token that has expired
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // expired
		},
	})
	signedString, err := expiredToken.SignedString(jwtSecret)

	_, err = Verify("Bearer " + signedString)
	if err == nil {
		t.Errorf("expected error for expired token, got nil")
	}
}
