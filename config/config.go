package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

const (
    DefaultPort         = 8080
    DefaultMaxFileSize  = 100 * 1024 * 1024 // 100MB
    DefaultStoragePath  = "./storage"
    DefaultLogFile      = "transfers.log"
    DefaultConfigFile   = "config.json"
    DefaultRateLimit    = 60
    DefaultMaxUploadSize  = 32 << 20
    DefaultTokenExpiry    = 24 * time.Hour
    DefaultMaxConcurrent  = 5
)

var (
    StoragePath string
)

type AppConfig struct {
    Port             int    `json:"port"`
    StorageDirectory string `json:"storage_directory"`
    jsonDBDirectory  string `json:"json_db_directory"`
    MaxFileSize      int64  `json:"max_file_size"` 
    LogDirectory     string `json:"log_directory"`
    JWTSecret        string `json:"jwt_secret"`
    TokenExpiryHours int    `json:"token_expiry_hours"`
    TransferLogFile string `json:"transferLogFile"`
    SystemLogFile   string `json:"systemLogFile"`
    DataPath       string `json:"dataPath"`
    DebugMode      bool   `json:"debugMode"`
    StoragePath    string `json:"storage_path"`
    LogFile        string `json:"log_file"`
    RateLimit      int    `json:"rate_limit"`
    MaxUploadSize  int64
    TokenExpiry    time.Duration
    MaxConcurrent  int
}

var config *AppConfig

func LoadConfig() (*AppConfig, error) {
    if config != nil {
        return config, nil
    }

    config = &AppConfig{
        Port:             8080,
        StorageDirectory: "storage",
        jsonDBDirectory:  "db",
        MaxFileSize:      100 * 1024 * 1024, // 100MB
        LogDirectory:     "logs",
        JWTSecret:        "luna-transfer-secret-key",
        TokenExpiryHours: 24,
        TransferLogFile: "transfers.log",
        SystemLogFile:   "system.log",
        DataPath:       "./data",
        DebugMode:      true,
        StoragePath:    DefaultStoragePath,
        LogFile:        DefaultLogFile,
        RateLimit:      DefaultRateLimit,
        MaxUploadSize:  DefaultMaxUploadSize,
        TokenExpiry:    DefaultTokenExpiry,
        MaxConcurrent:  DefaultMaxConcurrent,
    }

    if _, err := os.Stat(DefaultConfigFile); err == nil {
        file, err := os.ReadFile(DefaultConfigFile)
        if err == nil {
            err = json.Unmarshal(file, config)
            if err != nil {
                return nil, fmt.Errorf("error parsing config file: %w", err)
            }
        }
    }

    if port := os.Getenv("LUNA_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            config.Port = p
        }
    }

    if size := os.Getenv("LUNA_MAX_FILE_SIZE"); size != "" {
        if s, err := strconv.ParseInt(size, 10, 64); err == nil {
            config.MaxFileSize = s
        }
    }

    if path := os.Getenv("LUNA_STORAGE_PATH"); path != "" {
        config.StoragePath = path
    }

    if logFile := os.Getenv("LUNA_LOG_FILE"); logFile != "" {
        config.LogFile = logFile
    }

    if rateLimit := os.Getenv("LUNA_RATE_LIMIT"); rateLimit != "" {
        if r, err := strconv.Atoi(rateLimit); err == nil {
            config.RateLimit = r
        }
    }

    if maxUploadSize := os.Getenv("LunaTransfer_MAX_UPLOAD"); maxUploadSize != "" {
        if s, err := strconv.ParseInt(maxUploadSize, 10, 64); err == nil {
            config.MaxUploadSize = s
        }
    }

    if tokenExpiry := os.Getenv("LunaTransfer_TOKEN_EXPIRY"); tokenExpiry != "" {
        if d, err := time.ParseDuration(tokenExpiry); err == nil {
            config.TokenExpiry = d
        }
    }

    if maxConcurrent := os.Getenv("LunaTransfer_MAX_CONCURRENT"); maxConcurrent != "" {
        if c, err := strconv.Atoi(maxConcurrent); err == nil {
            config.MaxConcurrent = c
        }
    }

    StoragePath = getEnv("STORAGE_DIR", DefaultStoragePath)
    config.StoragePath = StoragePath

    if config.Port <= 0 || config.Port > 65535 {
        return nil, fmt.Errorf("invalid port number: %d", config.Port)
    }

    if config.MaxFileSize <= 0 {
        return nil, fmt.Errorf("max file size must be positive")
    }

    if config.RateLimit <= 0 {
        return nil, fmt.Errorf("rate limit must be positive")
    }

    return config, nil
}

func EnsureStorageExists() error {
    config, err := LoadConfig()
    if err != nil {
        return err
    }

    // Create main storage directory
    if err := os.MkdirAll(config.StorageDirectory, 0755); err != nil {
        return err
    }
    
    // Create db storage directory
    if err := os.MkdirAll(config.jsonDBDirectory, 0755); err != nil {
        return err
    }

    logDir := filepath.Dir(config.LogFile)
    if logDir != "." && logDir != "" {
        if err := os.MkdirAll(logDir, 0755); err != nil {
            return fmt.Errorf("failed to create log directory: %w", err)
        }
    }

    return nil
}

func GetUserStoragePath(username string) string {
    cfg, _ := LoadConfig()
    username = strings.Replace(username, "..", "", -1)
    username = strings.Replace(username, "/", "", -1)
    username = strings.Replace(username, "\\", "", -1)
    return filepath.Join(cfg.StoragePath, username)
}

func EnsureUserStorage(username string) error {
    userPath := GetUserStoragePath(username)
    return os.MkdirAll(userPath, 0755)
}

func SaveConfig() error {
    if config == nil {
        return fmt.Errorf("no configuration to save")
    }

    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("error encoding config: %w", err)
    }

    return os.WriteFile(DefaultConfigFile, data, 0600)
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists && value != "" {
        return value
    }
    return defaultValue
}

func (c *AppConfig) GetDataDirectory() string {
    return c.jsonDBDirectory
}