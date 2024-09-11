package controllers

import (
	"net/http"

	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := httphelpers.ParseJSON(r, &req); err != nil {
		logger.Errof("Invalid request payload: %v", err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid request payload")
		return
	}

	user, err := database.Queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		logger.Errof("User not found or invalid credentials for email: %s, err: %v", req.Email, err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if user.Password != req.Password {
		logger.Errof("Invalid password for user: %s", user.Email)
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenString, expirationTime, err := helpers.GenerateJWT(&user)
	if err != nil {
		logger.Errof("Failed to generate JWT for user: %s, err: %v", user.Email, err)
		httphelpers.WriteError(w, http.StatusUnauthorized, "failed to generate token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	httphelpers.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Login successful",
	})
}
