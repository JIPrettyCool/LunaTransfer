package middleware

import (
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "context"
    "net/http"
    "strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            http.Error(w, "Authorization header must be in format: Bearer {token}", http.StatusUnauthorized)
            return
        }

        tokenString := parts[1]
        claims, err := utils.ValidateJWT(tokenString)
        if err != nil {
            if err == utils.ErrExpiredToken {
                http.Error(w, "Token expired", http.StatusUnauthorized)
            } else {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
            }
            return
        }

        ctx := r.Context()
        ctx = context.WithValue(ctx, common.UsernameContextKey, claims.Username)
        ctx = context.WithValue(ctx, common.RoleContextKey, claims.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}