package controllers

import (
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/google/uuid"
)

func MeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.GetUserID(w, r)
	if err != nil {
		httphelpers.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := database.Queries.GetUserById(r.Context(), id)
	if err != nil {
		logger.Errof("Failed to get user: %v", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	submissions, err := database.Queries.GetSubmissionsWithRoundByUserId(r.Context(), uuid.NullUUID{UUID: id, Valid: true})
	if err != nil {
		logger.Errof("Failed to get submissions: %v", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to get submissions")
		return
	}

	var submissionsByRound []db.GetSubmissionsWithRoundByUserIdRow

	for _, submission := range submissions {
		submissionsByRound = append(submissionsByRound, db.GetSubmissionsWithRoundByUserIdRow(submission))
	}

	data := map[string]any{
		"username":    user.Name,
		"round":       user.RoundQualified,
		"score":       user.Score.Int32,
		"submissions": submissionsByRound,
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "User details fetched successfully",
		"data":    data,
	})
}
