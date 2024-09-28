package controllers

import (
	"errors"
	"net/http"

	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		logger.Errof("Refresh token not found: %v", err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "refresh token not found")
		return
	}

	claims, err := jwtauth.VerifyToken(helpers.TokenAuth, cookie.Value)
	if err != nil || claims == nil {
		logger.Errof("Invalid refresh token: %v", err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid refresh token: "+err.Error())
		return
	}

	userId, ok := claims.PrivateClaims()["user_id"].(string)
	if !ok {
		logger.Errof("Invalid token claims, user_id not found")
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid token claims")
		return
	}

	check, err := auth.CheckRefreshToken(r.Context(), userId, cookie.Value)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !check {
		httphelpers.WriteError(w, http.StatusUnauthorized, "Token does not match with cache token")
		return
	}

	userIdUUID, err := uuid.Parse(userId)
	if err != nil {
		logger.Errof("Invalid user_id: %s, err: %v", userId, err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid user_id")
		return
	}

	user, err := database.Queries.GetUserById(r.Context(), userIdUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httphelpers.WriteError(w, http.StatusUnauthorized, "user not found")
			return
		}
		logger.Errof("User not found: %s, err: %v", user.Name, err)
		return
	}

	accessToken, err := helpers.GenerateJWT(&user, false)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    accessToken,
		MaxAge:   1000,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})

	httphelpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Token refreshed",
	})
}
