package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/utils"
    "LunaTransfer/common"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    "encoding/json"
)

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