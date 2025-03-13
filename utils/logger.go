package utils

import (
    "LunaTransfer/config"
    "LunaTransfer/models"
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

var (
    systemLogger    *log.Logger
    errorLogger     *log.Logger
    accessLogger    *log.Logger
    transferLogger  *log.Logger
    logFiles        []*os.File
    transferLogFile string
    logDir string
)

func InitLoggers() error {
    appConfig, err := config.LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config for logging: %w", err)
    }
        logDir = appConfig.LogDirectory
    if logDir == "" {
        logDir = "logs" 
    }
    fmt.Printf("[INFO] Creating log directory: %s\n", logDir)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
    }
    currentDate := time.Now().Format("2005-08-08")
    fmt.Printf("[INFO] Initializing loggers for date: %s\n", currentDate)
        transferLogFile = filepath.Join(logDir, fmt.Sprintf("transfer_%s.log", currentDate))
    if err := initSystemLogger(currentDate); err != nil {
        return err
    }
    if err := initErrorLogger(currentDate); err != nil {
        return err
    }
    if err := initAccessLogger(currentDate); err != nil {
        return err
    }
    if err := initTransferLogger(currentDate); err != nil {
        return err
    }
    
    systemLogger.Println("⚙️ System logger initialized successfully")
    errorLogger.Println("⚠️ Error logger initialized successfully")
    accessLogger.Println("🔍 Access logger initialized successfully")
    transferLogger.Println("📦 Transfer logger initialized successfully")
    fmt.Println("[INFO] All loggers initialized successfully")
    return nil
}

func initSystemLogger(currentDate string) error {
    systemFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("system_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return fmt.Errorf("failed to open system log file: %w", err)
    }
    logFiles = append(logFiles, systemFile)
    systemLogger = log.New(systemFile, "SYSTEM: ", log.Ldate|log.Ltime)
    return nil
}

func initErrorLogger(currentDate string) error {
    errorFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("error_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        closeLogFiles()
        return fmt.Errorf("failed to open error log file: %w", err)
    }
    logFiles = append(logFiles, errorFile)
    errorLogger = log.New(errorFile, "ERROR:  ", log.Ldate|log.Ltime)
    return nil
}

func initAccessLogger(currentDate string) error {
    accessFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("access_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        closeLogFiles()
        return fmt.Errorf("failed to open access log file: %w", err)
    }
    logFiles = append(logFiles, accessFile)
    accessLogger = log.New(accessFile, "ACCESS: ", log.Ldate|log.Ltime)
    return nil
}

func initTransferLogger(currentDate string) error {
    transferFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("transfer_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        closeLogFiles()
        return fmt.Errorf("failed to open transfer log file: %w", err)
    }
    logFiles = append(logFiles, transferFile)
    transferLogger = log.New(transferFile, "TRANSFER: ", log.Ldate|log.Ltime)
    return nil
}

func closeLogFiles() {
    for _, f := range logFiles {
        f.Close()
    }
    logFiles = nil
}

func appendToFile(filepath string, content string) error {
    f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    
    if _, err := f.WriteString(content); err != nil {
        return err
    }
    
    return nil
}

func LogSystem(event, username, ip string, message ...any) {
    if systemLogger == nil {
        fmt.Printf("[WARNING] System logger not initialized. Event: %s Message: %s\n", 
            event, fmt.Sprint(message...))
        return
    }
    systemLogger.Printf("📝 [%s] [User: %s] [IP: %s] %s", 
        event, username, ip, fmt.Sprint(message...))
}

func LogError(event string, err error, details ...any) {
    if errorLogger == nil {
        fmt.Printf("[ERROR] Logger not initialized. Event: %s Error: %v Details: %s\n", 
            event, err, fmt.Sprint(details...))
        return
    }
    errorLogger.Printf("❌ [%s] Error: %v - Details: %s", 
        event, err, fmt.Sprint(details...))
}

func LogAccess(method, path, username, ip string, statusCode int, duration time.Duration) {
    if accessLogger == nil {
        fmt.Printf("[WARNING] Access logger not initialized. %s %s [%d] User: %s\n", 
            method, path, statusCode, username)
        return
    }
    
    statusIndicator := "✅"
    if statusCode >= 400 && statusCode < 500 {
        statusIndicator = "⚠️"
    } else if statusCode >= 500 {
        statusIndicator = "🔴"
    } else if statusCode >= 300 && statusCode < 400 {
        statusIndicator = "🔄"
    }
    accessLogger.Printf("%s [%s %s] [User: %s] [IP: %s] [Status: %d] [Duration: %v]", 
        statusIndicator, method, path, username, ip, statusCode, duration.Round(time.Millisecond))
}

type TransferOperation string

const (
    OpUpload   TransferOperation = "UPLOAD"
    OpDownload TransferOperation = "DOWNLOAD"
    OpDelete   TransferOperation = "DELETE"
)
type TransferLog struct {
    Username    string
    Filename    string
    Size        int64
    Action      string
    Timestamp   time.Time
    Success     bool
    RemoteIP    string
    UserAgent   string
    ElapsedTime time.Duration
}

func LogTransfer(t TransferLog) {
    if transferLogger == nil {
        fmt.Printf("[WARNING] Transfer logger not initialized. Action: %s File: %s User: %s\n", 
            t.Action, t.Filename, t.Username)
        return
    }
    status := "SUCCESS"
    indicator := "✅"
    if (!t.Success) {
        status = "FAILED"
        indicator = "❌"
    }
    
    icon := "📤" // Upload
    if t.Action == string(OpDownload) {
        icon = "📥" // Download
    } else if t.Action == string(OpDelete) {
        icon = "🗑️" // Delete
    }
    
    transferLogger.Printf("%s %s [%s] [Status: %s] [User: %s] [File: %s] [Size: %s] [IP: %s] [Duration: %v]",
        icon, indicator, t.Action, status, t.Username, t.Filename, 
        formatFileSize(t.Size), t.RemoteIP, t.ElapsedTime.Round(time.Millisecond))
    
    LogFileTransfer(t.Action, t.Filename, t.Username, t.RemoteIP, t.Size)
}

func formatFileSize(size int64) string {
    const unit = 1024
    if size < unit {
        return fmt.Sprintf("%d B", size)
    }
    div, exp := int64(unit), 0
    for n := size / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func GetUserActivity(username string, limit int) ([]models.FileActivity, error) {
    activities := []models.FileActivity{}
    
    logFile, err := os.Open(transferLogFile)
    if err != nil {
        if os.IsNotExist(err) {
            return activities, nil
        }
        return nil, fmt.Errorf("failed to open log file: %w", err)
    }
    defer logFile.Close()

    scanner := bufio.NewScanner(logFile)
    var lines []string
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    
    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading log file: %w", err)
    }
    
    for i := len(lines) - 1; i >= 0 && len(activities) < limit; i-- {
        line := lines[i]
        
        parts := strings.Split(line, "|")
        if len(parts) < 5 {
            continue
        }
        
        timestamp, _ := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
        operation := strings.TrimSpace(parts[1])
        fileName := strings.TrimSpace(parts[2])
        user := strings.TrimSpace(parts[3])
        remoteIP := strings.TrimSpace(parts[4])
        
        if user != username {
            continue
        }
        
        var fileSize int64 = 0
        if len(parts) > 5 {
            fileSize, _ = strconv.ParseInt(strings.TrimSpace(parts[5]), 10, 64)
        }
        
        activities = append(activities, models.FileActivity{
            Timestamp: timestamp,
            Operation: operation,
            Filename:  fileName,
            Username:  user,
            RemoteIP:  remoteIP,
            Size:      fileSize,
        })
    }

    return activities, nil
}

func LogFileTransfer(operation, filename, username, remoteAddr string, fileSize int64) {
    timestamp := time.Now().Unix()
    logEntry := fmt.Sprintf("%d|%s|%s|%s|%s|%d\n",
        timestamp,
        operation,
        filename,
        username,
        remoteAddr,
        fileSize,
    )
    
    if err := appendToFile(transferLogFile, logEntry); err != nil {
        LogError("LOG_APPEND_ERROR", err, "Failed to append to transfer log")
    }
}

func LogMetadata(operation, filename, username, remoteAddr string, timestamp int64) {
    logEntry := fmt.Sprintf("%d|%s|%s|%s|%s|\n",
        timestamp,
        operation,
        filename,
        username,
        remoteAddr,
    )
    
    if err := appendToFile(transferLogFile, logEntry); err != nil {
        LogError("LOG_APPEND_ERROR", err, "Failed to append to transfer log")
    }
}

func CloseLoggers() {
    LogSystem("SYSTEM_SHUTDOWN", "system", "localhost", "Closing logger files")
    closeLogFiles()
}
