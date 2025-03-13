package handlers

import (
    "LunaTransfer/auth"
    "LunaTransfer/utils"
    "encoding/json"
    "fmt"
    "net/http"
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
    fmt.Println("CreateUser handler called") // Console debug
    
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.LogError("USER_CREATE_ERROR", err, "Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate required fields
    if req.Username == "" || req.Password == "" {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("missing required fields"), req.Username)
        http.Error(w, "Username and password are required", http.StatusBadRequest)
        return
    }

    // Validate username and password
    if !validateUsername(req.Username) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("invalid username format"), req.Username)
        http.Error(w, "Invalid username format", http.StatusBadRequest)
        return
    }

    if !validatePassword(req.Password) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("password does not meet requirements"), req.Username)
        http.Error(w, "Password must be at least 8 characters with uppercase, lowercase, and number", http.StatusBadRequest)
        return
    }

    // Check if email is valid if provided
    if req.Email != "" && !validateEmail(req.Email) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("invalid email format"), req.Username)
        http.Error(w, "Invalid email format", http.StatusBadRequest)
        return
    }

    // Set default role if not provided
    if req.Role == "" {
        req.Role = "user"
    }

    // Try to create user
    user, apiKey, err := auth.CreateUser(req.Username, req.Password, req.Role, req.Email)
    if err != nil {
        utils.LogError("USER_CREATE_ERROR", err, req.Username)
        http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Log success with explicit information
    utils.LogSystem("USER_CREATED", "system", r.RemoteAddr, 
        fmt.Sprintf("Created user: %s with role: %s", user.Username, user.Role))

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(CreateUserResponse{
        Success: true,
        Message: "User created successfully",
        ApiKey:  apiKey,
    })
}