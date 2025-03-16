package auth

import (
	"LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"

)

var (
    groupsLock   sync.RWMutex
    ErrGroupExists = errors.New("group already exists")
    ErrGroupNotFound = errors.New("group not found")
    ErrUserAlreadyInGroup = errors.New("user already in group")
    ErrUserNotInGroup = errors.New("user not in group")
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

func CreateGroup(name, description, createdBy string) (*Group, error) {
    groupsLock.Lock()
    defer groupsLock.Unlock()

    groups, err := LoadGroups()
    if err != nil {
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

    groupsFile := filepath.Join(appConfig.StorageDirectory, "groups.json")
    
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

    groupsFile := filepath.Join(appConfig.StorageDirectory, "groups.json")
    
    if err := os.MkdirAll(filepath.Dir(groupsFile), 0755); err != nil {
        return fmt.Errorf("failed to create data directory: %w", err)
    }

    if err := os.WriteFile(groupsFile, data, 0644); err != nil {
        return fmt.Errorf("failed to write groups file: %w", err)
    }

    return nil
}

func AddUserToGroup(groupID, username, role, addedBy string) error {
    if role != "member" && role != "admin" {
        role = "member"
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

    membersFile := filepath.Join(appConfig.StorageDirectory, "group_members.json")
    
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

    membersFile := filepath.Join(appConfig.StorageDirectory, "group_members.json")
    
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

    accessFile := filepath.Join(appConfig.StorageDirectory, "fileaccess.json")
    
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

    accessFile := filepath.Join(appConfig.StorageDirectory, "fileaccess.json")
    
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

    membersFile := filepath.Join(appConfig.StorageDirectory, "group_members.json")
    
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