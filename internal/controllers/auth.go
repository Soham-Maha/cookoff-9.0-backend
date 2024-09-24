package controllers

import (
	"errors"
	"net/http"
	"os"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignupRequest struct {
	Email string `json:"email"    validate:"required"`
	Name  string `json:"name"     validate:"required"`
	RegNo string `json:"reg_no"   validate:"required"`
	Key   string `json:"fuck_you" validate:"required"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var payload SignupRequest

	if err := httphelpers.ParseJSON(r, &payload); err != nil {
		httphelpers.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validator.ValidatePayload(w, payload); err != nil {
		httphelpers.WriteError(w, http.StatusNotAcceptable, "invalid input")
		return
	}

	if payload.Key != os.Getenv("SECRET_KEY_FUCKERS") {
		httphelpers.WriteError(w, http.StatusUnauthorized, "I WILL POP YOUR CHERRY BRO")
		return
	}

	passwd := auth.PasswordGenerator(6)
	hashed, err := bcrypt.GenerateFromPassword([]byte(passwd), 10)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := uuid.NewV7()
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = database.Queries.CreateUser(r.Context(), db.CreateUserParams{
		ID:             id,
		Email:          payload.Email,
		RegNo:          payload.RegNo,
		Password:       string(hashed),
		Role:           "user",
		RoundQualified: 0,
		Score:          pgtype.Int4{},
		Name:           payload.Name,
	})
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]any{
		"message":  "user added",
		"email":    payload.Email,
		"password": passwd,
	})
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
		if errors.Is(err, pgx.ErrNoRows) {
			httphelpers.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		httphelpers.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			httphelpers.WriteError(w, http.StatusBadRequest, "invalid password")
			return
		}
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
	}

	accessToken, err := auth.GenerateJWT(&user, false)
	if err != nil {
		logger.Errof("Failed to generate access token for user: %s, err: %v", user.Email, err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateJWT(&user, true)
	if err != nil {
		logger.Errof("Failed to generate refresh token for user: %s, err: %v", user.Email, err)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
	})

	data := map[string]any{
		"username": user.Name,
		"round":    user.RoundQualified,
		"score":    user.Score,
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Login successful",
		"user":    data,
	})
}
