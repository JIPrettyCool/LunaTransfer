package handlers

import (
    "LunaTransfer/utils"
    "encoding/json"
    "net/http"
    "strings"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(w, "Authorization header required", http.StatusUnauthorized)
        return
    }

    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        http.Error(w, "Authorization header must be in format: Bearer {token}", http.StatusBadRequest)
        return
    }

    tokenString := parts[1]
    
    // Get claims to find expiration time
    claims, err := utils.ValidateJWT(tokenString)
    if err == nil {
        // Only blacklist valid tokens
        expiryTime := claims.ExpiresAt.Time
        utils.BlacklistToken(tokenString, expiryTime)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}