package handlers

import (
    "LunaMFT/auth"
    "encoding/json"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Username string `json:"username"`
    APIKey   string `json:"apiKey"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var loginReq LoginRequest
    
    if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if loginReq.Username == "" || loginReq.Password == "" {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }
    
    user, err := auth.AuthUser(loginReq.Username, loginReq.Password)
    if err != nil {
        switch err {
        case auth.ErrUserNotFound, auth.ErrInvalidPassword:
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        default:
            http.Error(w, "Authentication failed", http.StatusInternalServerError)
        }
        return
    }
    
    response := LoginResponse{
        Username: user.Username,
        APIKey:   user.APIKey,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}