package common

import "context"

// ContextKey type for context values
type ContextKey string

const (
    // UsernameContextKey for storing username in context
    UsernameContextKey ContextKey = "username"
    // RoleContextKey for storing role in context
    RoleContextKey ContextKey = "role"
)

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
    username, ok := ctx.Value(UsernameContextKey).(string)
    return username, ok
}

// GetRoleFromContext extracts role from context
func GetRoleFromContext(ctx context.Context) (string, bool) {
    role, ok := ctx.Value(RoleContextKey).(string)
    return role, ok
}