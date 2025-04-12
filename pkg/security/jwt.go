package security

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const TokenTTL = 72 * time.Hour

// TODO: move to secret
var jwtSecret = []byte("your_secret_key")

func GenerateJWT(userID string, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     TokenTTL,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
