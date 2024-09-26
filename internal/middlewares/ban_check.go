package middlewares

import (
    "net/http"
    "github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
    "github.com/google/uuid"
)

func BanCheckMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userIDStr, ok := r.Context().Value("user_id").(string)
        if !ok {
            http.Error(w, "user_id not found in context", http.StatusUnauthorized)
            return
        }
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            http.Error(w, "Invalid user ID format", http.StatusBadRequest)
            return
        }
        ctx := r.Context()
        user, err := database.Queries.GetUserById(ctx, userID)
        if err != nil {
            http.Error(w, "Unable to fetch user data", http.StatusInternalServerError)
            return
        }

        if user.IsBanned {
            http.Error(w, "Your account is banned", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
