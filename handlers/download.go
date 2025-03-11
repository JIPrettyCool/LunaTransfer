package handlers

import (
    "LunaMFT/auth"
    "LunaMFT/utils"
    "github.com/gorilla/mux"
    "net/http"
    "os"
    "path/filepath"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
    username := r.Header.Get("Username")
    apiKey := r.Header.Get("API-Key")
    
    if username == "" || apiKey == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    user, err := auth.GetUser(username)
    if (err != nil || user.APIKey != apiKey) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    vars := mux.Vars(r)
    filename := vars["filename"]
    if filename == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    sanitizedFilename := filepath.Base(filename)

    filePath := filepath.Join("uploads", sanitizedFilename)
    _, err = os.Stat(filePath)
    if os.IsNotExist(err) {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Disposition", "attachment; filename="+sanitizedFilename)
    w.Header().Set("Content-Type", "application/octet-stream")

    http.ServeFile(w, r, filePath)

    utils.LogMetadata("DOWNLOAD", sanitizedFilename, user.Username, r.RemoteAddr, 0)
}
