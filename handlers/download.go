package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/common"
    "LunaTransfer/utils"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "github.com/gorilla/mux"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {
    username, ok := common.GetUsernameFromContext(r.Context())
    if !ok {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }
    vars := mux.Vars(r)
    filename := vars["filename"]
    if !ok || filename == "" {
        http.Error(w, "Filename required", http.StatusBadRequest)
        return
    }
    sanitizedFilename := filepath.Base(filename)

    filePath := filepath.Join(config.StoragePath, username, sanitizedFilename)
    file, err := os.Open(filePath)
    if (err != nil) {
        if os.IsNotExist(err) {
            http.Error(w, "File not found", http.StatusNotFound)
        } else {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    defer file.Close()
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sanitizedFilename))
    w.Header().Set("Content-Type", "application/octet-stream")
    if _, err := io.Copy(w, file); err != nil {
        http.Error(w, "Failed to send file", http.StatusInternalServerError)
        return
    }
    utils.LogMetadata("DOWNLOAD", sanitizedFilename, username, r.RemoteAddr, time.Now().Unix())
}
