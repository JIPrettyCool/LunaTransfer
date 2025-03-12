package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/models"
    "LunaTransfer/utils"
    "LunaTransfer/middleware"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // Get who's uploading
    username := r.Context().Value(middleware.UsernameContextKey)
    if username == nil {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }
    
    user, ok := username.(string)
    if !ok {
        // This shouldn't happen but just in case
        http.Error(w, "Invalid session", http.StatusInternalServerError)
        return
    }

    // 100MB max - might need to increase this later
    maxSize := 100 * 1024 * 1024
    err := r.ParseMultipartForm(int64(maxSize))
    if err != nil {
        http.Error(w, "File too big! Max is 100MB", http.StatusBadRequest)
        return
    }

    // Get the file
    f, h, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "No file in request", http.StatusBadRequest)
        return
    }
    defer f.Close()
    
    fname := h.Filename
    
    // Quick security check - we don't want any directory traversal attacks xd
    if strings.Contains(fname, "..") || strings.Contains(fname, "/") || strings.Contains(fname, "\\") {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }
    
    // Create user folder if it doesn't exist
    userFolder := filepath.Join(config.StoragePath, user)
    if _, err := os.Stat(userFolder); os.IsNotExist(err) {
        // TODO: set better permissions for production
        os.MkdirAll(userFolder, 0755)
    }
    
    // Where to save the file
    savePath := filepath.Join(userFolder, fname)
    
    if _, err := os.Stat(savePath); err == nil {
        // TODO: add option to prevent overwrites or add version numbers rn it just overwrites
    }
    
    destFile, err := os.Create(savePath)
    if err != nil {
        utils.LogError("FILE_CREATE_ERR", err, user)
        http.Error(w, "Couldn't save your file", http.StatusInternalServerError)
        return
    }
    defer destFile.Close()
    
    written, err := io.Copy(destFile, f)
    if err != nil {
        destFile.Close()
        os.Remove(savePath)
        utils.LogError("WRITE_ERR", err, user)
        http.Error(w, "Upload failed midway", http.StatusInternalServerError)
        return
    }
    
    uploadTime := time.Since(start)
    
    utils.LogTransfer(utils.TransferLog{
        Username:    user,
        Filename:    fname,
        Size:        written,
        Action:      "UPLOAD", // I prefer strings over constants here, easier to read logs
        Timestamp:   time.Now(),
        Success:     true,
        RemoteIP:    r.RemoteAddr,
        UserAgent:   r.UserAgent(),
        ElapsedTime: uploadTime,
    })
    
    // Let user know via WS if they're connected
    // Don't block for this - it's just a nice to have
    go func() {
        notification := models.Notification{
            Type:      models.NoteFileUploaded,
            Filename:  fname,
            Message:   fmt.Sprintf("File uploaded: %s (%d bytes)", fname, written),
            Timestamp: time.Now(),
        }
        utils.NotifyUser(user, notification)
    }()

    resp := fmt.Sprintf(`{"ok":true,"message":"Upload successful","file":"%s","bytes":%d,"took":"%v"}`,
        fname, written, uploadTime.Round(time.Millisecond))
        
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(resp))
}