package handlers

import (
    "LunaTransfer/auth"
    "LunaTransfer/common"
    "LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type UploadFileRequest struct {
    Path     string   `json:"path"`
    GroupIDs []string `json:"groupIds"`
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        utils.LogError("UPLOAD_ERROR", err, "unknown", "Failed to parse form")
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        utils.LogError("UPLOAD_ERROR", fmt.Errorf("unauthorized access"), "unknown", r.RemoteAddr)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    utils.LogSystem("UPLOAD_START", username, r.RemoteAddr, "Upload process initiated")
    file, header, err := r.FormFile("file")
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "No file provided")
        http.Error(w, "No file provided", http.StatusBadRequest)
        return
    }
    defer file.Close()
    path := r.FormValue("path")
    if path == "" {
        path = "."
    }
    
    path = filepath.Clean(path)
    if strings.Contains(path, "..") {
        utils.LogError("UPLOAD_ERROR", fmt.Errorf("path traversal attempt"), username)
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }
    
    userStorageDir := filepath.Join(appConfig.StorageDirectory, username)
    targetDir := filepath.Join(userStorageDir, path)
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to create user folder")
        http.Error(w, "Failed to create directory", http.StatusInternalServerError)
        return
    }
    filename := filepath.Clean(header.Filename)
    filePath := filepath.Join(targetDir, filename)
    
    dst, err := os.Create(filePath)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to create file")
        http.Error(w, "Failed to create file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    
    size, err := io.Copy(dst, file)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed during file write")
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }
    
    if r.FormValue("groupIds") != "" {
        var groupIds []string
        if err := json.Unmarshal([]byte(r.FormValue("groupIds")), &groupIds); err == nil {
            fileAccess := auth.FileAccess{
                Path:      filePath,
                Owner:     username,
                IsPublic:  false,
                GroupIDs:  groupIds,
                CreatedAt: time.Now(),
            }
            
            if err := auth.SaveFileAccess(fileAccess); err != nil {
                utils.LogError("UPLOAD_ERROR", err, username, "Failed to save file access info")
            }
        }
    }
    
    uploadTime := time.Since(start)
    utils.LogSystem("UPLOAD_SUCCESS", username, r.RemoteAddr, 
        fmt.Sprintf("Uploaded file %s to %s (size: %d bytes)", filename, path, size))
    
    utils.LogTransfer(utils.TransferLog{
        Username:    username,
        Filename:    filename,
        Size:        size,
        Action:      string(utils.OpUpload),
        Timestamp:   time.Now(),
        Success:     true,
        RemoteIP:    r.RemoteAddr,
        UserAgent:   r.UserAgent(),
        ElapsedTime: uploadTime,
    })
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "File uploaded successfully",
        "filename": filename,
        "path": path,
        "size": size,
        "elapsed": uploadTime.String(),
    })
}

func UploadFileWithGroupAccess(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to parse multipart form")
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    file, handler, err := r.FormFile("file")
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to get file from request")
        http.Error(w, "No file provided", http.StatusBadRequest)
        return
    }
    defer file.Close()
    groupID := r.FormValue("groupId")
    if groupID == "" {
        utils.LogError("UPLOAD_ERROR", fmt.Errorf("missing group ID"), username, "")
        http.Error(w, "Group ID is required", http.StatusBadRequest)
        return
    }
    uploadPath := r.FormValue("path")
    if uploadPath != "" && (strings.Contains(uploadPath, "..") || strings.HasPrefix(uploadPath, "/")) {
        utils.LogError("UPLOAD_ERROR", fmt.Errorf("path traversal attempt"), username, uploadPath)
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    group, err := auth.GetGroupByID(groupID)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, fmt.Sprintf("Group not found: %s", groupID))
        http.Error(w, "Group not found", http.StatusNotFound)
        return
    }
    isAdmin := false
    user, err := auth.GetUserByUsername(username)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to get user details")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    if user.Role == auth.RoleAdmin {
        isAdmin = true
    }
    if !isAdmin {
        members, err := auth.GetGroupMembers(groupID)
        if err != nil {
            utils.LogError("UPLOAD_ERROR", err, username, "Failed to get group members")
            http.Error(w, "Server error", http.StatusInternalServerError)
            return
        }
        isMember := false
        for _, member := range members {
            if member.Username == username {
                isMember = true
                break
            }
        }
        if !isMember {
            utils.LogSystem("ACCESS_DENIED", username, r.RemoteAddr, 
                fmt.Sprintf("Attempted to upload to group without membership: %s", group.Name))
            http.Error(w, "Access denied - you are not a member of this group", http.StatusForbidden)
            return
        }
    }
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }
    groupDir := filepath.Join(appConfig.StorageDirectory, "groups", groupID)
    targetDir := groupDir
    if uploadPath != "" {
        targetDir = filepath.Join(groupDir, uploadPath)
        if err := os.MkdirAll(targetDir, 0755); err != nil {
            utils.LogError("UPLOAD_ERROR", err, username, fmt.Sprintf("Failed to create directory: %s", uploadPath))
            http.Error(w, "Failed to create directory", http.StatusInternalServerError)
            return
        }
    }
    filePath := filepath.Join(targetDir, filepath.Base(handler.Filename))
    dst, err := os.Create(filePath)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, fmt.Sprintf("Failed to create file: %s", handler.Filename))
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()
    if _, err := io.Copy(dst, file); err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, fmt.Sprintf("Failed to write file: %s", handler.Filename))
        http.Error(w, "Failed to save file", http.StatusInternalServerError)
        return
    }
    relFilePath := filepath.Join("groups", groupID)
    if uploadPath != "" {
        relFilePath = filepath.Join(relFilePath, uploadPath)
    }
    relFilePath = filepath.Join(relFilePath, handler.Filename)

    utils.LogSystem("GROUP_FILE_UPLOAD", username, r.RemoteAddr, 
        fmt.Sprintf("Uploaded file %s to group %s", handler.Filename, group.Name))

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "File uploaded successfully",
        "file": map[string]interface{}{
            "name": handler.Filename,
            "path": relFilePath,
            "size": handler.Size,
            "type": handler.Header.Get("Content-Type"),
            "group": group.Name,
        },
    })
}