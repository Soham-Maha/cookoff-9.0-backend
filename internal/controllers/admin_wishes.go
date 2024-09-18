package controllers

import (
	"context"
	"net/http"
	"encoding/json"
    "github.com/google/uuid"
	"github.com/go-chi/chi/v5"
    "strconv"
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
	if !ok {
		http.Error(w, "Invalid user_ids format", http.StatusBadRequest)
		return
	}

	var userIDs []uuid.UUID
	for _, idStr := range userIDsInterface {
		id, err := uuid.Parse(idStr.(string))
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, id)
	}
	roundFloat, ok := requestBody["round"].(float64)
	if !ok {
		http.Error(w, "Invalid round format", http.StatusBadRequest)
		return
	}
	round := int(roundFloat)
	ctx := context.Background()
	err := database.Queries.UpgradeUserToRound(ctx, userIDs, round)
	if err != nil {
		http.Error(w, "Unable to upgrade users to round", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func BanUsers(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	userIDsInterface, ok := requestBody["user_ids"].([]interface{})
	if !ok {
		http.Error(w, "Invalid user_ids format", http.StatusBadRequest)
		return
	}

	var userIDs []uuid.UUID
	for _, idStr := range userIDsInterface {
		id, err := uuid.Parse(idStr.(string))
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, id)
	}
	ctx := context.Background()
	err := database.Queries.BanUsers(ctx, userIDs)
	if err != nil {
		http.Error(w, "Unable to ban users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func UnbanUsers(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	userIDsInterface, ok := requestBody["user_ids"].([]interface{})
	if !ok {
		http.Error(w, "Invalid user_ids format", http.StatusBadRequest)
		return
	}

	var userIDs []uuid.UUID
	for _, idStr := range userIDsInterface {
		id, err := uuid.Parse(idStr.(string))
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, id)
	}
	ctx := context.Background()
	err := database.Queries.UnbanUsers(ctx, userIDs)
	if err != nil {
		http.Error(w, "Unable to unban users", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func EnableRound(w http.ResponseWriter, r *http.Request) {
	roundNumberStr := chi.URLParam(r, "round_number")
	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = database.Queries.EnableRound(ctx, roundNumber)
	if err != nil {
		http.Error(w, "Unable to enable round", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func DisableRound(w http.ResponseWriter, r *http.Request) {
	roundNumberStr := chi.URLParam(r, "round_number")
	roundNumber, err := strconv.Atoi(roundNumberStr)
	if err != nil {
		http.Error(w, "Invalid round number", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = database.Queries.DisableRound(ctx, roundNumber)
	if err != nil {
		http.Error(w, "Unable to disable round", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}