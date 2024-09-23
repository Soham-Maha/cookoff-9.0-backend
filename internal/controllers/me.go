package controllers

import (
	"net/http"

	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
)

func MeHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := helpers.GetUserID(w, r)

	user, err := database.Queries.GetUserById(r.Context(), id)
	if err != nil {
		logger.Errof("Failed to get user: %v", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	data := map[string]any{
		"username": user.Name,
		"round":    user.RoundQualified,
		"score":    user.Score,
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "User details fetched successfully",
		"user":    data,
	})
}
