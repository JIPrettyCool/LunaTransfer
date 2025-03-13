package utils

import (
    "LunaTransfer/models"
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    // Removed unused "sync" import
    "time"
)

var (
    systemLogger   *log.Logger
    errorLogger    *log.Logger
    accessLogger   *log.Logger
    transferLogger *log.Logger
    logFiles       []*os.File
    transferLogFile string // Added this variable
)

func InitLoggers() error {
    logDir := "logs"
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return err
    }

    currentDate := time.Now().Format("2006-01-02")
    
    // Define the transferLogFile path here
    transferLogFile = filepath.Join(logDir, fmt.Sprintf("transfer_%s.log", currentDate))
    
    systemFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("system_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return err
    }
    logFiles = append(logFiles, systemFile)
    systemLogger = log.New(systemFile, "SYSTEM: ", log.Ldate|log.Ltime)
    
    errorFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("error_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return err
    }
    logFiles = append(logFiles, errorFile)
    errorLogger = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime)
    
    accessFile, err := os.OpenFile(
        filepath.Join(logDir, fmt.Sprintf("access_%s.log", currentDate)),
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return err
    }
    logFiles = append(logFiles, accessFile)
    accessLogger = log.New(accessFile, "ACCESS: ", log.Ldate|log.Ltime)
    
    // Add the transfer logger
    transferFile, err := os.OpenFile(
        transferLogFile,
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0644,
    )
    if err != nil {
        return err
    }
    logFiles = append(logFiles, transferFile)
    transferLogger = log.New(transferFile, "TRANSFER: ", log.Ldate|log.Ltime)
    
    return nil
}

// Add the missing appendToFile function
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
        return
    }
    systemLogger.Printf("[%s] [User: %s] [IP: %s] %s", event, username, ip, fmt.Sprint(message...))
}

func LogError(event string, err error, details ...any) {
    if errorLogger == nil {
        return
    }
    errorLogger.Printf("[%s] Error: %v - Details: %s", event, err, fmt.Sprint(details...))
}

func LogAccess(method, path, username, ip string, statusCode int, duration time.Duration) {
    if accessLogger == nil {
        return
    }
    accessLogger.Printf("[%s %s] [User: %s] [IP: %s] [Status: %d] [Duration: %v]", 
        method, path, username, ip, statusCode, duration)
}

// TransferOperation defines the types of transfer operations
type TransferOperation string

const (
    OpUpload   TransferOperation = "UPLOAD"
    OpDownload TransferOperation = "DOWNLOAD"
    OpDelete   TransferOperation = "DELETE"
)

// TransferLog contains information about file transfers
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

// LogTransfer records file transfer activities (uploads, downloads, deletes)
func LogTransfer(t TransferLog) {
    if transferLogger == nil {
        return
    }
    
    status := "SUCCESS"
    if !t.Success {
        status = "FAILED"
    }
    
    // Format: [ACTION] [STATUS] [USER] [FILE] [SIZE] [IP] [AGENT] [DURATION]
    transferLogger.Printf("[%s] [%s] [User: %s] [File: %s] [Size: %d bytes] [IP: %s] [Agent: %s] [Duration: %v]",
        t.Action, status, t.Username, t.Filename, t.Size, t.RemoteIP, t.UserAgent, t.ElapsedTime.Round(time.Millisecond))
    
    // Also write to the structured file format for backward compatibility
    LogFileTransfer(t.Action, t.Filename, t.Username, t.RemoteIP, t.Size)
}

func GetUserActivity(username string, limit int) ([]models.FileActivity, error) {
    activities := []models.FileActivity{}
    
    // No need to load config since we're using the global transferLogFile
    
    // Use transferLogFile directly
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

// LogFileTransfer writes a structured log entry for file transfers
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
    
    // Use the appendToFile function we defined
    if err := appendToFile(transferLogFile, logEntry); err != nil {
        LogError("LOG_APPEND_ERROR", err, "Failed to append to transfer log")
    }
}

// LogMetadata writes a structured log entry for metadata operations
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
    for _, f := range logFiles {
        f.Close()
    }
}
