package middlewares

import (
    "context"
    "net/http"
    "github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
    "github.com/google/uuid"
)

func BanCheckMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userIDStr := r.Context().Value("user_id").(string)
        userID, err := uuid.Parse(userIDStr)
        if err != nil {
            http.Error(w, "Invalid user ID format", http.StatusBadRequest)
            return
        }

        if isUserBanned(userID) {
            http.Error(w, "Your account is banned", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func isUserBanned(userID uuid.UUID) bool {
    ctx := context.Background()
    user, err := database.Queries.GetUserById(ctx, userID)
    if err != nil {
        return true
    }
    return user.IsBanned
}