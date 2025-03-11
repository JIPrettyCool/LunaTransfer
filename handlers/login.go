package handlers

import (
    "LunaMFT/auth"
    "encoding/json"
    "net/http"
    "LunaMFT/utils"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Success bool   `json:"success"`
    ApiKey  string `json:"apiKey"`
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

    user, apiKey, err := auth.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        // Log the failed login attempt
        utils.LogSystem("LOGIN_FAIL", req.Username, r.RemoteAddr, "Invalid credentials")
        http.Error(w, "Invalid username or password", http.StatusUnauthorized)
        return
    }

    // Log successful login
    utils.LogSystem("LOGIN_SUCCESS", req.Username, r.RemoteAddr)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{
        Success: true,
        ApiKey:  apiKey,
        Role:    user.Role,
    })
}