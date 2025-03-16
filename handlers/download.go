package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "github.com/gorilla/mux"
    "strings"
    "mime"
    "LunaTransfer/auth"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    filename := vars["filename"]
    if filename == "" {
        http.Error(w, "Filename required", http.StatusBadRequest)
        return
    }

    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("DOWNLOAD_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    user, err := auth.GetUserByUsername(username)
    if err != nil {
        utils.LogError("DOWNLOAD_ERROR", err, username, "Failed to get user details")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    var filePath string
    isGroupFile := strings.HasPrefix(filename, "groups/")

    if isGroupFile {
        parts := strings.SplitN(filename[7:], "/", 2)
        if len(parts) < 2 {
            utils.LogError("DOWNLOAD_ERROR", fmt.Errorf("invalid group path"), username, filename)
            http.Error(w, "Invalid group file path", http.StatusBadRequest)
            return
        }
        
        groupID := parts[0]
        groupFilePath := parts[1]
        
        _, err := auth.GetGroupByID(groupID)
        if err != nil {
            utils.LogError("DOWNLOAD_ERROR", err, username, fmt.Sprintf("Group not found: %s", groupID))
            http.Error(w, "Group not found", http.StatusNotFound)
            return
        }
        
        hasAccess := user.Role == auth.RoleAdmin
        if !hasAccess {
            hasPermission, err := auth.HasGroupPermission(username, groupID, "read")
            if err != nil {
                utils.LogError("DOWNLOAD_ERROR", err, username, "Failed to check group permissions")
                http.Error(w, "Server error", http.StatusInternalServerError)
                return
            }
            
            hasAccess = hasPermission
        }
        
        if !hasAccess {
            utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                fmt.Sprintf("Attempted to download group file without permission: %s", filename))
            http.Error(w, "Access denied", http.StatusForbidden)
            return
        }
        
        filePath = filepath.Join(appConfig.StorageDirectory, "groups", groupID, groupFilePath)
    } else {
        if !isGroupFile && !strings.HasPrefix(filename, username) {
            hasSharedAccess, err := auth.HasAccessToSharedFile(username, filename, false)
            if err != nil || !hasSharedAccess {
                utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                    fmt.Sprintf("Attempted to download file without permission: %s", filename))
                http.Error(w, "Access denied", http.StatusForbidden)
                return
            }
        }
        filePath = filepath.Join(appConfig.StorageDirectory, username, filename)
    }

    info, err := os.Stat(filePath)
    if os.IsNotExist(err) {
        utils.LogError("DOWNLOAD_ERROR", err, username, fmt.Sprintf("File not found: %s", filename))
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    if info.IsDir() {
        utils.LogError("DOWNLOAD_ERROR", fmt.Errorf("attempted directory download"), username, filename)
        http.Error(w, "Cannot download directories", http.StatusBadRequest)
        return
    }

    file, err := os.Open(filePath)
    if err != nil {
        utils.LogError("DOWNLOAD_ERROR", err, username, fmt.Sprintf("Failed to open file: %s", filename))
        http.Error(w, "Failed to open file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
    
    contentType := mime.TypeByExtension(filepath.Ext(filename))
    if contentType == "" {
        contentType = "application/octet-stream"
    }
    w.Header().Set("Content-Type", contentType)
    
    w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

    if isGroupFile {
        parts := strings.SplitN(filename[7:], "/", 2)
        groupID := parts[0]
        utils.LogSystem("GROUP_FILE_DOWNLOAD", username, r.RemoteAddr, 
            fmt.Sprintf("Downloaded group file: %s from group %s", parts[1], groupID), time.Now().Unix())
    } else {
        utils.LogSystem("FILE_DOWNLOAD", username, r.RemoteAddr, 
            fmt.Sprintf("Downloaded file: %s", filename), time.Now().Unix())
    }

    http.ServeContent(w, r, filepath.Base(filename), info.ModTime(), file)
}
