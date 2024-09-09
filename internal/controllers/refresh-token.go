package controllers

import (
    "encoding/json"
    "net/http"
    "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
    RefreshToken := r.FormValue("refresh_token")

    if RefreshToken == "" {
        http.Error(w, "Refresh token is required", http.StatusBadRequest)
        return
    }

    ctx := r.Context()

    UserID, err := auth.Tokens.GetUserID(ctx, RefreshToken)
    if err != nil {
        http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
        return
    }

    // Generate a new access token
    newAccessToken, err := auth.Tokens.GenerateAccessToken(UserID)
    if err != nil {
        http.Error(w, "Failed to generate new access token", http.StatusInternalServerError)
        return
    }

    oldAccessToken := r.Header.Get("Authorization") 
    err = auth.Tokens.DeleteAccessToken(ctx, oldAccessToken)
    if err != nil {
        http.Error(w, "Failed to delete old access token", http.StatusInternalServerError)
        return
    }

    // Save the new access token
    err = auth.Tokens.AddAccessToken(ctx, newAccessToken, UserID)
    if err != nil {
        http.Error(w, "Failed to save new access token", http.StatusInternalServerError)
        return
    }

    // Return the new access token
    response := map[string]string{
        "access_token": newAccessToken,
    }
    jsonResponse, _ := json.Marshal(response)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResponse)
}
