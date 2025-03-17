package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"
    
    "LunaTransfer/auth"
    "LunaTransfer/common"
    "LunaTransfer/config"
    "LunaTransfer/models"
    "LunaTransfer/utils"
)

type ShareRequest struct {
    FilePath     string `json:"file_path"`
    SourceGroup  string `json:"source_group"`
    TargetGroup  string `json:"target_group"`
    Permission   string `json:"permission"`
}

type WebSocketMessage struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}

func ShareFileHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req ShareRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        utils.LogError("SHARE_ERROR", err, username, "Invalid request body")
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    utils.LogSystem("SHARE_REQUEST", username, r.RemoteAddr,
        fmt.Sprintf("Attempting to share: %s from %s to %s",
            req.FilePath, req.SourceGroup, req.TargetGroup), time.Now().Unix())

    if req.Permission != "read" && req.Permission != "write" {
        utils.LogSystem("SHARE_ERROR", username, r.RemoteAddr,
            fmt.Sprintf("Invalid permission type: %s", req.Permission), time.Now().Unix())
        http.Error(w, "Invalid permission type. Must be 'read' or 'write'", http.StatusBadRequest)
        return
    }

    hasPermission, err := auth.HasGroupPermission(username, req.SourceGroup, "admin")
    if err != nil || !hasPermission {
        utils.LogSystem("SHARE_ERROR", username, r.RemoteAddr,
            fmt.Sprintf("No permission to share from group %s", req.SourceGroup), time.Now().Unix())
        http.Error(w, "You don't have permission to share files from this group", http.StatusForbidden)
        return
    }

    _, err = auth.GetGroupByID(req.TargetGroup)
    if err != nil {
        utils.LogSystem("SHARE_ERROR", username, r.RemoteAddr,
            fmt.Sprintf("Target group not found: %s", req.TargetGroup), time.Now().Unix())
        http.Error(w, "Target group not found", http.StatusNotFound)
        return
    }

    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("SHARE_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    possiblePaths := []string{
        filepath.Join(appConfig.StorageDirectory, "groups", req.SourceGroup, req.FilePath),
        filepath.Join(appConfig.StorageDirectory, req.FilePath),
    }
    
    var fileExists bool
    var foundFilePath string
    for _, path := range possiblePaths {
        if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
            foundFilePath = path
            fileExists = true
            break
        }
    }

    if !fileExists {
        utils.LogSystem("SHARE_ERROR", username, r.RemoteAddr,
            fmt.Sprintf("File not found: %s (tried paths: %v)", req.FilePath, possiblePaths), time.Now().Unix())
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    utils.LogSystem("SHARE_DEBUG", username, r.RemoteAddr,
        fmt.Sprintf("Found file at: %s", foundFilePath), time.Now().Unix())

    shareID := utils.GenerateUUID()
    share := models.FileShare{
        ID:          shareID,
        FilePath:    req.FilePath,
        SourceGroup: req.SourceGroup,
        TargetGroup: req.TargetGroup,
        Permission:  req.Permission,
        SharedBy:    username,
        SharedAt:    time.Now(),
    }

    err = models.SaveFileShare(share)
    if err != nil {
        utils.LogError("SHARE_ERROR", err, username, "Failed to save share record")
        http.Error(w, "Failed to share file", http.StatusInternalServerError)
        return
    }

    utils.LogSystem("NOTIFICATION", username, r.RemoteAddr,
        fmt.Sprintf("Would notify target group %s about new shared file", req.TargetGroup), 
        time.Now().Unix())

    utils.LogSystem("FILE_SHARED", username, r.RemoteAddr,
        fmt.Sprintf("Shared %s from group %s to group %s with %s permission",
            req.FilePath, req.SourceGroup, req.TargetGroup, req.Permission),
        time.Now().Unix())

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":   "File shared successfully",
        "share_id":  shareID,
    })
}

func ListSharedFilesHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    groupID := r.URL.Query().Get("groupId")
    if groupID == "" {
        http.Error(w, "Group ID required", http.StatusBadRequest)
        return
    }

    hasAccess, err := auth.HasGroupPermission(username, groupID, "read")
    if err != nil || !hasAccess {
        utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr,
            fmt.Sprintf("Attempted to list shared files for group without permission: %s", groupID),
            time.Now().Unix())
        http.Error(w, "Access denied", http.StatusForbidden)
        return
    }

    shares, err := models.GetFileSharesForGroup(groupID)
    if err != nil {
        utils.LogError("LIST_SHARED_ERROR", err, username, 
            fmt.Sprintf("Failed to get shared files for group: %s", groupID))
        http.Error(w, "Failed to retrieve shared files", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "shares": shares,
    })
}

func RemoveShareHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    shareID := r.URL.Path[len("/api/share/"):]
    if shareID == "" {
        http.Error(w, "Share ID required", http.StatusBadRequest)
        return
    }

    share, err := models.GetFileShareByID(shareID)
    if err != nil {
        utils.LogError("REMOVE_SHARE_ERROR", err, username, 
            fmt.Sprintf("Failed to get share: %s", shareID))
        http.Error(w, "Share not found", http.StatusNotFound)
        return
    }

    isAdmin, err := auth.HasGroupPermission(username, share.SourceGroup, "admin")
    if err != nil || (!isAdmin && share.SharedBy != username) {
        utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr,
            fmt.Sprintf("Attempted to remove share without permission: %s", shareID),
            time.Now().Unix())
        http.Error(w, "Access denied", http.StatusForbidden)
        return
    }
    err = models.DeleteFileShare(shareID)
    if err != nil {
        utils.LogError("REMOVE_SHARE_ERROR", err, username, 
            fmt.Sprintf("Failed to delete share: %s", shareID))
        http.Error(w, "Failed to remove share", http.StatusInternalServerError)
        return
    }

    utils.LogSystem("NOTIFICATION", username, r.RemoteAddr,
        fmt.Sprintf("Would notify target group %s about removed share", share.TargetGroup), 
        time.Now().Unix())

    utils.LogSystem("SHARE_REMOVED", username, r.RemoteAddr,
        fmt.Sprintf("Removed share %s for file %s", shareID, share.FilePath),
        time.Now().Unix())

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Share removed successfully",
    })
}