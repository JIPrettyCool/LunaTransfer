package handlers

import (
    "LunaMFT/auth"
    "LunaMFT/utils"
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "os"
    "path/filepath"
)

func DeleteFile(w http.ResponseWriter, r *http.Request) {
    username := r.Header.Get("Username")
    apiKey   := r.Header.Get("API-Key")
    
    if username == "" || apiKey == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    user, err := auth.GetUser(username)
    if err != nil || user.APIKey != apiKey {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars     := mux.Vars(r)
    filename := vars["filename"]
    if filename == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    
    sanitizedFilename := filepath.Base(filename)

    filePath := filepath.Join("uploads", sanitizedFilename)
    if err := os.Remove(filePath); err != nil {
        if os.IsNotExist(err) {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    utils.LogMetadata("DELETE", sanitizedFilename, user.Username, r.RemoteAddr, 0)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "file"  : sanitizedFilename,
    })
}
