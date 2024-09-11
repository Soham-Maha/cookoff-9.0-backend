package middleware

import (
    "net/http"
    "strings"

    "github.com/go-chi/render"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
)

func VerifyRefreshTokenMiddleware(tokenManager *auth.AuthTokenManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                render.JSON(w, r, map[string]string{"error": "Missing Authorization header"})
                w.WriteHeader(http.StatusUnauthorized)
                return
            }
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                render.JSON(w, r, map[string]string{"error": "Invalid token format"})
                w.WriteHeader(http.StatusUnauthorized)
                return
            }

            refreshToken := parts[1]

            isValid, err := tokenManager.VerifyRefreshToken(refreshToken)
            if err != nil || !isValid {
                render.JSON(w, r, map[string]string{"error": "Invalid or expired refresh token"})
                w.WriteHeader(http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
