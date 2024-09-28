package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/CodeChefVIT/cookoff-backend/internal/db"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	helpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/database"
	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/CodeChefVIT/cookoff-backend/internal/helpers/validator"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
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
		logger.Infof("received error from database %v", err.Error())
		httphelpers.WriteError(w, http.StatusInternalServerError, "some error occurred")
		return
	}

	loggedin, err := auth.RefreshTokenExist(r.Context(), user.ID.String())
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Error while checking cache")
		logger.Errof(fmt.Sprintf("Error while checking cache %v", err.Error()))
		return
	}

	if !loggedin {
		httphelpers.WriteError(w, http.StatusForbidden, "Someone already logged in")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			httphelpers.WriteError(w, http.StatusConflict, "invalid password")
			return
		}
		httphelpers.WriteError(w, http.StatusInternalServerError, err.Error())
	}

	accessToken, err := auth.GenerateJWT(&user, false)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := auth.GenerateJWT(&user, true)
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	expiration := (time.Hour + 25*time.Minute)
	err = database.RedisClient.Set(r.Context(), user.ID.String(), refreshToken, expiration).Err()
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "failed to set token in cache")
		logger.Errof(fmt.Sprintf("failed to set token in cache %v", err.Error()))
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
		"data":    data,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	jwt, err := r.Cookie("jwt")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		httphelpers.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	refresh, err := r.Cookie("refresh_token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		httphelpers.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	claims, err := jwtauth.VerifyToken(helpers.TokenAuth, jwt.Value)
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

	err = database.RedisClient.Del(r.Context(), userId).Err()
	if err != nil {
		httphelpers.WriteError(w, http.StatusInternalServerError, "Error remove token from cache")
		return
	}

	if jwt != nil {
		jwt.MaxAge = -1
		jwt.Value = ""
		jwt.Expires = time.Now()
		jwt.Path = "/"
		jwt.SameSite = http.SameSiteNoneMode
		jwt.Secure = true
		jwt.HttpOnly = true
		http.SetCookie(w, jwt)
	}

	if refresh != nil {
		refresh.MaxAge = -1
		refresh.Value = ""
		refresh.Expires = time.Now()
		refresh.Path = "/"
		refresh.SameSite = http.SameSiteNoneMode
		refresh.Secure = true
		refresh.HttpOnly = true
		http.SetCookie(w, refresh)
	}

	httphelpers.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "logged out successfully",
	})
}
