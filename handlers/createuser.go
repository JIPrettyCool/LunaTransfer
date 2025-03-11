package handlers

import (
    "LunaMFT/auth"
    "LunaMFT/config"
    "LunaMFT/utils"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

type CreateUserRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

type CreateUserResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    ApiKey  string `json:"apiKey,omitempty"`
}

func validateUsername(username string) bool {
    match, _ := regexp.MatchString(`^[a-zA-Z0-9_]{3,32}$`, username)
    return match
}

func validatePassword(password string) bool {
    // At least 8 characters, with at least one uppercase, one lowercase, one number
    if len(password) < 8 {
        return false
    }
    hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
    hasNumber := strings.ContainsAny(password, "0123456789")
    return hasUpper && hasLower && hasNumber
}

func validateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if !validateUsername(req.Username) {
        http.Error(w, "Invalid username (3-32 chars, alphanumeric with underscore)", http.StatusBadRequest)
        return
    }

    if !validatePassword(req.Password) {
        http.Error(w, "Password must be at least 8 characters with uppercase, lowercase and number", http.StatusBadRequest)
        return
    }

    if req.Email != "" && !validateEmail(req.Email) {
        http.Error(w, "Invalid email format", http.StatusBadRequest)
        return
    }

    if req.Role == "" || (req.Role != "admin" && req.Role != "user") {
        req.Role = "user"
    }

    if auth.UserExists(req.Username) {
        http.Error(w, "Username already exists", http.StatusConflict)
        return
    }

    user, apiKey, err := auth.CreateUser(req.Username, req.Password, req.Email, req.Role)
    if err != nil {
        if err == auth.ErrUserExists {
            utils.LogSystem("USER_CREATE_FAIL", req.Username, r.RemoteAddr, "User already exists")
            http.Error(w, "User already exists", http.StatusConflict)
        } else {
            utils.LogError("USER_CREATE_ERROR", err, req.Username)
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
        }
        return
    }

    // Log successful user creation
    utils.LogSystem("USER_CREATED", req.Username, r.RemoteAddr, req.Role)

    userDir := filepath.Join(config.StoragePath, user.Username)
    if err := os.MkdirAll(userDir, 0755); err != nil {
        http.Error(w, "Failed to create user directory", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(CreateUserResponse{
        Success: true,
        Message: "User created successfully",
        ApiKey:  apiKey,
    })
}