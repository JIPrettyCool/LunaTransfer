package middleware

import (
    "LunaMFT/auth"
    "context"
    "net/http"
    "strings"
)

type ContextKey string
const (
    UsernameContextKey ContextKey = "username"
    APIKeyContextKey   ContextKey = "api_key"
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
            http.Error(w, "Authorization header must be in format: Bearer {api_key}", http.StatusUnauthorized)
            return
        }

        apiKey := parts[1]
        username, valid := auth.GetUserByAPIKey(apiKey)
        if (!valid) {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }

        ctx := r.Context()
        ctx = context.WithValue(ctx, UsernameContextKey, username)
        ctx = context.WithValue(ctx, APIKeyContextKey, apiKey)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
    username, ok := ctx.Value(UsernameContextKey).(string)
    return username, ok
}

func GetAPIKeyFromContext(ctx context.Context) (string, bool) {
    apiKey, ok := ctx.Value(APIKeyContextKey).(string)
    return apiKey, ok
}