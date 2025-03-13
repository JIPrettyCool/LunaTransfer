package middleware

import (
    "LunaTransfer/auth"
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "net/http"
)

func RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role, ok := common.GetRoleFromContext(r.Context())
            if !ok {
                utils.LogSystem("ACCESS_DENIED", "unknown", r.RemoteAddr, 
                    "Role not found in context")
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            var hasAccess bool
            switch requiredRole {
            case auth.RoleAdmin:
                hasAccess = auth.IsAdmin(role)
            case auth.RoleUser:
                hasAccess = auth.IsUser(role)
            default:
                hasAccess = false
            }
            if !hasAccess {
                username, _ := common.GetUsernameFromContext(r.Context())
                utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                    "Insufficient permissions")
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

func PermissionMiddleware(action, resource string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            role, ok := common.GetRoleFromContext(r.Context())
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            if !auth.HasPermission(role, action, resource) {
                username, _ := common.GetUsernameFromContext(r.Context())
                utils.LogSystem("PERMISSION_DENIED", username, r.RemoteAddr, 
                    "No permission for "+action+" on "+resource)
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}