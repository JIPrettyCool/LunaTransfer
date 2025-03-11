package handlers

import (
    "LunaMFT/auth"
    "LunaMFT/config"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

type FileInfo struct {
    FileName     string    `json:"filename"`
    FileSize     int64     `json:"filesize"`
    LastModified time.Time `json:"lastmodified"`
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
    username := r.Header.Get("Username")
    apiKey := r.Header.Get("API-Key")

    if username == "" || apiKey == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    user, err := auth.GetUser(username)
    if err != nil || user.APIKey != apiKey {
        http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
        return
    }

    files, err := getUserFiles(username)
    if err != nil {
        http.Error(w, "Failed to retrieve files: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "success",
        "count":  len(files),
        "files":  files,
    }); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

func getUserFiles(username string) ([]FileInfo, error) {
    // Get user directory using the config package
    userDir := config.UserStoragePath(username)
    
    if _, err := os.Stat(userDir); os.IsNotExist(err) {
        // If directory doesn't exist, create it
        if err := config.EnsureUserStorageExists(username); err != nil {
            return nil, fmt.Errorf("failed to create user directory: %w", err)
        }
        return []FileInfo{}, nil
    }
    
    entries, err := os.ReadDir(userDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory: %w", err)
    }
    
    fileInfos := make([]FileInfo, 0, len(entries))
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        
        info, err := entry.Info()
        if err != nil {
            continue
        }
        
        fileInfos = append(fileInfos, FileInfo{
            FileName:     entry.Name(),
            FileSize:     info.Size(),
            LastModified: info.ModTime(),
        })
    }
    
    return fileInfos, nil
}