package security

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

const TokenTTL = 72 * time.Hour

// TODO: move to secret
var jwtSecret = []byte("your_secret_key")

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID string, email string) (string, error) {
	expirationTime := time.Now().Add(TokenTTL)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GetUserFromToken(tokenStr string) (*Claims, error) {
	parsedToken := strings.TrimPrefix(tokenStr, "Bearer ")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(parsedToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
