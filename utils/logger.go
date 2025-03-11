package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	logFile = "transfers.log"
)

var (
	logMutex sync.Mutex
)

func LogMetadata(action, filename, username, ip string, size int64) {
	logMutex.Lock()
	defer logMutex.Unlock()
	os.MkdirAll(filepath.Dir(logFile), 0755)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()
	timestamp := time.Now().Format(time.RFC3339)
	logEntry := fmt.Sprintf("%s - %s - %s - User: %s - IP: %s - %d bytes\n",
		timestamp, action, filename, username, ip, size)
	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}
