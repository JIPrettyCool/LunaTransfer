package main

import (
    "LunaTransfer/auth"
    "LunaTransfer/config"
    "LunaTransfer/handlers"
    "LunaTransfer/middleware"
    "LunaTransfer/utils"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path/filepath"
    "time"
    "github.com/gorilla/mux"
)

func main() {
    logger := log.New(os.Stdout, "LunaTransfer: ", log.LstdFlags|log.Lshortfile)
    fmt.Println("LunaTransfer starting up...")
    appConfig, err := config.LoadConfig()
    if (err != nil) {
        logger.Fatalf("Failed to load configuration: %v", err)
    }
        logPath := filepath.Join(appConfig.LogDirectory, "logs") 
    fmt.Printf("Initializing logs in: %s\n", logPath)
    
    if err := utils.InitLoggers(); err != nil {
        logger.Fatalf("Failed to initialize loggers: %v", err)
    }
    defer utils.CloseLoggers()
    utils.LogSystem("SERVER_START", "system", "localhost", 
        fmt.Sprintf("Server starting on port %d", appConfig.Port))
    fmt.Println("Loggers initialized successfully")
        utils.InitJWT(appConfig)
    
    users, err := auth.LoadUsers()
    if err != nil {
        logger.Fatalf("Failed to load users: %v", err)
    }
    logger.Printf("Loaded %d users from database", len(users))

    if err := config.EnsureStorageExists(); err != nil {
        logger.Fatalf("Failed to create storage directory: %v", err)
    }

    r := mux.NewRouter()    

    // Non-authenticated routes with validation
    r.HandleFunc("/setup", handlers.SetupHandler).Methods("POST")
    r.Handle("/signup", middleware.ValidationMiddleware(middleware.ValidateSignupRequest)(http.HandlerFunc(handlers.CreateUserHandler))).Methods("POST")
    r.Handle("/login", middleware.ValidationMiddleware(middleware.ValidateLoginRequest)(http.HandlerFunc(handlers.LoginHandler))).Methods("POST")
    r.Handle("/logout", middleware.AuthMiddleware(http.HandlerFunc(handlers.LogoutHandler))).Methods("POST")

    r.HandleFunc("/api/system/setup-status", setupStatusHandler).Methods("GET")

    r.HandleFunc("/debug/setup", func(w http.ResponseWriter, r *http.Request) {
        usersFile := "users.json"
        fileExists := false
        userCount := 0
        
        data, err := os.ReadFile(usersFile)
        if err == nil {
            fileExists = true
            
            var users []auth.User
            if err := json.Unmarshal(data, &users); err == nil {
                userCount = len(users)
            }
        }
        
        setupStatus, err := auth.IsSetupCompleted()
        errMsg := ""
        if err != nil {
            errMsg = err.Error()
        }
        
        absPath, absErr := filepath.Abs(usersFile)
        if absErr != nil {
            absPath = usersFile
        }
        
        result := map[string]interface{}{
            "fileExists": fileExists,
            "fileSize": len(data),
            "userCount": userCount,
            "setupStatus": setupStatus,
            "error": errMsg,
            "filePath": absPath,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    })

    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.AuthMiddleware)
    api.Use(middleware.RateLimitMiddleware)

    // Add validation to API routes
    maxUploadSize := int64(100 * 1024 * 1024) // 100MB
    api.Handle("/upload", 
        middleware.MaxBodySizeMiddleware(maxUploadSize)(
            middleware.ParamValidationMiddleware(middleware.ValidateUploadRequest)(
                http.HandlerFunc(handlers.UploadFile),
            ),
        ),
    ).Methods("POST")
    api.Handle("/download/{filename:.*}", middleware.ParamValidationMiddleware(middleware.ValidateFilenameParam)(http.HandlerFunc(handlers.DownloadFile))).Methods("GET")
    api.Handle("/delete/{filename}", middleware.ParamValidationMiddleware(middleware.ValidateFilenameParam)(http.HandlerFunc(handlers.DeleteFile))).Methods("DELETE")
    api.Handle("/files", middleware.ParamValidationMiddleware(middleware.ValidateListFilesRequest)(http.HandlerFunc(handlers.ListFiles))).Methods("GET")
    api.Handle("/refresh", http.HandlerFunc(handlers.RefreshTokenHandler)).Methods("POST")
    api.Handle("/dashboard", http.HandlerFunc(handlers.DashboardHandler)).Methods("GET")
    api.Handle("/files", 
        middleware.PermissionMiddleware("read", "files")(
            middleware.ParamValidationMiddleware(middleware.ValidateListFilesRequest)(
                http.HandlerFunc(handlers.ListFiles),
            ),
        ),
    ).Methods("GET")

    api.Handle("/upload", 
        middleware.PermissionMiddleware("write", "files")(
            middleware.MaxBodySizeMiddleware(maxUploadSize)(
                middleware.ParamValidationMiddleware(middleware.ValidateUploadRequest)(
                    http.HandlerFunc(handlers.UploadFile),
                ),
            ),
        ),
    ).Methods("POST")

    api.Handle("/upload/group", 
        middleware.PermissionMiddleware("write", "files")(
            middleware.MaxBodySizeMiddleware(maxUploadSize)(
                middleware.ParamValidationMiddleware(middleware.ValidateGroupUploadRequest)(
                    http.HandlerFunc(handlers.UploadFileWithGroupAccess),
                ),
            ),
        ),
    ).Methods("POST")

    api.Handle("/delete/{filename:.*}", 
        middleware.PermissionMiddleware("delete", "files")(
            middleware.ParamValidationMiddleware(middleware.ValidateFilenameParam)(
                http.HandlerFunc(handlers.DeleteFile),
            ),
        ),
    ).Methods("DELETE")

    api.Handle("/directory", 
        middleware.PermissionMiddleware("write", "files")(
            middleware.ValidationMiddleware(middleware.ValidateDirectoryRequest)(
                http.HandlerFunc(handlers.CreateDirectoryHandler),
            ),
        ),
    ).Methods("POST")

    api.Handle("/search", 
        middleware.PermissionMiddleware("read", "files")(
            middleware.ParamValidationMiddleware(middleware.ValidateSearchRequest)(
                http.HandlerFunc(handlers.SearchFilesHandler),
            ),
        ),
    ).Methods("GET")

    api.Handle("/share", 
        middleware.AuthMiddleware(
            http.HandlerFunc(handlers.ShareFileHandler),
        ),
    ).Methods("POST")

    api.Handle("/share/{shareId}", 
        middleware.AuthMiddleware(
            http.HandlerFunc(handlers.RemoveShareHandler),
        ),
    ).Methods("DELETE")

    api.Handle("/shared", 
        middleware.AuthMiddleware(
            http.HandlerFunc(handlers.ListSharedFilesHandler),
        ),
    ).Methods("GET")

    r.Handle("/ws", middleware.AuthMiddleware(http.HandlerFunc(utils.HandleWebSocket))).Methods("GET")

    admin := api.PathPrefix("/admin").Subrouter()
    admin.Use(middleware.RoleMiddleware(auth.RoleAdmin))
    admin.HandleFunc("/users", handlers.ListUsersHandler).Methods("GET")
    admin.HandleFunc("/users/{username}", handlers.DeleteUserHandler).Methods("DELETE")
    admin.HandleFunc("/system/stats", handlers.SystemStatsHandler).Methods("GET")
    admin.HandleFunc("/groups", handlers.CreateGroupHandler).Methods("POST")
    admin.HandleFunc("/groups", handlers.ListGroupsHandler).Methods("GET")
    admin.HandleFunc("/groups/{groupId}/members", handlers.AddUserToGroupHandler).Methods("POST")
    admin.HandleFunc("/groups/{groupId}/members", handlers.GetGroupMembersHandler).Methods("GET")
    admin.HandleFunc("/groups/{groupId}/members/{username}", handlers.RemoveUserFromGroupHandler).Methods("DELETE")

    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", appConfig.Port),
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
        IdleTimeout:  60 * time.Second,
        Handler:      r,
    }
    logger.Printf("Starting LunaTransfer Server on port %d", appConfig.Port)
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

func setupStatusHandler(w http.ResponseWriter, r *http.Request) {
    userFilePath := "users.json"
    setupCompleted := false
    data, err := os.ReadFile(userFilePath)
    if err == nil {
        var usersList []interface{}
        if json.Unmarshal(data, &usersList) == nil {
            setupCompleted = len(usersList) > 0
        }
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "setupCompleted": setupCompleted,
    })
}