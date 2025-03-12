package main

import (
    "LunaTransfer/auth"
    "LunaTransfer/config"
    "LunaTransfer/handlers"
    "LunaTransfer/middleware"
    "LunaTransfer/utils"
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
    "github.com/gorilla/mux"
)

func main() {
    logger := log.New(os.Stdout, "LunaTransfer: ", log.LstdFlags|log.Lshortfile)
    appConfig, err := config.LoadConfig()
    if err != nil {
        logger.Fatalf("Failed to load configuration: %v", err)
    }
    if err := utils.InitLoggers(); err != nil {
        logger.Fatalf("Failed to initialize loggers: %v", err)
    }
    defer utils.CloseLoggers()
    
    users, err := auth.LoadUsers()
    if err != nil {
        logger.Fatalf("Failed to load users: %v", err)
    }
    logger.Printf("Loaded %d users from database", len(users))

    if err := config.EnsureStorageExists(); err != nil {
        logger.Fatalf("Failed to create storage directory: %v", err)
    }

    r := mux.NewRouter()
    
    r.HandleFunc("/signup", handlers.CreateUserHandler).Methods("POST")
    r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
    
    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.AuthMiddleware)
    api.Use(middleware.RateLimitMiddleware)
    
    api.HandleFunc("/upload", handlers.UploadFile).Methods("POST")
    api.HandleFunc("/download/{filename}", handlers.DownloadFile).Methods("GET")
    api.HandleFunc("/delete/{filename}", handlers.DeleteFile).Methods("DELETE")
    api.HandleFunc("/files", handlers.ListFiles).Methods("GET")
    api.HandleFunc("/dashboard", handlers.DashboardHandler).Methods("GET")
    
    r.Handle("/ws", middleware.AuthMiddleware(http.HandlerFunc(utils.HandleWebSocket))).Methods("GET")

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", appConfig.Port),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
        IdleTimeout:  60 * time.Second,
        Handler:      r,
    }

    go func() {
        logger.Printf("LunaTransfer Server is running on :%d", appConfig.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Server failed: %v", err)
        }
    }()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    
    logger.Println("Shutting down server...")
    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatalf("Server forced to shutdown: %v", err)
    }
    logger.Println("Server gracefully stopped")
}