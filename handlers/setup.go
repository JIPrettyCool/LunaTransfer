package handlers

import (
    "LunaTransfer/auth"
    "encoding/json"
    "net/http"
    "os"
    "log"
    "fmt"
)

func SetupHandler(w http.ResponseWriter, r *http.Request) {
    setupCompleted, err := auth.IsSetupCompleted()
    if err != nil {
        http.Error(w, fmt.Sprintf("Error checking setup status: %s", err), http.StatusInternalServerError)
        return
    }

    if setupCompleted {
        http.Error(w, "Setup has already been completed", http.StatusBadRequest)
        return
    }

    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Email    string `json:"email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
        return
    }

    _, _, err = auth.CreateUser(req.Username, req.Password, req.Email, "admin")
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to create admin user: %s", err), http.StatusInternalServerError)
        return
    }

    user, token, err := auth.AuthenticateUser(req.Username, req.Password)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message": "Admin user created successfully, but token generation failed",
            "username": req.Username,
            "role": "admin",
            "error": err.Error(),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Setup completed successfully",
        "token": token,
        "username": req.Username,
        "role": user.Role,
    })
}

func SetupStatusHandler(w http.ResponseWriter, r *http.Request) {
    usersFile := "users.json"
    setupCompleted := false
    data, err := os.ReadFile(usersFile)
    if err == nil && len(data) > 10 {
        setupCompleted = true
    } else {
        setupCompleted, err = auth.IsSetupCompleted()
        if err != nil {
            log.Printf("Error checking setup status: %v", err)
            setupCompleted = false
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "setupCompleted": setupCompleted,
    })
}