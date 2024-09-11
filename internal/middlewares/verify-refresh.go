package middleware

import (
    "net/http"

    "github.com/go-chi/render"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
)

func VerifyRefreshTokenMiddleware(tokenManager *auth.AuthTokenManager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("refresh_token")
            if err != nil {
                if err == http.ErrNoCookie {
                    render.JSON(w, r, map[string]string{"error": "Missing refresh token cookie"})
                    w.WriteHeader(http.StatusUnauthorized)
                    return
                }
                render.JSON(w, r, map[string]string{"error": "Error retrieving refresh token cookie"})
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            refreshToken := cookie.Value
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
