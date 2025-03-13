package handlers

import (
    "LunaTransfer/auth"
    "LunaTransfer/config"
    "LunaTransfer/utils"
    "encoding/json"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
    "time"
)
func SystemStatsHandler(w http.ResponseWriter, r *http.Request) {
    stats := map[string]interface{}{}
    users, err := auth.LoadUsers()
    if err == nil {
        stats["total_users"] = len(users)
        
        roleCount := map[string]int{}
        for _, user := range users {
            roleCount[user.Role]++
        }
        stats["users_by_role"] = roleCount
    }
    
    appConfig, err := config.LoadConfig()
    if err == nil {
        var totalSize int64
        var fileCount int
        
        filepath.Walk(appConfig.StorageDirectory, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return nil
            }
            if !info.IsDir() {
                totalSize += info.Size()
                fileCount++
            }
            return nil
        })
        
        stats["total_storage_used"] = totalSize
        stats["total_files"] = fileCount
        stats["storage_used_readable"] = utils.FormatFileSize(totalSize)
    }
    
    stats["go_version"] = runtime.Version()
    stats["os"] = runtime.GOOS
    stats["arch"] = runtime.GOARCH
    stats["cpu_cores"] = runtime.NumCPU()
    stats["goroutines"] = runtime.NumGoroutine()
    stats["uptime"] = time.Since(startTime).String()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

var startTime = time.Now()