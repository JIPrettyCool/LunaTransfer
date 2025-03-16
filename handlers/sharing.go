package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
	"strings"
    "LunaTransfer/auth"
    "LunaTransfer/common"
    "LunaTransfer/utils"
    
    "github.com/gorilla/mux"
)

func ShareFileHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    var req struct {
        FilePath    string `json:"file_path"`
        SourceGroup string `json:"source_group"`
        TargetGroup string `json:"target_group"`
        Permission  string `json:"permission"`   // "read" or "read_write"
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.FilePath == "" || req.TargetGroup == "" {
        http.Error(w, "File path and target group are required", http.StatusBadRequest)
        return
    }
    
    if req.SourceGroup != "" {
        hasPermission, err := auth.HasGroupPermission(username, req.SourceGroup, "manage")
        if err != nil || !hasPermission {
            utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                fmt.Sprintf("Attempted to share group file without permission: %s", req.FilePath))
            http.Error(w, "You don't have permission to share this file", http.StatusForbidden)
            return
        }
    } else {
        if !strings.HasPrefix(req.FilePath, username+"/") {
            utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                fmt.Sprintf("Attempted to share file they don't own: %s", req.FilePath))
            http.Error(w, "You can only share your own files", http.StatusForbidden)
            return
        }
    }
    
    permission := auth.SharePermissionRead
    if req.Permission == string(auth.SharePermissionReadWrite) {
        permission = auth.SharePermissionReadWrite
    }
    
    share, err := auth.ShareFileWithGroup(req.FilePath, req.SourceGroup, req.TargetGroup, username, permission)
    if err != nil {
        if err == auth.ErrAlreadyShared {
            http.Error(w, "File is already shared with this group", http.StatusConflict)
        } else {
            utils.LogError("SHARE_ERROR", err, username, fmt.Sprintf("Failed to share file: %s", req.FilePath))
            http.Error(w, "Failed to share file", http.StatusInternalServerError)
        }
        return
    }
    
    utils.LogSystem("FILE_SHARED", username, r.RemoteAddr, 
        fmt.Sprintf("Shared file %s with group %s", req.FilePath, req.TargetGroup))
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "File shared successfully",
        "share": share,
    })
}

func RemoveShareHandler(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    vars := mux.Vars(r)
    shareID := vars["shareId"]
    
    if shareID == "" {
        http.Error(w, "Share ID is required", http.StatusBadRequest)
        return
    }
    
    err := auth.RemoveFileSharing(shareID, username)
    if err != nil {
        if err == auth.ErrShareNotFound {
            http.Error(w, "Share not found", http.StatusNotFound)
        } else {
            utils.LogError("SHARE_ERROR", err, username, fmt.Sprintf("Failed to remove share: %s", shareID))
            http.Error(w, "Failed to remove share", http.StatusInternalServerError)
        }
        return
    }
    
    utils.LogSystem("SHARE_REMOVED", username, r.RemoteAddr, 
        fmt.Sprintf("Removed file sharing (ID: %s)", shareID))
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "File sharing removed successfully",
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
        http.Error(w, "Group ID is required", http.StatusBadRequest)
        return
    }
    
    hasAccess, err := auth.HasGroupPermission(username, groupID, "read")
    if err != nil || !hasAccess {
        utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
            fmt.Sprintf("Attempted to list shared files for group without access: %s", groupID))
        http.Error(w, "You don't have access to this group", http.StatusForbidden)
        return
    }
    
    sharedFiles, err := auth.GetFilesSharedWithGroup(groupID)
    if err != nil {
        utils.LogError("SHARE_ERROR", err, username, fmt.Sprintf("Failed to list shared files for group: %s", groupID))
        http.Error(w, "Failed to list shared files", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "shared_files": sharedFiles,
    })
}