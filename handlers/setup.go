package handlers

import (
	"LunaTransfer/auth"
	"encoding/json"
	"net/http"
	"errors"
)

func SetupHandler(w http.ResponseWriter, r *http.Request) {
    users, err := auth.LoadUsers()
    if err != nil && !errors.Is(err, auth.ErrUsersFileNotFound) {
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    
    if len(users) > 0 {
        http.Error(w, "Setup already completed", http.StatusForbidden)
        return
    }
    
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    user, apiKey, err := auth.CreateUser(req.Username, req.Password, req.Email, "admin")
    if err != nil {
        http.Error(w, "Failed to create admin", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Initial setup completed - admin user created",
        "apiKey": apiKey,
		"username": user.Username,
		"role": user.Role,
		"email": user.Email,
    })
}