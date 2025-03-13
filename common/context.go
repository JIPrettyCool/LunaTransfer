package common

import (
    "context"
)

type contextKey string

const (
    UsernameContextKey contextKey = "username"
    RoleContextKey     contextKey = "role"
)

func GetUsernameFromContext(ctx context.Context) (string, bool) {
    username, ok := ctx.Value(UsernameContextKey).(string)
    return username, ok
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
    role, ok := ctx.Value(RoleContextKey).(string)
    return role, ok
}