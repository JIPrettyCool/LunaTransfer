package auth

const (
    RoleAdmin     = "admin"
    RoleUser      = "user"
    RoleGuest     = "guest" // Need to implement this..
)
func IsAdmin(role string) bool {
    return role == RoleAdmin
}
func IsUser(role string) bool {
    return role == RoleAdmin || role == RoleUser
}

func HasPermission(role, action, resource string) bool {
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