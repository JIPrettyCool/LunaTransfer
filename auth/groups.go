package auth

import (
	"LunaTransfer/common"
	"LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"
    "github.com/gorilla/mux"
)

var (
    groupsLock   sync.RWMutex
    ErrGroupExists = errors.New("group already exists")
    ErrGroupNotFound = errors.New("group not found")
    ErrUserAlreadyInGroup = errors.New("user already in group")
    ErrUserNotInGroup = errors.New("user not in group")
    sharingLock sync.RWMutex
    ErrAlreadyShared = errors.New("file is already shared with this group")
    ErrShareNotFound = errors.New("share not found")
)

// Group role constants
const (
    GroupRoleAdmin       = "admin"       // Can manage group and all files
    GroupRoleContributor = "contributor" // Can upload and modify files
    GroupRoleReader      = "reader"      // Can only view and download files
)

type Group struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedBy   string    `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
}

type GroupMember struct {
    GroupID  string    `json:"group_id"`
    Username string    `json:"username"`
    Role     string    `json:"role"`
    AddedBy  string    `json:"added_by"`
    AddedAt  time.Time `json:"added_at"`
}

type GroupPermission struct {
    GroupID     string    `json:"group_id"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Path        string    `json:"path"`
    GrantedBy   string    `json:"granted_by"`
    GrantedAt   time.Time `json:"granted_at"`
}

type FileAccess struct {
    Path        string   `json:"path"`
    Owner       string   `json:"owner"`
    IsPublic    bool     `json:"is_public"`
    GroupIDs    []string `json:"group_ids"`
    CreatedAt   time.Time `json:"created_at"`
}

type SharePermission string

const (
    SharePermissionRead     SharePermission = "read"
    SharePermissionReadWrite SharePermission = "read_write"
)

type SharedFile struct {
    ID          string         `json:"id"`
    SourcePath  string         `json:"source_path"`
    GroupID     string         `json:"group_id"`     // Group the file is shared with
    SourceGroup string         `json:"source_group"` // Original group
    SharedBy    string         `json:"shared_by"`
    SharedAt    time.Time      `json:"shared_at"`
    Permission  SharePermission `json:"permission"`  // Read or read/write
}

func CreateGroup(name, description, createdBy string) (*Group, error) {
    groupsLock.Lock()
    defer groupsLock.Unlock()

    groups, err := LoadGroups()
    if (err != nil) {
        return nil, err
    }

    for _, group := range groups {
        if group.Name == name {
            return nil, ErrGroupExists
        }
    }

    newGroup := &Group{
        ID:          utils.GenerateUUID(),
        Name:        name,
        Description: description,
        CreatedBy:   createdBy,
        CreatedAt:   time.Now(),
    }

    groups = append(groups, *newGroup)

    if err := saveGroups(groups); err != nil {
        return nil, err
    }

    appConfig, err := config.LoadConfig()
    if (err != nil) {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    groupDir := filepath.Join(appConfig.StorageDirectory, "groups", newGroup.ID)
    if err := os.MkdirAll(groupDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create group directory: %w", err)
    }

    return newGroup, nil
}

func LoadGroups() ([]Group, error) {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    groupsFile := filepath.Join(appConfig.GetDataDirectory(), "groups.json")
    
    if _, err := os.Stat(groupsFile); os.IsNotExist(err) {
        return []Group{}, nil
    }

    data, err := os.ReadFile(groupsFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read groups file: %w", err)
    }

    var groups []Group
    if len(data) > 0 {
        if err := json.Unmarshal(data, &groups); err != nil {
            return nil, fmt.Errorf("failed to parse groups data: %w", err)
        }
    }

    return groups, nil
}

func saveGroups(groups []Group) error {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    data, err := json.MarshalIndent(groups, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal groups data: %w", err)
    }

    groupsFile := filepath.Join(appConfig.GetDataDirectory(), "groups.json")
    
    if err := os.MkdirAll(filepath.Dir(groupsFile), 0755); err != nil {
        return fmt.Errorf("failed to create data directory: %w", err)
    }

    if err := os.WriteFile(groupsFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write groups file: %w", err)
    }

    return nil
}

func AddUserToGroup(groupID, username, role, addedBy string) error {
    if role != GroupRoleAdmin && role != GroupRoleContributor && role != GroupRoleReader {
        role = GroupRoleReader
    }

    _, err := GetUserByUsername(username)
    if err != nil {
        return err
    }

    _, err = GetGroupByID(groupID)
    if err != nil {
        return err
    }

    members, err := GetGroupMembers(groupID)
    if err != nil {
        return err
    }

    for _, member := range members {
        if member.Username == username {
            return ErrUserAlreadyInGroup
        }
    }

    newMember := GroupMember{
        GroupID:  groupID,
        Username: username,
        Role:     role,
        AddedBy:  addedBy,
        AddedAt:  time.Now(),
    }

    return saveGroupMember(newMember)
}

func GetGroupByID(id string) (*Group, error) {
    groups, err := LoadGroups()
    if err != nil {
        return nil, err
    }

    for _, group := range groups {
        if group.ID == id {
            return &group, nil
        }
    }

    return nil, ErrGroupNotFound
}

func GetGroupMembers(groupID string) ([]GroupMember, error) {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    membersFile := filepath.Join(appConfig.GetDataDirectory(), "group_members.json")
    
    if _, err := os.Stat(membersFile); os.IsNotExist(err) {
        return []GroupMember{}, nil
    }

    data, err := os.ReadFile(membersFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read members file: %w", err)
    }

    var allMembers []GroupMember
    if len(data) > 0 {
        if err := json.Unmarshal(data, &allMembers); err != nil {
            return nil, fmt.Errorf("failed to parse members data: %w", err)
        }
    }

    var groupMembers []GroupMember
    for _, member := range allMembers {
        if member.GroupID == groupID {
            groupMembers = append(groupMembers, member)
        }
    }

    return groupMembers, nil
}

func saveGroupMember(member GroupMember) error {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    membersFile := filepath.Join(appConfig.GetDataDirectory(), "group_members.json")
    
    var members []GroupMember
    if _, err := os.Stat(membersFile); !os.IsNotExist(err) {
        data, err := os.ReadFile(membersFile)
        if err != nil {
            return fmt.Errorf("failed to read members file: %w", err)
        }

        if len(data) > 0 {
            if err := json.Unmarshal(data, &members); err != nil {
                return fmt.Errorf("failed to parse members data: %w", err)
            }
        }
    }

    members = append(members, member)

    data, err := json.MarshalIndent(members, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal members data: %w", err)
    }

    if err := os.MkdirAll(filepath.Dir(membersFile), 0755); err != nil {
        return fmt.Errorf("failed to create data directory: %w", err)
    }

    if err := os.WriteFile(membersFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write members file: %w", err)
    }

    return nil
}

func HasFileAccess(username, filePath string) (bool, error) {
    user, err := GetUserByUsername(username)
    if err != nil {
        return false, err
    }
    if user.Role == RoleAdmin {
        return true, nil
    }

    fileAccess, err := GetFileAccess(filePath)
    if err != nil {
        return false, nil
    }

    if fileAccess.Owner == username {
        return true, nil
    }

    if fileAccess.IsPublic {
        return true, nil
    }

    for _, groupID := range fileAccess.GroupIDs {
        members, err := GetGroupMembers(groupID)
        if err != nil {
            continue
        }

        for _, member := range members {
            if member.Username == username {
                return true, nil
            }
        }
    }

    return false, nil
}

func GetFileAccess(filePath string) (*FileAccess, error) {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    accessFile := filepath.Join(appConfig.GetDataDirectory(), "fileaccess.json")
    
    if _, err := os.Stat(accessFile); os.IsNotExist(err) {
        return nil, fmt.Errorf("no access control defined for path")
    }

    data, err := os.ReadFile(accessFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read file access data: %w", err)
    }

    var accessList []FileAccess
    if err := json.Unmarshal(data, &accessList); err != nil {
        return nil, fmt.Errorf("failed to parse file access data: %w", err)
    }

    for _, access := range accessList {
        if access.Path == filePath {
            return &access, nil
        }
    }

    return nil, fmt.Errorf("no access control defined for path")
}

func SaveFileAccess(access FileAccess) error {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    accessFile := filepath.Join(appConfig.GetDataDirectory(), "fileaccess.json")
    
    var accessList []FileAccess
    
    if _, err := os.Stat(accessFile); !os.IsNotExist(err) {
        data, err := os.ReadFile(accessFile)
        if err != nil {
            return fmt.Errorf("failed to read file access data: %w", err)
        }

        if len(data) > 0 {
            if err := json.Unmarshal(data, &accessList); err != nil {
                return fmt.Errorf("failed to parse file access data: %w", err)
            }
        }
    }

    found := false
    for i, a := range accessList {
        if a.Path == access.Path {
            accessList[i] = access
            found = true
            break
        }
    }

    if !found {
        accessList = append(accessList, access)
    }

    data, err := json.MarshalIndent(accessList, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal file access data: %w", err)
    }

    if err := os.WriteFile(accessFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write file access data: %w", err)
    }

    return nil
}

func RemoveUserFromGroup(groupID, username, removedBy string) error {
    group, err := GetGroupByID(groupID)
    if err != nil {
        return err
    }

    appConfig, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    membersFile := filepath.Join(appConfig.GetDataDirectory(), "group_members.json")
    
    if _, err := os.Stat(membersFile); os.IsNotExist(err) {
        return ErrUserNotInGroup
    }

    data, err := os.ReadFile(membersFile)
    if err != nil {
        return fmt.Errorf("failed to read members file: %w", err)
    }

    var allMembers []GroupMember
    if err := json.Unmarshal(data, &allMembers); err != nil {
        return fmt.Errorf("failed to parse members data: %w", err)
    }

    found := false
    var updatedMembers []GroupMember
    for _, member := range allMembers {
        if member.GroupID == groupID && member.Username == username {
            found = true
        } else {
            updatedMembers = append(updatedMembers, member)
        }
    }

    if !found {
        return ErrUserNotInGroup
    }
    data, err = json.MarshalIndent(updatedMembers, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal members data: %w", err)
    }
    if err := os.WriteFile(membersFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write members file: %w", err)
    }

    utils.LogSystem("GROUP_USER_REMOVED", removedBy, "", 
        fmt.Sprintf("User %s was removed from group %s by %s", username, group.Name, removedBy))

    return nil
}

func HasGroupPermission(username, groupID, action string) (bool, error) {
    user, err := GetUserByUsername(username)
    if err == nil && user.Role == RoleAdmin {
        return true, nil
    }
    
    members, err := GetGroupMembers(groupID)
    if err != nil {
        return false, err
    }
    
    var userRole string
    isMember := false
    for _, member := range members {
        if member.Username == username {
            userRole = member.Role
            isMember = true
            break
        }
    }
    
    if !isMember {
        return false, nil
    }
    
    switch action {
    case "read":
        return true, nil
    case "write", "upload":
        return userRole == GroupRoleAdmin || userRole == GroupRoleContributor, nil
    case "manage", "delete":
        return userRole == GroupRoleAdmin, nil
    }
    return false, nil
}

func AddUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
    adminUsername, ok := common.GetUsernameFromContext(r.Context())
    if (!ok) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    vars := mux.Vars(r)
    groupID := vars["groupId"]
    
    var req struct {
        Username string `json:"username"`
        Role     string `json:"role"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.Role != GroupRoleAdmin && req.Role != GroupRoleContributor && req.Role != GroupRoleReader {
        http.Error(w, "Invalid role. Must be 'admin', 'contributor', or 'reader'", http.StatusBadRequest)
        return
    }
    
    _, err := GetUserByUsername(req.Username)
    if err != nil {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    
    err = AddUserToGroup(groupID, req.Username, req.Role, adminUsername)
    if err != nil {
        if err == ErrGroupNotFound {
            http.Error(w, "Group not found", http.StatusNotFound)
        } else if err == ErrUserAlreadyInGroup {
            http.Error(w, "User already in group", http.StatusConflict)
        } else {
            http.Error(w, "Failed to add user to group", http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("User %s added to group with role %s", req.Username, req.Role),
    })
}

func LoadSharedFiles() ([]SharedFile, error) {
    sharingLock.RLock()
    defer sharingLock.RUnlock()
    
    appConfig, err := config.LoadConfig()
    if err != nil {
        return nil, err
    }
    
    sharesFile := filepath.Join(appConfig.GetDataDirectory(), "shared_files.json")
    
    if _, err := os.Stat(sharesFile); os.IsNotExist(err) {
        return []SharedFile{}, nil
    }
    
    data, err := os.ReadFile(sharesFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read shared files: %w", err)
    }
    
    var shares []SharedFile
    if err := json.Unmarshal(data, &shares); err != nil {
        return nil, fmt.Errorf("failed to parse shared files data: %w", err)
    }
    
    return shares, nil
}

func saveSharedFiles(shares []SharedFile) error {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return err
    }
    
    sharesFile := filepath.Join(appConfig.GetDataDirectory(), "shared_files.json")
    
    data, err := json.MarshalIndent(shares, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal shared files data: %w", err)
    }
    
    return os.WriteFile(sharesFile, data, 0644)
}

func ShareFileWithGroup(sourcePath, sourceGroup, targetGroupID, sharedBy string, permission SharePermission) (*SharedFile, error) {
    sharingLock.Lock()
    defer sharingLock.Unlock()
    
    if permission != SharePermissionRead && permission != SharePermissionReadWrite {
        permission = SharePermissionRead
    }
    
    if sourceGroup != "" {
        if _, err := GetGroupByID(sourceGroup); err != nil {
            return nil, fmt.Errorf("source group not found: %w", err)
        }
    }
    
    if _, err := GetGroupByID(targetGroupID); err != nil {
        return nil, fmt.Errorf("target group not found: %w", err)
    }
    
    if sourceGroup == targetGroupID {
        return nil, errors.New("cannot share within the same group")
    }
    
    shares, err := LoadSharedFiles()
    if err != nil {
        return nil, err
    }
    
    for _, s := range shares {
        if s.SourcePath == sourcePath && s.GroupID == targetGroupID {
            return nil, ErrAlreadyShared
        }
    }
    
    newShare := &SharedFile{
        ID:          utils.GenerateUUID(),
        SourcePath:  sourcePath,
        GroupID:     targetGroupID,
        SourceGroup: sourceGroup,
        SharedBy:    sharedBy,
        SharedAt:    time.Now(),
        Permission:  permission,
    }
    
    shares = append(shares, *newShare)
    if err := saveSharedFiles(shares); err != nil {
        return nil, err
    }
    
    return newShare, nil
}

func RemoveFileSharing(shareID, username string) error {
    sharingLock.Lock()
    defer sharingLock.Unlock()
    
    shares, err := LoadSharedFiles()
    if err != nil {
        return err
    }
    
    found := false
    var updatedShares []SharedFile
    
    for _, share := range shares {
        if share.ID == shareID {
            found = true
            
            if !IsUserAdmin(username) && share.SharedBy != username {
                return errors.New("you can only remove your own shared files")
            }
        } else {
            updatedShares = append(updatedShares, share)
        }
    }
    
    if !found {
        return ErrShareNotFound
    }
    
    return saveSharedFiles(updatedShares)
}

func GetFilesSharedWithGroup(groupID string) ([]SharedFile, error) {
    allShares, err := LoadSharedFiles()
    if err != nil {
        return nil, err
    }
    
    var groupShares []SharedFile
    for _, share := range allShares {
        if share.GroupID == groupID {
            groupShares = append(groupShares, share)
        }
    }
    
    return groupShares, nil
}

func HasAccessToSharedFile(username, filePath string, writeAccess bool) (bool, error) {
    user, err := GetUserByUsername(username)
    if err != nil {
        return false, err
    }
    
    if user.Role == RoleAdmin {
        return true, nil
    }
    
    userGroups, err := GetUserGroups(username)
    if err != nil {
        return false, err
    }
    
    shares, err := LoadSharedFiles()
    if err != nil {
        return false, err
    }
    
    for _, share := range shares {
        if share.SourcePath == filePath {
            for _, group := range userGroups {
                if group.ID == share.GroupID {
                    if writeAccess {
                        return share.Permission == SharePermissionReadWrite, nil
                    }
                    return true, nil
                }
            }
        }
    }
    
    return false, nil
}

func GetUserGroups(username string) ([]Group, error) {
    allGroups, err := LoadGroups()
    if err != nil {
        return nil, err
    }
    
    var userGroups []Group
    
    for _, group := range allGroups {
        members, err := GetGroupMembers(group.ID)
        if err != nil {
            continue
        }
        
        for _, member := range members {
            if member.Username == username {
                userGroups = append(userGroups, group)
                break
            }
        }
    }
    
    return userGroups, nil
}

func IsUserAdmin(username string) bool {
    user, err := GetUserByUsername(username)
    if err != nil {
        return false
    }
    return user.Role == RoleAdmin
}