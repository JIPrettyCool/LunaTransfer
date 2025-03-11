package main

import (
    "LunaMFT/auth"
    "LunaMFT/config"
    "LunaMFT/handlers"
    "LunaMFT/utils"
    "fmt"
    "github.com/gorilla/mux"
    "log"
    "net/http"
)

func main() {
    _, err := auth.LoadUsers()
    if err != nil {
        log.Fatal("Failed to load users:", err)
    }

    if err := config.EnsureStorageExists(); err != nil {
        log.Fatalf("Failed to create storage directory: %v", err)
    }

    r := mux.NewRouter()
    r.HandleFunc("/signup", handlers.CreateUserHandler).Methods("POST")
    r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
    r.HandleFunc("/upload", utils.RateLimitMiddleware(handlers.UploadFile)).Methods("POST")
    r.HandleFunc("/download/{filename}", utils.RateLimitMiddleware(handlers.DownloadFile)).Methods("GET")
    r.HandleFunc("/delete/{filename}", utils.RateLimitMiddleware(handlers.DeleteFile)).Methods("DELETE")
    r.HandleFunc("/files", handlers.ListFiles).Methods("GET")

    fmt.Println("LunaMFT Server is running on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}