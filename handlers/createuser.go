package handlers

import (
    "LunaMFT/auth"
    "encoding/json"
    "net/http"
)

type UserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var userReq UserRequest
    
    if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    
    if userReq.Username == "" || userReq.Password == "" {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }
    
    user, err := auth.CreateUser(userReq.Username, userReq.Password)
    if err != nil {
        switch err {
        case auth.ErrUserExists:
            http.Error(w, "User already exists", http.StatusConflict)
        case auth.ErrInternalServer:
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        default:
            http.Error(w, "Registration failed", http.StatusBadRequest)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "username": user.Username,
        "apiKey": user.APIKey,
    })
}