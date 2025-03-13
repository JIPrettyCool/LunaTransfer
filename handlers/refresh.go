package handlers

import (
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "encoding/json"
    "net/http"
)

type RefreshResponse struct {
    Success bool   `json:"success"`
    Token   string `json:"token"`
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
    // Get username from context (set by AuthMiddleware)
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Invalid session", http.StatusUnauthorized)
        return
    }
    
    role, _ := common.GetRoleFromContext(r.Context())
    
    // Generate a new token
    token, err := utils.GenerateJWT(username, role)
    if err != nil {
        utils.LogError("JWT_REFRESH_ERROR", err, username)
        http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
        return
    }

    utils.LogSystem("TOKEN_REFRESH", username, r.RemoteAddr, "")

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(RefreshResponse{
        Success: true,
        Token:   token,
    })
}