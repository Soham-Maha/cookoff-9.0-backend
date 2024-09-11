package controllers

import (
    "time"
    "encoding/json"
    "net/http"
    "github.com/CodeChefVIT/cookoff-backend/internal/helpers/auth"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    refreshToken, ok := ctx.Value("refresh_token").(string)

    if !ok || refreshToken == "" {
        http.Error(w, "Refresh token is required", http.StatusBadRequest)
        return
    }

    UserID, err := auth.Tokens.GetUserID(ctx, refreshToken)
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

    cookie := &http.Cookie{
        Name:     "access_token",         
        Value:    newAccessToken,        
        Expires:  time.Now().Add(30* time.Minute), 
        HttpOnly: true,                   
        Secure:   true,                   
        Path:     "/",                   
    }
    http.SetCookie(w, cookie)
    response := map[string]string{
        "access_token": newAccessToken,
    }
    jsonResponse, _ := json.Marshal(response)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResponse)
}
