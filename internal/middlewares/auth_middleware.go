package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

type contextKey string

const ClaimsContextKey contextKey = "jwt_claims"

func JWTAuthMiddleware(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(r *http.Request) map[string]interface{} {
	if claims, ok := r.Context().Value(ClaimsContextKey).(map[string]interface{}); ok {
		return claims
	}
	return nil
}
