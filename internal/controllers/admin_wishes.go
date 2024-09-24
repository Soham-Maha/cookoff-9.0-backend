package controllers

import (
	"context"
	"net/http"
	"encoding/json"
	"log"
    "github.com/google/uuid"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	"github.com/go-chi/render"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	users, err := database.Queries.GetAllUsers(ctx)
	if err != nil {
		http.Error(w, "Unable to fetch users", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, users)
}
func UpgradeUserToRound(w http.ResponseWriter, r *http.Request) {
    var requestBody map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    userIDsInterface, ok := requestBody["user_ids"].([]interface{})
    if !ok || len(userIDsInterface) == 0 {
        http.Error(w, "Invalid user_ids format", http.StatusBadRequest)
        return
    }
    userIDs := make([]uuid.UUID, len(userIDsInterface))
    for i, idStr := range userIDsInterface {
        id, err := uuid.Parse(idStr.(string))
        if err != nil {
            http.Error(w, "Invalid user_id", http.StatusBadRequest)
            return
        }
        userIDs[i] = id
    }

    roundFloat, ok := requestBody["round"].(float64)
    if !ok {
        http.Error(w, "Invalid round format", http.StatusBadRequest)
        return
    }
    round := int(roundFloat)

    ctx := context.Background()
    err := database.Queries.UpgradeUserToRound(ctx, db.UpgradeUserToRoundParams{
        Column1:        userIDs,
        RoundQualified: int32(round),
    })
    if err != nil {
        http.Error(w, "Unable to upgrade users to round", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
func BanUser(w http.ResponseWriter, r *http.Request) {
    var requestBody map[string]string
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    userIDStr, ok := requestBody["user_id"]
    if !ok {
        http.Error(w, "user_id not provided", http.StatusBadRequest)
        return
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        http.Error(w, "Invalid user_id", http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    err = database.Queries.BanUser(ctx, userID)
    if err != nil {
        http.Error(w, "Unable to ban user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
func UnbanUser(w http.ResponseWriter, r *http.Request) {
    var requestBody map[string]string
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    userIDStr, ok := requestBody["user_id"]
    if !ok {
        http.Error(w, "user_id not provided", http.StatusBadRequest)
        return
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        http.Error(w, "Invalid user_id", http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    err = database.Queries.UnbanUser(ctx, userID)
    if err != nil {
        http.Error(w, "Unable to unban user", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
type RoundRequest struct {
    RoundID string `json:"round_id"`
    Enabled bool   `json:"enabled"`
}
func SetRoundStatus(w http.ResponseWriter, r *http.Request) {
    var reqBody RoundRequest
    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    status := "disabled"
    if reqBody.Enabled {
        status = "enabled"
    }

    err := RedisClient.Set(ctx, reqBody.RoundID, status, 0).Err()
    if err != nil {
        log.Printf("Failed to update round status: %v\n", err)
        http.Error(w, "Failed to update round status", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Round status updated successfully"))
}