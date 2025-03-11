package handlers

import (
    "LunaMFT/config"
    "LunaMFT/middleware"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "sort"
    "time"
)

type FileInfo struct {
    Name         string    `json:"name"`
    Size         int64     `json:"size"`
    LastModified time.Time `json:"lastModified"`
    IsDirectory  bool      `json:"isDirectory"`
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
    username, ok := r.Context().Value(middleware.UsernameContextKey).(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    path := r.URL.Query().Get("path")
    if path != "" && filepath.IsAbs(path) {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }

    userDir := filepath.Join(config.StoragePath, username)
    if path != "" {
        userDir = filepath.Join(userDir, path)
    }

    if err := os.MkdirAll(userDir, 0755); err != nil {
        http.Error(w, "Failed to access directory", http.StatusInternalServerError)
        return
    }

    files, err := os.ReadDir(userDir)
    if err != nil {
        http.Error(w, "Failed to list files", http.StatusInternalServerError)
        return
    }

    fileInfos := make([]FileInfo, 0, len(files))
    for _, file := range files {
        info, err := file.Info()
        if err != nil {
            continue
        }

        fileInfos = append(fileInfos, FileInfo{
            Name:         file.Name(),
            Size:         info.Size(),
            LastModified: info.ModTime(),
            IsDirectory:  file.IsDir(),
        })
    }

    sort.Slice(fileInfos, func(i, j int) bool {
        if fileInfos[i].IsDirectory != fileInfos[j].IsDirectory {
            return fileInfos[i].IsDirectory
        }
        return fileInfos[i].Name < fileInfos[j].Name
    })

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(struct {
        Files []FileInfo `json:"files"`
        Path  string     `json:"path"`
    }{
        Files: fileInfos,
        Path:  path,
    }); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}