package utils

import (
	"main/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, email string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JwtSecret))
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(config.JwtSecret), nil
	})
}
