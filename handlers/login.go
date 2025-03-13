package handlers

import (
    "LunaTransfer/auth"
    "LunaTransfer/utils"
    "encoding/json"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Success bool   `json:"success"`
    Token   string `json:"token"`
    Role    string `json:"role"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if req.Username == "" || req.Password == "" {
        http.Error(w, "Username and password are required", http.StatusBadRequest)
        return
    }

    user, _, err := auth.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        utils.LogSystem("LOGIN_FAIL", req.Username, r.RemoteAddr, "Invalid credentials")
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, err := utils.GenerateJWT(user.Username, user.Role)
    if err != nil {
        utils.LogError("JWT_GENERATION_ERROR", err, req.Username)
        http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
        return
    }

    utils.LogSystem("LOGIN_SUCCESS", req.Username, r.RemoteAddr, "")

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{
        Success: true,
        Token:   token,
        Role:    user.Role,
    })
}