package handlers

import (
    "LunaTransfer/auth"
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
)

type CreateGroupRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    var req CreateGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.LogError("GROUP_ERROR", err, username, "Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    group, err := auth.CreateGroup(req.Name, req.Description, username)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, username, fmt.Sprintf("Failed to create group: %s", req.Name))
        if err == auth.ErrGroupExists {
            http.Error(w, "Group with that name already exists", http.StatusConflict)
        } else {
            http.Error(w, "Failed to create group", http.StatusInternalServerError)
        }
        return
    }

    err = auth.AddUserToGroup(group.ID, username, "admin", username)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, username, fmt.Sprintf("Failed to add creator to group: %s", req.Name))
    }

    utils.LogSystem("GROUP_CREATED", username, r.RemoteAddr, 
        fmt.Sprintf("Created group: %s", req.Name))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Group created successfully",
        "group": group,
    })
}

func ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    groups, err := auth.LoadGroups()
    if err != nil {
        utils.LogError("GROUP_ERROR", err, username, "Failed to load groups")
        http.Error(w, "Failed to load groups", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "groups": groups,
    })
}

type AddUserToGroupRequest struct {
    Username string `json:"username"`
    Role     string `json:"role"`
}

func AddUserToGroupHandler(w http.ResponseWriter, r *http.Request) {
    adminUsername, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    groupID := vars["groupId"]
    if groupID == "" {
        http.Error(w, "Group ID required", http.StatusBadRequest)
        return
    }

    group, err := auth.GetGroupByID(groupID)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, adminUsername, fmt.Sprintf("Group not found: %s", groupID))
        http.Error(w, "Group not found", http.StatusNotFound)
        return
    }

    var req AddUserToGroupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.LogError("GROUP_ERROR", err, adminUsername, "Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    err = auth.AddUserToGroup(groupID, req.Username, req.Role, adminUsername)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, adminUsername, 
            fmt.Sprintf("Failed to add user %s to group %s", req.Username, group.Name))
        
        switch err {
        case auth.ErrUserAlreadyInGroup:
            http.Error(w, "User is already in group", http.StatusConflict)
        case auth.ErrGroupNotFound:
            http.Error(w, "Group not found", http.StatusNotFound)
        default:
            http.Error(w, "Failed to add user to group", http.StatusInternalServerError)
        }
        return
    }

    utils.LogSystem("GROUP_USER_ADDED", adminUsername, r.RemoteAddr, 
        fmt.Sprintf("Added user %s to group %s with role %s", req.Username, group.Name, req.Role))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("User %s added to group successfully", req.Username),
    })
}

func GetGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    groupID := vars["groupId"]
    if groupID == "" {
        http.Error(w, "Group ID required", http.StatusBadRequest)
        return
    }

    group, err := auth.GetGroupByID(groupID)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, username, fmt.Sprintf("Group not found: %s", groupID))
        http.Error(w, "Group not found", http.StatusNotFound)
        return
    }

    members, err := auth.GetGroupMembers(groupID)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, username, 
            fmt.Sprintf("Failed to get members for group %s", group.Name))
        http.Error(w, "Failed to get group members", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "group": group,
        "members": members,
    })
}

func RemoveUserFromGroupHandler(w http.ResponseWriter, r *http.Request) {
    adminUsername, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    groupID := vars["groupId"]
    username := vars["username"]
    if groupID == "" || username == "" {
        http.Error(w, "Group ID and username required", http.StatusBadRequest)
        return
    }
    group, err := auth.GetGroupByID(groupID)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, adminUsername, fmt.Sprintf("Group not found: %s", groupID))
        http.Error(w, "Group not found", http.StatusNotFound)
        return
    }
    err = auth.RemoveUserFromGroup(groupID, username, adminUsername)
    if err != nil {
        utils.LogError("GROUP_ERROR", err, adminUsername, 
            fmt.Sprintf("Failed to remove user %s from group %s", username, group.Name))
        
        if err == auth.ErrUserNotInGroup {
            http.Error(w, "User is not in this group", http.StatusNotFound)
        } else {
            http.Error(w, "Failed to remove user from group", http.StatusInternalServerError)
        }
        return
    }

    utils.LogSystem("GROUP_USER_REMOVED", adminUsername, r.RemoteAddr, 
        fmt.Sprintf("Removed user %s from group %s", username, group.Name))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": fmt.Sprintf("User %s removed from group successfully", username),
    })
}