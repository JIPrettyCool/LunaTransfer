package handlers

import (
    "LunaTransfer/common"
    "LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type DirectoryRequest struct {
    Path string `json:"path"`
    Name string `json:"name"`
}

func CreateDirectoryHandler(w http.ResponseWriter, r *http.Request) {
    var req DirectoryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.LogError("DIRECTORY_ERROR", err, "unknown", "Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        utils.LogError("DIRECTORY_ERROR", fmt.Errorf("username not found in context"), "unknown")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    cleanPath := filepath.Clean(req.Path)
    cleanName := filepath.Clean(req.Name)
    if strings.Contains(cleanPath, "..") || strings.Contains(cleanName, "..") {
        utils.LogError("DIRECTORY_ERROR", fmt.Errorf("path traversal attempt"), username)
        http.Error(w, "Invalid directory path", http.StatusBadRequest)
        return
    }
    
    appConfig, err := config.LoadConfig()
    if err != nil {
        utils.LogError("DIRECTORY_ERROR", err, username)
        http.Error(w, "Server configuration error", http.StatusInternalServerError)
        return
    }
    
    userRootDir := filepath.Join(appConfig.StorageDirectory, username)
    newDirPath := filepath.Join(userRootDir, cleanPath, cleanName)
    
    if err := os.MkdirAll(newDirPath, 0755); err != nil {
        utils.LogError("DIRECTORY_ERROR", err, username)
        http.Error(w, "Failed to create directory", http.StatusInternalServerError)
        return
    }
    
    utils.LogSystem("DIRECTORY_CREATED", username, r.RemoteAddr, fmt.Sprintf("Created directory: %s", filepath.Join(cleanPath, cleanName)))
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Directory created successfully",
        "path":    filepath.Join(cleanPath, cleanName),
    })
}