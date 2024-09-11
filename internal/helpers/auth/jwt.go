package auth

import (
	"github.com/go-chi/jwtauth/v5"
	"os"
	"time"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
)

var TokenAuth *jwtauth.JWTAuth

func InitJWT() {
	TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_KEY")), nil)
}

func GenerateJWT(user *db.User) (string, time.Time, error) {
	expirationTime := time.Now().Add(time.Hour / 2)

	_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
		"username": user.Name,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
	})
	return tokenString, expirationTime, err
}
