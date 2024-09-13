package auth

import (
	"os"
	"time"

	"github.com/go-chi/jwtauth/v5"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
)

var TokenAuth *jwtauth.JWTAuth

func InitJWT() {
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		panic("JWT_KEY environment variable not set")
	}
	TokenAuth = jwtauth.New("HS256", []byte(jwtKey), nil)
}

func GenerateJWT(user *db.User, isRefresh bool) (string, time.Time, error) {
	var expirationTime time.Time

	if !isRefresh {
		expirationTime = time.Now().Add(time.Hour / 2)
		_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
			"username": user.Name,
			"role":     user.Role,
			"type":     "access",
			"exp":      expirationTime.Unix(),
		})
		return tokenString, expirationTime, err
	} else {
		expirationTime = time.Now().Add(time.Hour * 2)
		_, tokenString, err := TokenAuth.Encode(map[string]interface{}{
			"username": user.Name,
			"type":     "refresh",
			"exp":      expirationTime.Unix(),
		})
		return tokenString, expirationTime, err
	}

}
