package handlers

import (
    "LunaTransfer/common"
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
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.LogError("USER_CREATE_ERROR", err, "unknown", "Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Username == "" || req.Password == "" {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("missing required fields"), "anonymous", req.Username)
        http.Error(w, "Username and password are required", http.StatusBadRequest)
        return
    }
    
    if !validateUsername(req.Username) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("invalid username format"), "anonymous", req.Username)
        http.Error(w, "Username must be 3-32 characters and contain only letters, numbers, and underscores", http.StatusBadRequest)
        return
    }
    
    if !validatePassword(req.Password) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("weak password"), "anonymous", req.Username)
        http.Error(w, "Password must be at least 8 characters with uppercase, lowercase, and numbers", http.StatusBadRequest)
        return
    }
    
    if req.Email != "" && !validateEmail(req.Email) {
        utils.LogError("USER_CREATE_ERROR", fmt.Errorf("invalid email format"), "anonymous", req.Username)
        http.Error(w, "Invalid email format", http.StatusBadRequest)
        return
    }
    userRole, ok := common.GetRoleFromContext(r.Context())
    
    if req.Role == "" {
        req.Role = auth.RoleUser
    }
        if req.Role == auth.RoleAdmin {
        if !ok || !auth.IsAdmin(userRole) {
            utils.LogSystem("ACCESS_DENIED", "unknown", r.RemoteAddr, 
                "Attempted to create user with elevated privileges")
            http.Error(w, "You don't have permission to create users with this role", http.StatusForbidden)
            return
        }
    }
    
    user, apiKey, err := auth.CreateUser(req.Username, req.Password, req.Role, req.Email)
    if err != nil {
        utils.LogError("USER_CREATE_ERROR", err, req.Username)
        http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    utils.LogSystem("USER_CREATED", "system", r.RemoteAddr, 
    fmt.Sprintf("User %s created with role %s", user.Username, user.Role))
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "User created successfully",
        "apiKey":  apiKey,
        "role":    user.Role,
    })
}