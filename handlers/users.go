package handlers

import (
	"LunaTransfer/auth"
    "LunaTransfer/utils"
    "encoding/json"
	"time"
    "net/http"
	"github.com/gorilla/mux"
)

func ListUsersHandler(w http.ResponseWriter, r *http.Request) {
    users, err := auth.LoadUsers()
    if err != nil {
        utils.LogError("ADMIN_ERROR", err, "admin", "Failed to load users")
        http.Error(w, "Failed to load users", http.StatusInternalServerError)
        return
    }
    
    type UserResponse struct {
        Username  string    `json:"username"`
        Role      string    `json:"role"`
        CreatedAt time.Time `json:"created_at"`
        Email     string    `json:"email,omitempty"`
    }
    
    response := make([]UserResponse, 0, len(users))
    for _, user := range users {
        response = append(response, UserResponse{
            Username:  user.Username,
            Role:      user.Role,
            CreatedAt: user.CreatedAt,
            Email:     user.Email,
        })
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "users": response,
        "total": len(response),
    })
}
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    username := vars["username"]
    if err := auth.DeleteUser(username); err != nil {
        utils.LogError("ADMIN_ERROR", err, "admin", "Failed to delete user")
        http.Error(w, "Failed to delete user", http.StatusInternalServerError)
        return
    }
    
    utils.LogSystem("USER_DELETED", "admin", r.RemoteAddr, "Deleted user: "+username)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "User deleted successfully",
    })
}