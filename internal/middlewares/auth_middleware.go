package middlewares

import (
	"context"
	"net/http"
	"strings"

	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
)

type ContextKey string

const UsernameKey ContextKey = "username"

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httphelpers.WriteError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := helpers.ValidateJWT(tokenString)
		if err != nil {
			httphelpers.WriteError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UsernameKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
