package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenPayload struct {
	UserID int
	Email  string
	Exp    time.Time
}

func GenerateJWTToken(secret string, payload JWTTokenPayload) (string, error) {
	claims := jwt.MapClaims{
		"user_id": payload.UserID,
		"email":   payload.Email,
		"exp":     payload.Exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
