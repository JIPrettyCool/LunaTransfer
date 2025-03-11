package handlers

import (
    "LunaMFT/auth"
    "LunaMFT/utils"
    "encoding/json"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
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
    err = r.ParseMultipartForm(32 << 20)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    defer file.Close()

    err = os.MkdirAll("uploads", 0755)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    filename := sanitizeFileName(header.Filename)
    filePath := filepath.Join("uploads", filename)

    dst, err := os.Create(filePath)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    size, err := io.Copy(dst, file)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    utils.LogMetadata("UPLOAD", filename, username, r.RemoteAddr, size)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":   "success",
        "filename": filename,
        "size":     size,
    })
}

func sanitizeFileName(filename string) string {
    filename = filepath.Base(filename)
    filename = strings.ReplaceAll(filename, " ", "_")    
    return filename
}