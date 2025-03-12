package handlers

import (
    "LunaTransfer/config"
    "LunaTransfer/middleware"
    "LunaTransfer/utils"
    "net/http"
    "os"
    "path/filepath"
    "github.com/gorilla/mux"
)

func DeleteFile(w http.ResponseWriter, r *http.Request) {
    username, ok := r.Context().Value(middleware.UsernameContextKey).(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    vars := mux.Vars(r)
    filename, ok := vars["filename"]
    if !ok || filename == "" {
        http.Error(w, "Filename required", http.StatusBadRequest)
        return
    }

    sanitizedFilename := filepath.Base(filename)
    userDir := filepath.Join(config.StoragePath, username)
    filePath := filepath.Join(userDir, sanitizedFilename)
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    if err := os.Remove(filePath); err != nil {
        utils.LogError("DELETE_ERROR", err, username)
        http.Error(w, "Failed to delete file", http.StatusInternalServerError)
        return
    }

    utils.LogFileTransfer("DELETE", sanitizedFilename, username, r.RemoteAddr, 0)
    go utils.NotifyFileDeleted(username, sanitizedFilename)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"success": true, "message": "File deleted successfully"}`))
}
