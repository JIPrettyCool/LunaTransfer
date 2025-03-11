package config

import (
    "os"
    "path/filepath"
    "strconv"
    "time"
)

const (
    DefaultStoragePath    = "./storage"
    DefaultMaxUploadSize  = 32 << 20
    DefaultTokenExpiry    = 24 * time.Hour
    DefaultMaxConcurrent  = 5
)
type AppConfig struct {
    StoragePath    string
    MaxUploadSize  int64
    TokenExpiry    time.Duration
    MaxConcurrent  int
}
var Config = loadConfig()
func loadConfig() AppConfig {
    return AppConfig{
        StoragePath:    getEnvOrDefault("LUNAMFT_STORAGE_PATH", DefaultStoragePath),
        MaxUploadSize:  getEnvAsInt64OrDefault("LUNAMFT_MAX_UPLOAD", DefaultMaxUploadSize),
        TokenExpiry:    getEnvAsDurationOrDefault("LUNAMFT_TOKEN_EXPIRY", DefaultTokenExpiry),
        MaxConcurrent:  getEnvAsIntOrDefault("LUNAMFT_MAX_CONCURRENT", DefaultMaxConcurrent),
    }
}

func getEnvOrDefault(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}

func getEnvAsInt64OrDefault(key string, defaultVal int64) int64 {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.ParseInt(val, 10, 64); err == nil {
            return i
        }
    }
    return defaultVal
}

func getEnvAsIntOrDefault(key string, defaultVal int) int {
    if val := os.Getenv(key); val != "" {
        if i, err := strconv.Atoi(val); err == nil {
            return i
        }
    }
    return defaultVal
}

func getEnvAsDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
    if val := os.Getenv(key); val != "" {
        if d, err := time.ParseDuration(val); err == nil {
            return d
        }
    }
    return defaultVal
}

func UserStoragePath(username string) string {
    safeUsername := filepath.Base(username)
    return filepath.Join(Config.StoragePath, safeUsername)
}

func EnsureStorageExists() error {
    return os.MkdirAll(Config.StoragePath, 0755)
}

func EnsureUserStorageExists(username string) error {
    userPath := UserStoragePath(username)
    return os.MkdirAll(userPath, 0755)
}