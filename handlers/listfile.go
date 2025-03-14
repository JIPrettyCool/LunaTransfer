package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/common"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "strings"
    "fmt"
    "LunaTransfer/utils"
)

type FileInfo struct {
    Name      string    `json:"name"`
    Size      int64     `json:"size"`
    IsDir     bool      `json:"isDir"`
    Path      string    `json:"path"`
    Modified  time.Time `json:"modified"`
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    query := r.URL.Query()
    path := query.Get("path")
    
    path = filepath.Clean(path)
    if strings.Contains(path, "..") {
        utils.LogError("LIST_FILES_ERROR", fmt.Errorf("path traversal attempt"), username)
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("LIST_FILES_ERROR", err, username)
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }
    
    userStorageDir := filepath.Join(appConfig.StorageDirectory, username)
    fullPath := filepath.Join(userStorageDir, path)
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        if path == "" || path == "." {
            if err := os.MkdirAll(userStorageDir, 0755); err != nil {
                utils.LogError("LIST_FILES_ERROR", err, username)
                http.Error(w, "Failed to create storage directory", http.StatusInternalServerError)
                return
            }
        } else {
            utils.LogError("LIST_FILES_ERROR", err, username)
            http.Error(w, "Directory not found", http.StatusNotFound)
            return
        }
    }
    
    entries, err := os.ReadDir(fullPath)
    if err != nil {
        utils.LogError("LIST_FILES_ERROR", err, username)
        http.Error(w, "Failed to read directory", http.StatusInternalServerError)
        return
    }
    
    files := make([]FileInfo, 0, len(entries))
    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            continue
        }
        
        entryPath := filepath.Join(path, entry.Name())
        if path == "." || path == "" {
            entryPath = entry.Name()
        }
        
        files = append(files, FileInfo{
            Name:     entry.Name(),
            Size:     info.Size(),
            IsDir:    entry.IsDir(),
            Path:     entryPath,
            Modified: info.ModTime(),
        })
    }
    utils.LogSystem("LIST_FILES", username, r.RemoteAddr, fmt.Sprintf("Listed files in directory: %s", path))
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "path":    path,
        "files":   files,
    })
}