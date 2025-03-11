package utils

import (
    "LunaMFT/config"
    "LunaMFT/models"
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

var (
    transferLogFile *os.File
    systemLogFile   *os.File
    logMutex        sync.Mutex
)

func InitLoggers() error {
    cfg := config.GetConfig()
    
    tFile, err := os.OpenFile(cfg.TransferLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("failed to open transfer log: %w", err)
    }
    transferLogFile = tFile
    
    sFile, err := os.OpenFile(cfg.SystemLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        transferLogFile.Close()
        return fmt.Errorf("failed to open system log: %w", err)
    }
    systemLogFile = sFile
    fmt.Println("✓ Log files successfully initialized:")
    fmt.Printf("  - Transfer log: %s\n", cfg.TransferLogFile)
    fmt.Printf("  - System log: %s\n", cfg.SystemLogFile)
    return nil
}

func LogTransfer(log TransferLog) {
    timestamp := log.Timestamp.Unix()
    elapsedMs := float64(log.ElapsedTime.Milliseconds())
    logEntry := fmt.Sprintf("%d|%s|%s|%s|%s|%d|%t|%s|%.2fms\n",
        timestamp,
        log.Action,
        log.Filename,
        log.Username,
        log.RemoteIP,
        log.Size,
        log.Success,
        log.UserAgent,
        elapsedMs,
    )
    
    appendToFile(transferLogFile, logEntry)
}

func LogSystem(eventType, username, remoteAddr string, details ...string) {
    timestamp := time.Now().Unix()
    detailsStr := ""
    if len(details) > 0 {
        detailsStr = "|" + strings.Join(details, "|")
    }
    
    logEntry := fmt.Sprintf("%d|%s|%s|%s%s\n",
        timestamp,
        eventType,
        username,
        remoteAddr,
        detailsStr,
    )
    
    appendToFile(systemLogFile, logEntry)
    if config.GetConfig().DebugMode {
        fmt.Printf("[SYSTEM] %s - %s from %s %s\n", 
            time.Now().Format("2006-01-02 15:04:05"),
            eventType, 
            username,
            remoteAddr)
    }
}

func LogError(errorType string, err error, username string) {
    timestamp := time.Now().Unix()
    logEntry := fmt.Sprintf("%d|ERROR|%s|%s|%s\n",
        timestamp,
        errorType,
        username,
        err.Error(),
    )
    
    appendToFile(systemLogFile, logEntry)
    fmt.Printf("[ERROR] %s - %s: %s\n",
        time.Now().Format("2005-08-08 15:04:05"),
        errorType,
        err.Error())
}

func appendToFile(file *os.File, entry string) {
    if file == nil {
        fmt.Print(entry)
        return
    }
    
    logMutex.Lock()
    defer logMutex.Unlock()
    
    file.WriteString(entry)
    file.Sync()
}
func CloseLoggers() {
    if transferLogFile != nil {
        transferLogFile.Close()
    }
    if systemLogFile != nil {
        systemLogFile.Close()
    }
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

func GetUserActivity(username string, limit int) ([]models.FileActivity, error) {
    activities := []models.FileActivity{}
    
    logFile, err := os.Open(config.GetConfig().LogFile)
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
    
    appendToFile(transferLogFile, logEntry)
}

func LogMetadata(operation, filename, username, remoteAddr string, timestamp int64) {
    logEntry := fmt.Sprintf("%d|%s|%s|%s|%s|\n",
        timestamp,
        operation,
        filename,
        username,
        remoteAddr,
    )
    
    appendToFile(transferLogFile, logEntry)
}
