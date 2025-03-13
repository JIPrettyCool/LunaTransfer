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
    
    // Get username from context
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        utils.LogError("UPLOAD_ERROR", fmt.Errorf("unauthorized access"), "unknown", r.RemoteAddr)
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }
    
    // Log that we're starting an upload
    utils.LogSystem("UPLOAD_START", username, r.RemoteAddr, "Upload process initiated")
    
    // Parse form with size limit
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to load config")
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }
    
    maxSize := appConfig.MaxFileSize
    err = r.ParseMultipartForm(maxSize)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to parse form")
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }
    
    // Get the file
    f, h, err := r.FormFile("file")
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "No file in request")
        http.Error(w, "No file in request", http.StatusBadRequest)
        return
    }
    defer f.Close()
    
    fname := h.Filename
    
    // Security check
    if strings.Contains(fname, "..") || strings.Contains(fname, "/") || strings.Contains(fname, "\\") {
        utils.LogError("UPLOAD_SECURITY", fmt.Errorf("path traversal attempt"), username, fname)
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }
    
    // Create user directory
    userFolder := filepath.Join(appConfig.StorageDirectory, username)
    if _, err := os.Stat(userFolder); os.IsNotExist(err) {
        if err := os.MkdirAll(userFolder, 0755); err != nil {
            utils.LogError("UPLOAD_ERROR", err, username, "Failed to create user folder")
            http.Error(w, "Storage error", http.StatusInternalServerError)
            return
        }
    }
    savePath := filepath.Join(userFolder, fname)
    // Check for existing file
    if _, err := os.Stat(savePath); err == nil {
        utils.LogSystem("UPLOAD_OVERWRITE", username, r.RemoteAddr, fmt.Sprintf("File %s will be overwritten", fname))
    }
    // Create destination file
    destFile, err := os.Create(savePath)
    if err != nil {
        utils.LogError("UPLOAD_ERROR", err, username, "Failed to create file")
        http.Error(w, "Couldn't save your file", http.StatusInternalServerError)
        return
    }
    defer destFile.Close()
    // Copy file data
    written, err := io.Copy(destFile, f)
    if err != nil {
        destFile.Close()
        os.Remove(savePath)
        utils.LogError("UPLOAD_ERROR", err, username, "Failed during file write")
        http.Error(w, "Upload failed midway", http.StatusInternalServerError)
        return
    }
    // Log success
    uploadTime := time.Since(start)
    utils.LogTransfer(utils.TransferLog{
        Username:    username,
        Filename:    fname,
        Size:        written,
        Action:      string(utils.OpUpload),
        Timestamp:   time.Now(),
        Success:     true,
        RemoteIP:    r.RemoteAddr,
        UserAgent:   r.UserAgent(),
        ElapsedTime: uploadTime,
    })
    // Respond to client
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "filename": fname,
        "size": written,
        "elapsed": uploadTime.String(),
    })
}