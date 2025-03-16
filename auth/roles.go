package auth

import (
    "fmt"
    "net/http"
    "LunaTransfer/common"
    "LunaTransfer/utils"
)

const (
    RoleAdmin     = "admin"
    RoleUser      = "user"
    RoleGuest     = "guest"
)

func IsAdmin(role string) bool {
    return role == RoleAdmin
}

func IsUser(role string) bool {
    return role == RoleAdmin || role == RoleUser
}

func HasPermission(role, action, resource string) bool {
    if role == RoleAdmin {
        return true
    }
    permissions := getRolePermissions(role)
    for _, perm := range permissions {
        if perm.Action == action && (perm.Resource == resource || perm.Resource == "*") {
            return true
        }
    }
    
    return false
}

type Permission struct {
    Action   string
    Resource string
}

func getRolePermissions(role string) []Permission {
    switch role {
    case RoleAdmin:
        return []Permission{
            {Action: "*", Resource: "*"},
        }
    case RoleUser:
        return []Permission{
            {Action: "read", Resource: "files"},
            {Action: "write", Resource: "files"},
            {Action: "delete", Resource: "files"},
        }
    default:
        return []Permission{}
    }
}

func EnsureAdminPermissions(username string) error {
    user, err := GetUserByUsername(username)
    if (err != nil) {
        return err
    }
    
    if user.Role == RoleAdmin {
        return nil
    }
    
    return fmt.Errorf("user does not have admin role")
}

func PermissionMiddleware(action, resource string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            username, ok := common.GetUsernameFromContext(r.Context())
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            user, err := GetUserByUsername(username) 
            if err == nil && user.Role == RoleAdmin {
                next.ServeHTTP(w, r)
                return
            }
            
            if !HasPermission(username, action, resource) {
                utils.LogSystem("PERMISSION_DENIED", username, r.RemoteAddr,
                    fmt.Sprintf("No permission for %s on %s", action, resource))
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}